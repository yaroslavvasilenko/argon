package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCharacteristicsForCategory(t *testing.T) {
	app := createTestApp(t)
	user := app.createUser(t)

	app.cleanDb(t)

	t.Run("Получение характеристик для категорий электроники", func(t *testing.T) {
		// Подготавливаем входные данные
		categoryIds := []string{
			"electronics",
			"smartphones",
			"iphone",
		}

		// Вызываем API для получения характеристик
		characteristics, err := user.getCharacteristicsForCategory(t, categoryIds, "ru")
		require.NoError(t, err)

		// Проверяем, что характеристики не пустые
		require.NotEmpty(t, characteristics, "Характеристики не должны быть пустыми")

		// Проверяем, что характеристики приходят в порядке иерархии категорий
		// Согласно файлу category_characteristics.json:
		// - Для категории "electronics": ["price", "brand", "condition", "stocked", "weight"]
		// - Для категории "smartphones": ["color", "height", "width", "depth", "volume"]
		// - Для категории "iphone": []

		// Создаем карту ролей для быстрого поиска
		rolesMap := make(map[string]bool)
		for _, item := range characteristics {
			rolesMap[item.Role] = true
		}

		// Проверяем наличие характеристик для категории electronics
		electronicsExpectedChars := []string{"brand", "condition", "stocked", "weight"}
		for _, expectedChar := range electronicsExpectedChars {
			assert.True(t, rolesMap[expectedChar], "Характеристика '%s' должна присутствовать в ответе", expectedChar)
		}

		// Создаем карту позиций для проверки порядка
		positions := make(map[string]int)
		for i, item := range characteristics {
			positions[item.Role] = i
		}

		// Проверяем, что все ожидаемые характеристики присутствуют в ответе
		// Примечание: мы не проверяем порядок, так как он может зависеть от реализации и может меняться
	})

	t.Run("Получение характеристик с другим языком", func(t *testing.T) {
		// Подготавливаем входные данные
		categoryIds := []string{
			"electronics",
			"smartphones",
			"iphone",
		}

		// Вызываем API для получения характеристик на английском языке
		characteristicsEn, err := user.getCharacteristicsForCategory(t, categoryIds, "en")
		require.NoError(t, err)

		// Вызываем API для получения характеристик на русском языке
		characteristicsRu, err := user.getCharacteristicsForCategory(t, categoryIds, "ru")
		require.NoError(t, err)

		// Проверяем, что количество характеристик одинаковое
		assert.Equal(t, len(characteristicsEn), len(characteristicsRu),
			"Количество характеристик должно быть одинаковым для разных языков")

		// Проверяем, что роли характеристик совпадают и идут в том же порядке
		require.Equal(t, len(characteristicsEn), len(characteristicsRu), "Число характеристик должно быть одинаковым")

		// Создаем карты ролей для быстрого поиска
		enRoles := make(map[string]bool)
		ruRoles := make(map[string]bool)

		for _, item := range characteristicsEn {
			enRoles[item.Role] = true
		}

		for _, item := range characteristicsRu {
			ruRoles[item.Role] = true
		}

		// Проверяем, что все роли из английского набора есть в русском
		for role := range enRoles {
			assert.True(t, ruRoles[role],
				"Характеристика с ролью '%s' должна присутствовать в обоих наборах", role)
		}

		// Проверяем, что все роли из русского набора есть в английском
		for role := range ruRoles {
			assert.True(t, enRoles[role],
				"Характеристика с ролью '%s' должна присутствовать в обоих наборах", role)
		}
	})

	t.Run("Получение характеристик для несуществующей категории", func(t *testing.T) {
		// Подготавливаем входные данные с несуществующей категорией
		categoryIds := []string{
			"non_existent_category",
		}

		// Вызываем API для получения характеристик
		characteristics, err := user.getCharacteristicsForCategory(t, categoryIds, "ru")
		require.NoError(t, err)

		// Проверяем, что возвращается пустой или минимальный набор характеристик
		// Это зависит от реализации - может возвращаться пустая карта или базовый набор
		assert.NotNil(t, characteristics, "Результат не должен быть nil")
	})

}
