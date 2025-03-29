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
		// Toml содержит данные категорий в формате TOML
		Toml string
		// Категории в структурированном виде
		Data CategoriesData
		Lang struct {
			Ru string
			En string
			Es string
		}
		// LangCharacteristics содержит переводы характеристик категорий
		LangCharacteristics struct {
			Ru string
			En string
			Es string
		}
		// LangOptions содержит переводы опций характеристик в формате JSON
		LangOptions struct {
			Ru string
			En string
			Es string
		}
		// OptionsTranslations содержит распарсенные переводы опций
		OptionsTranslations struct {
			Ru map[string]map[string]string
			En map[string]map[string]string
			Es map[string]map[string]string
		}
		// CharacteristicOptions содержит опции для характеристик
		CharacteristicOptions string
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

// CategoriesData представляет структуру данных категорий в TOML
type CategoriesData struct {
	Categories []CategoryNode `toml:"categories"`
}

// CategoryNode представляет узел категории
type CategoryNode struct {
	ID              string               `toml:"id"`
	Characteristics []CharacteristicNode `toml:"characteristics"`
	Subcategories   []CategoryNode       `toml:"subcategories"`
}

// CharacteristicNode представляет характеристику категории
type CharacteristicNode struct {
	Role    string   `toml:"role"`
	Options []string `toml:"options"`
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

	// Read categories.toml
	categoriesPath := filepath.Join(projectRoot, "./categories/categories.toml")
	categoriesFile, err := os.ReadFile(categoriesPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла categories.toml: %v", err)
	}

	cfg.Categories.Toml = string(categoriesFile)

	// Инициализируем map для ID категорий
	cfg.Categories.CategoryIds = make(map[string]bool)

	// Парсим категории и собираем их ID
	var categoriesData CategoriesData
	// Используем koanf для парсинга TOML
	catK := koanf.New(".")
	if err := catK.Load(file.Provider(categoriesPath), toml.Parser()); err != nil {
		log.Printf("Ошибка парсинга конфигурации категорий TOML: %v\n", err)
	} else if err := catK.Unmarshal("categories", &categoriesData.Categories); err != nil {
		log.Printf("Ошибка преобразования данных категорий: %v\n", err)
	} else {
		// Сохраняем структурированные данные
		cfg.Categories.Data = categoriesData

		// Рекурсивно собираем все ID категорий
		var collectCategoryIds func(nodes []CategoryNode)
		collectCategoryIds = func(nodes []CategoryNode) {
			for _, node := range nodes {
				cfg.Categories.CategoryIds[node.ID] = true
				collectCategoryIds(node.Subcategories)
			}
		}
		collectCategoryIds(categoriesData.Categories)
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

	// Загрузка переводов опций характеристик
	// Загрузка переводов опций характеристик
	optionsPath := filepath.Join(projectRoot, "./categories/lang_options/ru.json")
	optionsFile, err := os.ReadFile(optionsPath)
	if err != nil {
		log.Printf("Ошибка чтения файла lang_options/ru.json: %v", err)
	} else {
		cfg.Categories.LangOptions.Ru = string(optionsFile)
	}

	optionsPath = filepath.Join(projectRoot, "./categories/lang_options/en.json")
	optionsFile, err = os.ReadFile(optionsPath)
	if err != nil {
		log.Printf("Ошибка чтения файла lang_options/en.json: %v", err)
	} else {
		cfg.Categories.LangOptions.En = string(optionsFile)
	}

	optionsPath = filepath.Join(projectRoot, "./categories/lang_options/es.json")
	optionsFile, err = os.ReadFile(optionsPath)
	if err != nil {
		log.Printf("Ошибка чтения файла lang_options/es.json: %v", err)
	} else {
		cfg.Categories.LangOptions.Es = string(optionsFile)
	}

	// Инициализация структур для переводов опций
	cfg.Categories.OptionsTranslations.Ru = make(map[string]map[string]string)
	cfg.Categories.OptionsTranslations.En = make(map[string]map[string]string)
	cfg.Categories.OptionsTranslations.Es = make(map[string]map[string]string)

	// Парсим переводы из JSON
	if cfg.Categories.LangOptions.Ru != "" {
		if err := json.Unmarshal([]byte(cfg.Categories.LangOptions.Ru), &cfg.Categories.OptionsTranslations.Ru); err != nil {
			log.Printf("Ошибка при парсинге переводов опций на русском: %v", err)
		}
	}

	if cfg.Categories.LangOptions.En != "" {
		if err := json.Unmarshal([]byte(cfg.Categories.LangOptions.En), &cfg.Categories.OptionsTranslations.En); err != nil {
			log.Printf("Ошибка при парсинге переводов опций на английском: %v", err)
		}
	}

	if cfg.Categories.LangOptions.Es != "" {
		if err := json.Unmarshal([]byte(cfg.Categories.LangOptions.Es), &cfg.Categories.OptionsTranslations.Es); err != nil {
			log.Printf("Ошибка при парсинге переводов опций на испанском: %v", err)
		}
	}

	// Загрузка опций характеристик
	characteristicOptionsPath := filepath.Join(projectRoot, "./categories/characteristic_options.json")
	characteristicOptionsFile, err := os.ReadFile(characteristicOptionsPath)
	if err != nil {
		log.Printf("Ошибка чтения файла characteristic_options.json: %v", err)
	} else {
		cfg.Categories.CharacteristicOptions = string(characteristicOptionsFile)
	}

	// Загрузка переводов характеристик категорий
	langCharPath := filepath.Join(projectRoot, "./categories/lang_characteristics/ru.json")
	langCharFile, err := os.ReadFile(langCharPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/ru.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.Ru = string(langCharFile)

	langCharPath = filepath.Join(projectRoot, "./categories/lang_characteristics/en.json")
	langCharFile, err = os.ReadFile(langCharPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/en.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.En = string(langCharFile)

	langCharPath = filepath.Join(projectRoot, "./categories/lang_characteristics/es.json")
	langCharFile, err = os.ReadFile(langCharPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла lang_characteristics/es.json: %v", err)
	}

	cfg.Categories.LangCharacteristics.Es = string(langCharFile)
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
