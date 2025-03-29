package service

import (
	"context"
	"encoding/json"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
	"github.com/yaroslavvasilenko/argon/internal/models"
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

	// Получаем перевод категории из конфигурации
	lang := models.Localization(parser.GetLang(ctx))
	switch lang {
	case models.LanguageRu:
		var categories map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.Categories.Lang.Ru), &categories); err == nil {
			if catData, ok := categories[categoryID].(map[string]interface{}); ok {
				if name, ok := catData["name"].(string); ok && name != "" {
					category.Name = name
				}
			}
		}

	case models.LanguageEn:
		var categories map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.Categories.Lang.En), &categories); err == nil {
			if catData, ok := categories[categoryID].(map[string]interface{}); ok {
				if name, ok := catData["name"].(string); ok && name != "" {
					category.Name = name
				}
			}
		}

	case models.LanguageEs:
		var categories map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.Categories.Lang.Es), &categories); err == nil {
			if catData, ok := categories[categoryID].(map[string]interface{}); ok {
				if name, ok := catData["name"].(string); ok && name != "" {
					category.Name = name
				}
			}
		}

	}

	// Если перевод не найден, продолжаем использовать ID как имя
	// В будущем можно добавить поле Name в структуру CategoryNode

	return category, nil
}
