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
	logger     *slog.Logger
	httpClient *http.Client
}

func NewClient(logger *slog.Logger) Client {
	return &client{
		logger:     logger,
		httpClient: &http.Client{},
	}
}

func (c *client) GetLimitedEditionDisplates() ([]Displate, error) {
	request, err := http.NewRequest(http.MethodGet, "https://sapi.displate.com/artworks/limited", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("user-agent", "https://t.me/displatebot")

	response, err := c.httpClient.Do(request)
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
