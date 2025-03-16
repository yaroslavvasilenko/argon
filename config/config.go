package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fmt"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Port string
	DB   struct {
		Url string `koanf:"url"`
	}
	Logger struct {
		Level string
	}

	Categories struct {
		Json string
		Lang struct {
			Ru string
			En string
			Es string
		}
		// CategoryCharacteristics содержит JSON с характеристиками категорий
		Characteristics string
		// LangCharacteristics содержит переводы характеристик категорий
		LangCharacteristics struct {
			Ru string
			En string
			Es string
		}
		// CategoryIds содержит все доступные ID категорий для быстрой валидации
		CategoryIds map[string]bool
	}
	Binance struct {
		APIKey    string
		SecretKey string
		Local     bool
	}
	Nominatim struct {
		BaseUrl string
	}
}

var cfg = Config{}

type CategoryNode struct {
	Category struct {
		ID string `json:"id"`
	} `json:"category"`
	Subcategories []CategoryNode `json:"subcategories,omitempty"`
}

func LoadConfig() {
	// Get project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		log.Fatalf("Ошибка определения корневой директории проекта: %v", err)
	}

	// Initialize Koanf
	k := koanf.New(".")

	// Load config from config.toml first
	configPath := filepath.Join(projectRoot, "config", "config.toml")
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		log.Printf("Warning: error loading config.toml: %v", err)
	}

	// Load .env file if exists
	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Info: .env file not found: %v", err)
	}

	// Load environment variables with prefix APP_
	err = k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "APP_")), "_", ".", -1)
	}), nil)
	if err != nil {
		log.Fatalf("Ошибка при загрузке переменных окружения: %v", err)
	}

	// Unmarshal config into struct
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("Ошибка при разборе конфигурации: %v", err)
	}

	// Read categories.json
	categoriesPath := filepath.Join(projectRoot, "./categories/categories.json")
	categoriesFile, err := os.ReadFile(categoriesPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories.json: %v", err)
	}

	cfg.Categories.Json = string(categoriesFile)

	// Инициализируем map для ID категорий
	cfg.Categories.CategoryIds = make(map[string]bool)

	// Парсим категории и собираем их ID
	var categories []CategoryNode
	if err := json.Unmarshal(categoriesFile, &categories); err != nil {
		log.Printf("Ошибка парсинга конфигурации категорий: %v\n", err)
	} else {
		// Рекурсивно собираем все ID категорий
		var collectCategoryIds func(nodes []CategoryNode)
		collectCategoryIds = func(nodes []CategoryNode) {
			for _, node := range nodes {
				cfg.Categories.CategoryIds[node.Category.ID] = true
				collectCategoryIds(node.Subcategories)
			}
		}
		collectCategoryIds(categories)
	}

	// Загрузка переводов категорий
	categoriesPath = filepath.Join(projectRoot, "./categories/lang/ru.json")
	categoriesFile, err = os.ReadFile(categoriesPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories/lang/ru.json: %v", err)
	}

	cfg.Categories.Lang.Ru = string(categoriesFile)

	categoriesPath = filepath.Join(projectRoot, "./categories/lang/en.json")
	categoriesFile, err = os.ReadFile(categoriesPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories/lang/en.json: %v", err)
	}

	cfg.Categories.Lang.En = string(categoriesFile)

	categoriesPath = filepath.Join(projectRoot, "./categories/lang/es.json")
	categoriesFile, err = os.ReadFile(categoriesPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories/lang/es.json: %v", err)
	}

	cfg.Categories.Lang.Es = string(categoriesFile)

	// Загрузка характеристик категорий
	characteristicsPath := filepath.Join(projectRoot, "./categories/category_characteristics.json")
	characteristicsFile, err := os.ReadFile(characteristicsPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла category_characteristics.json: %v", err)
	}

	cfg.Categories.Characteristics = string(characteristicsFile)

	// Загрузка переводов характеристик категорий
	characteristicsPath = filepath.Join(projectRoot, "./categories/lang_characteristics/ru.json")
	characteristicsFile, err = os.ReadFile(characteristicsPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/ru.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.Ru = string(characteristicsFile)

	characteristicsPath = filepath.Join(projectRoot, "./categories/lang_characteristics/en.json")
	characteristicsFile, err = os.ReadFile(characteristicsPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/en.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.En = string(characteristicsFile)

	characteristicsPath = filepath.Join(projectRoot, "./categories/lang_characteristics/es.json")
	characteristicsFile, err = os.ReadFile(characteristicsPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/es.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.Es = string(characteristicsFile)
}

// getProjectRoot returns the absolute path to the project root directory
func getProjectRoot() (string, error) {
	// Try to find go.mod file by walking up the directory tree
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("go.mod не найден, невозможно определить корень проекта")
		}
		currentDir = parentDir
	}
}

func GetConfig() Config {
	return cfg
}
