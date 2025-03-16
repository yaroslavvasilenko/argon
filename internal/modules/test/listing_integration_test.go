package modules

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

func TestCreateListing(t *testing.T) {
	// Инициализация тестовой БД и роутера
	app := createTestApp(t)
	user := app.createUser(t)
	t.Run("Success create listing", func(t *testing.T) {
		listingInput := listing.CreateListingRequest{
			Title:       "Тестовая квартира",
			Description: "Просторная квартира в центре",
			Price:       1000000,
			Currency:    models.RUB,
		}

		resp := user.createListing(t, listingInput)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		listOut := models.Listing{}
		json.NewDecoder(resp.Body).Decode(&listOut)

		// Проверка ответа
		assert.NotZero(t, listOut.ID)
		assert.Equal(t, listingInput.Title, listOut.Title)
		assert.Equal(t, listingInput.Description, listOut.Description)
		assert.Equal(t, listingInput.Price, listOut.Price)
		assert.Equal(t, listingInput.Currency, listOut.Currency)

		// Проверка времени в UTC
		// now := time.Now().UTC()
		// assert.WithinDuration(t, now, listOut.CreatedAt.UTC(), 2*time.Second)
		// assert.WithinDuration(t, now, listOut.UpdatedAt.UTC(), 2*time.Second)

		assert.Empty(t, listOut.ViewsCount)
		assert.Nil(t, listOut.DeletedAt)
	})
}
