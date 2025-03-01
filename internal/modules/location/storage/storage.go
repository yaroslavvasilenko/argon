package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
)

type LocationResponse struct {
	PlaceID     int64    `json:"place_id"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Address     Address  `json:"address"`
	BoundingBox []string `json:"boundingbox"`
}

type Address struct {
	HouseNumber string `json:"house_number,omitempty"`
	Road        string `json:"road,omitempty"`
	Suburb      string `json:"suburb,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	Country     string `json:"country,omitempty"`
}

type Location struct {
	baseUrl string
}

func NewLocation(baseUrlNominatim string) *Location {
	return &Location{
		baseUrl: baseUrlNominatim,
	}
}

func (l *Location) GetLocation(ctx context.Context, lat, lng float64, zoom int) (*LocationResponse, error) {
	url := fmt.Sprintf("%s/reverse?lat=%f&lon=%f&zoom=%d&format=json", l.baseUrl, lat, lng, zoom)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	language := parser.GetLang(ctx)

	if language != "" {
		req.Header.Set("Accept-Language", language)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result LocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(err, "unmarshaling response")
	}

	return &result, nil
}
