package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/boost"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

// Метод для получения буста объявления
func (u *user) getBoost(t *testing.T, listingID uuid.UUID) (*http.Response, error) {
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/boost/%s", listingID),
		nil,
	)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := u.fiber.Test(req, -1)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// Метод для обновления буста объявления
func (u *user) updateBoost(t *testing.T, listingID uuid.UUID, boosts []models.BoostType) (*http.Response, error) {
	updateReq := boost.UpdateBoostRequest{
		ListingID: listingID,
		Boosts:    boosts,
	}
	
	body, err := json.Marshal(updateReq)
	if err != nil {
		return nil, err
	}
	
	req := httptest.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/boost/%s", listingID),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := u.fiber.Test(req, -1)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// Тест для проверки работы контроллера буста
func TestBoostController(t *testing.T) {
	app := createTestApp(t)
	defer app.cleanDb(t)
	
	user := app.createUser(t)
	
	// Создаем тестовое объявление
	listingReq := listing.CreateListingRequest{
		Title:       "Тестовое объявление для буста",
		Description: "Описание тестового объявления для проверки работы буста",
		Price:       1000.0,
		Currency:    models.RUB,
		Location: models.Location{
			Name: "Москва, Россия",
			Area: models.Area{
				Coordinates: struct {
					Lat float64 `json:"lat" validate:"required"`
					Lng float64 `json:"lng" validate:"required"`
				}{
					Lat: 55.7558,
					Lng: 37.6173,
				},
				Radius: 10000,
			},
		},
		Categories: []string{"electronics"},
	}
	
	// Создаем объявление
	resp := user.createListing(t, listingReq)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Статус ответа должен быть 200 OK")
	
	// Получаем ID созданного объявления
	var createResp listing.CreateListingResponse
	err := json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err, "Ошибка при декодировании ответа")
	require.NotEqual(t, uuid.Nil, createResp.ID, "ID объявления не должен быть пустым")
	
	listingID := createResp.ID
	
	t.Run("Получение бустов для нового объявления", func(t *testing.T) {
		// Получаем текущие бусты объявления
		resp, err := user.getBoost(t, listingID)
		require.NoError(t, err, "Ошибка при получении бустов")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Status code should be 200 OK")
		
		var getResp boost.GetBoostResponse
		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Ошибка при декодировании ответа")
		
		assert.Len(t, getResp.AvailableBoosts, 3,"available_boosts should have 3 elements")
		
		assert.Empty(t, getResp.EnableBoostTypes, "enable_boost_types should be empty")
	})
	
	t.Run("Добавление бустов для объявления", func(t *testing.T) {
		// Добавляем бусты для объявления
		boosts := []models.BoostType{models.BoostTypeHighlight, models.BoostTypeUpfront}
		
		resp, err := user.updateBoost(t, listingID, boosts)
		require.NoError(t, err, "Ошибка при обновлении бустов")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Статус ответа должен быть 200 OK")
		
		var updateResp boost.GetBoostResponse
		err = json.NewDecoder(resp.Body).Decode(&updateResp)
		require.NoError(t, err, "Ошибка при декодировании ответа")
		
		// Проверяем, что бусты были добавлены
		require.Len(t, updateResp.EnableBoostTypes, 2, "Должно быть добавлено 2 буста")
		
		// Проверяем, что добавлены правильные типы бустов
		assert.Contains(t, updateResp.EnableBoostTypes, models.BoostTypeHighlight, "Должен быть добавлен буст типа Highlight")
		assert.Contains(t, updateResp.EnableBoostTypes, models.BoostTypeUpfront, "Должен быть добавлен буст типа Upfront")
		
		// Проверяем, что все доступные типы бустов присутствуют
		require.Len(t, updateResp.AvailableBoosts, 3, "Должно быть 3 доступных типа буста")
		
		// Проверяем, что комиссии установлены правильно для всех доступных бустов
		commissions := models.GetBoostTypesWithCommissions()
		for _, b := range updateResp.AvailableBoosts {
			expectedCommission := commissions[b.Type] // Комиссия уже в десятичном формате
			assert.Equal(t, expectedCommission, b.CommissionPercent, 
				"Комиссия для буста типа %s должна быть %.2f", b.Type, expectedCommission)
		}
	})
	
	t.Run("Обновление бустов для объявления", func(t *testing.T) {
		// Обновляем бусты для объявления, оставляем только один тип
		boosts := []models.BoostType{models.BoostTypeBase}
		
		resp, err := user.updateBoost(t, listingID, boosts)
		require.NoError(t, err, "Ошибка при обновлении бустов")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Статус ответа должен быть 200 OK")
		
		var updateResp boost.GetBoostResponse
		err = json.NewDecoder(resp.Body).Decode(&updateResp)
		require.NoError(t, err, "Ошибка при декодировании ответа")
		
		// Проверяем, что бусты были обновлены
		require.Len(t, updateResp.EnableBoostTypes, 1, "Должен остаться только 1 активный буст")
		assert.Equal(t, models.BoostTypeBase, updateResp.EnableBoostTypes[0], "Должен остаться только буст типа Base")
		
		// Проверяем, что все доступные типы бустов присутствуют
		require.Len(t, updateResp.AvailableBoosts, 3, "Должно быть 3 доступных типа буста")
		
		// Проверяем, что комиссии установлены правильно
		commissions := models.GetBoostTypesWithCommissions()
		
		// Находим буст типа Base в доступных бустах
		var baseBoost *boost.BoostResp
		for i, b := range updateResp.AvailableBoosts {
			if b.Type == models.BoostTypeBase {
				baseBoost = &updateResp.AvailableBoosts[i]
				break
			}
		}
		
		require.NotNil(t, baseBoost, "Буст типа Base должен присутствовать в доступных бустах")
		expectedCommission := commissions[models.BoostTypeBase] // Комиссия уже в десятичном формате
		assert.Equal(t, expectedCommission, baseBoost.CommissionPercent, 
			"Комиссия для буста типа %s должна быть %.2f", models.BoostTypeBase, expectedCommission)
	})
	
	t.Run("Проверка ошибки при неверном ID объявления", func(t *testing.T) {
		// Пробуем обновить бусты для несуществующего объявления
		boosts := []models.BoostType{models.BoostTypeBase}
		
		resp, err := user.updateBoost(t, uuid.New(), boosts)
		require.NoError(t, err, "Ошибка при обновлении бустов")
		
		// Ожидаем ошибку, так как объявление не существует
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, 
			"Статус ответа должен быть 500 Internal Server Error для несуществующего объявления")
	})
	
	t.Run("Проверка ошибки при пустом ID объявления", func(t *testing.T) {
		// Пробуем обновить бусты с пустым ID объявления
		boosts := []models.BoostType{models.BoostTypeBase}
		
		resp, err := user.updateBoost(t, uuid.Nil, boosts)
		require.NoError(t, err, "Ошибка при обновлении бустов")
		
		// Ожидаем ошибку, так как ID объявления пустой
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, 
			"Статус ответа должен быть 400 Bad Request для пустого ID объявления")
	})
}
