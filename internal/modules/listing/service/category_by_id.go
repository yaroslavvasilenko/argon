package service

import (
	"context"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

// GetCategoryById получает информацию о категории по её ID
func (s *Listing) GetCategoryById(ctx context.Context, categoryID string) (listing.Category, error) {
	if categoryID == "" {
		return listing.Category{}, nil
	}

	// Получаем конфигурацию
	cfg := config.GetConfig()

	// Ищем категорию по ID в дереве категорий
	var findCategory func(nodes []config.CategoryNode, id string) *config.CategoryNode
	findCategory = func(nodes []config.CategoryNode, id string) *config.CategoryNode {
		for i := range nodes {
			if nodes[i].ID == id {
				return &nodes[i]
			}
			if len(nodes[i].Subcategories) > 0 {
				if found := findCategory(nodes[i].Subcategories, id); found != nil {
					return found
				}
			}
		}
		return nil
	}

	// Ищем категорию в дереве
	categoryNode := findCategory(cfg.Categories.Data.Categories, categoryID)
	if categoryNode == nil {
		return listing.Category{}, nil
	}

	// Создаем объект категории
	category := listing.Category{
		ID:   categoryID,
		Name: categoryID, // По умолчанию используем ID как имя
	}

	// Получаем язык из контекста
	lang := parser.GetLang(ctx)

	// Получаем перевод имени категории в зависимости от языка
	// В реальном приложении здесь должна быть логика получения перевода из конфига
	// Сейчас мы просто используем ID как имя
	switch lang {
	case "ru":
		// Здесь должна быть логика получения русского перевода
		category.Name = "Категория " + categoryID
	case "en":
		// Здесь должна быть логика получения английского перевода
		category.Name = "Category " + categoryID
	case "es":
		// Здесь должна быть логика получения испанского перевода
		category.Name = "Categoría " + categoryID
	default:
		category.Name = "Category " + categoryID
	}

	return category, nil
}
