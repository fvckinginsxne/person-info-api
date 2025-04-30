package nationalize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	httpClient "person-info/internal/client"
	personClient "person-info/internal/client/person"
)

const (
	nationalizeBaseURL = "https://api.nationalize.io"
)

type Client struct {
	log    *slog.Logger
	client *resty.Client
}

func New(log *slog.Logger) *Client {
	return &Client{
		log:    log,
		client: resty.New(),
	}
}

type Response struct {
	Name    string `json:"name"`
	Count   int    `json:"count"`
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

func (c *Client) Nationality(ctx context.Context, name string) (string, error) {
	const op = "client.person.nationalize.Nationality"

	log := c.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)

	log.Info("predicting person nationality")

	ctx, cancel := context.WithTimeout(ctx, httpClient.APIRequestTimeout)
	defer cancel()

	var result Response
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("name", name).
		SetResult(&result).
		Get(nationalizeBaseURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("api.nationalize response status", slog.String("status", resp.Status()))

	if len(result.Country) == 0 {
		return "", fmt.Errorf("%s: %s", op, personClient.ErrInvalidName)
	}

	var nationality string

	maxProbability := 0.0
	for _, country := range result.Country {
		if country.Probability > maxProbability {
			maxProbability = country.Probability
			nationality = country.CountryID
		}
	}

	return nationality, nil
}
