package genderize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	httpClient "person-info/internal/client"
	personClient "person-info/internal/client/person"
)

const (
	genderizeBaseURL = "https://api.genderize.io"
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
	Name        string  `json:"name"`
	Count       int     `json:"count"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
}

func (c *Client) Gender(ctx context.Context, name string) (string, error) {
	const op = "client.person.genderize.Gender"

	log := c.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)

	log.Info("predicting person gender")

	ctx, cancel := context.WithTimeout(ctx, httpClient.APIRequestTimeout)
	defer cancel()

	var result Response
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("name", name).
		SetResult(&result).
		Get(genderizeBaseURL)
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("%s: %w", op, ctx.Err())
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("api.genderize response status", slog.String("status", resp.Status()))

	if result.Gender == "" {
		return "", fmt.Errorf("%s: %w", op, personClient.ErrInvalidName)
	}

	return result.Gender, nil
}
