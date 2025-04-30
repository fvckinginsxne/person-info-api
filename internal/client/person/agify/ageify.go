package agify

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	httpClient "person-info/internal/client"
	personClient "person-info/internal/client/person"
)

const (
	agifyBaseURL = "https://api.agify.io"
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
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Count int    `json:"count"`
}

func (c *Client) Age(ctx context.Context, name string) (int, error) {
	const op = "client.person.agify.Age"

	log := c.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)

	log.Info("predicting person age")

	ctx, cancel := context.WithTimeout(ctx, httpClient.APIRequestTimeout)
	defer cancel()

	var result Response
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("name", name).
		SetResult(&result).
		Get(agifyBaseURL)
	if err != nil {
		if ctx.Err() != nil {
			return 0, fmt.Errorf("%s: %w", op, ctx.Err())
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("api.agify response status", slog.String("status", resp.Status()))

	if result.Age == 0 {
		return 0, fmt.Errorf("%s: %w", op, personClient.ErrInvalidName)
	}

	return result.Age, nil
}
