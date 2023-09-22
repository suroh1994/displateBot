package displateApi

import (
	"encoding/json"
	"io"
	"net/http"
)

type LimitedEditionResponse struct {
	Data []Displate `json:"data"`
}
type Image struct {
	URL string `json:"url"`
	Alt any    `json:"alt"`
}
type Images struct {
	Main Image `json:"main"`
}
type Edition struct {
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Status      string `json:"status"`
	Available   int    `json:"available"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	TimeToStart int    `json:"timeToStart"`
}
type Displate struct {
	ID               int     `json:"id"`
	ItemCollectionID int     `json:"itemCollectionId"`
	Title            string  `json:"title"`
	URL              string  `json:"url"`
	Edition          Edition `json:"edition,omitempty"`
	Images           Images  `json:"images"`
}

type Fetcher interface {
	GetLimitedEditionDisplates() ([]Displate, error)
}

type apiClient struct {
}

func NewFetcher() Fetcher {
	return &apiClient{}
}

func (c *apiClient) GetLimitedEditionDisplates() ([]Displate, error) {
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
