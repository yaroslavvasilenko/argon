package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
)

// LocationDetailsResponse структура ответа от Nominatim API в формате /details
type LocationDetailsResponse struct {
	PlaceID     int64             `json:"place_id"`
	OsmType     string            `json:"osm_type"`
	OsmID       int64             `json:"osm_id"`
	Category    string            `json:"category"`
	Type        string            `json:"type"`
	DisplayName string            `json:"display_name,omitempty"`
	Names       map[string]string `json:"names"`
	AddressTags map[string]string `json:"addresstags"`
	Centroid    CentroidData      `json:"centroid"`
}

// CentroidData содержит координаты центра локации
type CentroidData struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// GetLocationByID получает информацию о локации по её ID из Nominatim
// Может принимать как OSM ID, так и place_id
func (l *Location) GetLocationByID(ctx context.Context, locationID string) (*LocationResponse, error) {
	// Пробуем сначала как place_id
	placeIDURL := fmt.Sprintf("%s/details?place_id=%s&format=json", l.baseUrl, locationID)
	
	// Затем подготовим URL для OSM ID
	osmID := locationID
	if !strings.HasPrefix(osmID, "N") && !strings.HasPrefix(osmID, "W") && !strings.HasPrefix(osmID, "R") {
		osmID = "R" + osmID
	}
	osmIDURL := fmt.Sprintf("%s/lookup?osm_ids=%s&format=json", l.baseUrl, osmID)

	// Сначала пробуем с place_id
	detailsResp, err := l.makeDetailsRequest(ctx, placeIDURL)
	if err == nil && detailsResp != nil {
		// Преобразуем ответ в формат LocationResponse
		return l.convertDetailsToLocationResponse(detailsResp), nil
	}

	// Если не получилось, пробуем с OSM ID
	fmt.Printf("Failed to get location by place_id, trying with osm_id. Error: %v\n", err)
	return l.makeRequest(ctx, osmIDURL)

	}

// makeDetailsRequest выполняет HTTP-запрос к API details и обрабатывает ответ
func (l *Location) makeDetailsRequest(ctx context.Context, url string) (*LocationDetailsResponse, error) {
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
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Пробуем разобрать ответ как одиночный объект
	var result LocationDetailsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(err, "unmarshaling response")
	}

	// Проверяем, что объект не пустой
	if result.PlaceID == 0 {
		return nil, errors.New("empty location details response")
	}

	return &result, nil
}

// convertDetailsToLocationResponse преобразует ответ в формате details в стандартный формат LocationResponse
func (l *Location) convertDetailsToLocationResponse(details *LocationDetailsResponse) *LocationResponse {
	// Создаем искусственный boundingbox на основе координат центра
	// Это необходимо для работы метода calculateRadiusFromBoundingBox
	lon := details.Centroid.Coordinates[0]
	lat := details.Centroid.Coordinates[1]
	
	// Создаем небольшой ограничивающий прямоугольник - примерно 100м во все стороны
	// 0.001 градуса это примерно 100 метров
	offset := 0.001
	boundingBox := []string{
		fmt.Sprintf("%f", lat-offset), // min lat
		fmt.Sprintf("%f", lat+offset), // max lat
		fmt.Sprintf("%f", lon-offset), // min lon
		fmt.Sprintf("%f", lon+offset), // max lon
	}
	
	// Создаем имя на основе доступных данных
	displayName := details.DisplayName
	if displayName == "" {
		// Если нет displayName, пробуем собрать из addressTags
		if street, ok := details.AddressTags["street"]; ok {
			displayName = street
			if housenumber, ok := details.AddressTags["housenumber"]; ok {
				displayName += ", " + housenumber
			}
		} else {
			// Если нет и этих данных, используем тип и ID
			displayName = fmt.Sprintf("%s %d", details.Type, details.PlaceID)
		}
	}
	
	return &LocationResponse{
		PlaceID:     details.PlaceID,
		Lat:         fmt.Sprintf("%f", lat),
		Lon:         fmt.Sprintf("%f", lon),
		DisplayName: displayName,
		BoundingBox: boundingBox,
	}
}

// makeRequest выполняет HTTP-запрос и обрабатывает ответ
func (l *Location) makeRequest(ctx context.Context, url string) (*LocationResponse, error) {
	fmt.Printf("Making request to: %s\n", url)

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

	fmt.Printf("Response status: %d, body: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Пробуем разобрать ответ как одиночный объект
	var result LocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// Если не получилось, пробуем как массив
		var results []LocationResponse
		if err2 := json.Unmarshal(body, &results); err2 != nil {
			return nil, errors.Wrap(err, "unmarshaling response")
		}

		if len(results) == 0 {
			return nil, errors.New("location not found")
		}

		return &results[0], nil
	}

	// Проверяем, что объект не пустой
	if result.PlaceID == 0 && result.DisplayName == "" {
		return nil, errors.New("empty location response")
	}

	return &result, nil
}
