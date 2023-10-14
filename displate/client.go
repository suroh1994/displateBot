package displate

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type Client interface {
	GetLimitedEditionDisplates() ([]Displate, error)
}

type client struct {
	logger *slog.Logger
}

func NewClient(logger *slog.Logger) Client {
	return &client{
		logger: logger,
	}
}

func (c *client) GetLimitedEditionDisplates() ([]Displate, error) {
	response, err := http.Get("https://sapi.displate.com/artworks/limited?miso=DE")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var limitedEditionResponse LimitedEditionResponse
	err = json.Unmarshal(bodyBytes, &limitedEditionResponse)
	if err != nil {
		return nil, err
	}

	return limitedEditionResponse.Data, nil
}
