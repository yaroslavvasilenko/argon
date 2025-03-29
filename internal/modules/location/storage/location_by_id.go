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

// GetLocationByID получает информацию о локации по её ID (place_id) из Nominatim
func (l *Location) GetLocationByID(ctx context.Context, placeID string) (*LocationResponse, error) {
	url := fmt.Sprintf("%s/lookup?osm_ids=R%s&format=json", l.baseUrl, placeID)

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

	// Nominatim возвращает массив объектов, нам нужен первый
	var results []LocationResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, errors.Wrap(err, "unmarshaling response")
	}

	if len(results) == 0 {
		return nil, errors.New("location not found")
	}

	return &results[0], nil
}
