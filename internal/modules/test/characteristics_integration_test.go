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

		// Создаем карту для быстрого поиска индекса характеристики по её роли
		charRoleToIndex := make(map[string]int)
		for i, char := range characteristics {
			charRoleToIndex[char.Role] = i
		}

		// Проверяем, что характеристики приходят в порядке иерархии категорий
		// Согласно файлу category_characteristics.json:
		// - Для категории "electronics": ["price", "brand", "condition", "stocked", "weight"]
		// - Для категории "smartphones": ["color", "height", "width", "depth", "volume"]
		// - Для категории "iphone": []

		// Проверяем порядок характеристик в рамках категории electronics
		electronicsChars := []string{"price", "brand", "condition", "stocked", "weight"}
		for i := 0; i < len(electronicsChars)-1; i++ {
			if firstIndex, hasFirst := charRoleToIndex[electronicsChars[i]]; hasFirst {
				if secondIndex, hasSecond := charRoleToIndex[electronicsChars[i+1]]; hasSecond {
					assert.Less(t, firstIndex, secondIndex, 
						"Характеристика '%s' должна идти раньше '%s' в категории electronics", 
						electronicsChars[i], electronicsChars[i+1])
				}
			}
		}

		// Проверяем порядок характеристик в рамках категории smartphones
		smartphonesChars := []string{"color", "height", "width", "depth", "volume"}
		for i := 0; i < len(smartphonesChars)-1; i++ {
			if firstIndex, hasFirst := charRoleToIndex[smartphonesChars[i]]; hasFirst {
				if secondIndex, hasSecond := charRoleToIndex[smartphonesChars[i+1]]; hasSecond {
					assert.Less(t, firstIndex, secondIndex, 
						"Характеристика '%s' должна идти раньше '%s' в категории smartphones", 
						smartphonesChars[i], smartphonesChars[i+1])
				}
			}
		}

		// Проверяем, что все характеристики из категории electronics идут раньше характеристик из smartphones
		for _, elChar := range electronicsChars {
			if elIndex, hasElChar := charRoleToIndex[elChar]; hasElChar {
				for _, smChar := range smartphonesChars {
					if smIndex, hasSmChar := charRoleToIndex[smChar]; hasSmChar {
						assert.Less(t, elIndex, smIndex, 
							"Характеристика '%s' из категории electronics должна идти раньше '%s' из smartphones", 
							elChar, smChar)
					}
				}
			}
		}
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
		
		// Проверяем, что роли характеристик совпадают и идут в том же порядке
		for i, charEn := range characteristicsEn {
			assert.Equal(t, charEn.Role, characteristicsRu[i].Role, 
				"Роль характеристики должна быть одинаковой для разных языков в позиции %d", i)
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
