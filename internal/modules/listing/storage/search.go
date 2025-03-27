package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
)

func (s *Listing) SearchListingsByTitle(ctx context.Context, query string, limit int, cursorID *uuid.UUID, sort, categoryID string, filters models.Filters, location models.Location) (*models.Listing, []models.ListingResult, error) {
	// Если limit == 0, возвращаем пустой результат
	if limit == 0 {
		return nil, []models.ListingResult{}, nil
	}

	if sort == "" {
		sort = "relevance"
	}

	var cursor *models.Listing
	if cursorID != nil {
		var cursorListing models.Listing
		if err := s.gorm.Table(itemTable).WithContext(ctx).
			Where("id = ? AND deleted_at IS NULL", cursorID).
			First(&cursorListing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fiber.NewError(fiber.StatusNotFound, "Курсор не найден")
			}
			return nil, nil, err
		}
		cursor = &cursorListing
	}

	// Определяем тип поиска на основе запроса
	searchType := determineSearchType(query)
	baseQuery := buildBaseQuery(searchType, categoryID, filters, location)

	orderExpr := getSortExpression(sort, searchType)
	searchQuery := createSearchQuery(query, searchType)

	sqlQuery, queryArgs := buildSQLQuery(baseQuery, orderExpr, searchQuery, limit, cursor, searchType)

	rows, err := s.pool.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	listings, err := s.scanListings(rows)
	if err != nil {
		return nil, nil, err
	}

	// Получаем полные данные для каждого объявления
	listingResults := make([]models.ListingResult, 0, len(listings))
	for _, listing := range listings {
		result, err := s.getListingWithRelatedData(ctx, listing)
		if err != nil {
			return nil, nil, err
		}
		listingResults = append(listingResults, result)
	}

	return cursor, listingResults, nil
}

// Тип поиска, определяющий какой алгоритм поиска использовать
type SearchType int

const (
	// Только нечеткий поиск
	FuzzySearch SearchType = iota
	// Только полнотекстовый поиск
	FullTextSearch
	// Комбинированный поиск (нечеткий + полнотекстовый)
	CombinedSearch
)

// buildBaseQuery создает базовый SQL запрос в зависимости от типа поиска
func buildBaseQuery(searchType SearchType, categoryID string, filters models.Filters, location models.Location) string {
	var categoryFilter string
	if categoryID != "" {
		categoryFilter = `
			AND EXISTS (
				SELECT 1 FROM listing_categories lc 
				WHERE lc.listing_id = l.id AND lc.category_id = '` + categoryID + `'
			)`
	}
	
	// Добавляем фильтр по локации, если указаны координаты
	var locationFilter string
	if location.Area.Coordinates.Lat != 0 && location.Area.Coordinates.Lng != 0 && location.Area.Radius > 0 {
		// Используем функцию ST_DWithin для поиска в радиусе
		// Преобразуем координаты в географические точки и вычисляем расстояние в метрах
		locationFilter = fmt.Sprintf(`
			AND EXISTS (
				SELECT 1 FROM locations loc
				WHERE loc.listing_id = l.id
				AND ST_DWithin(
					ST_SetSRID(ST_MakePoint(loc.longitude, loc.latitude), 4326)::geography,
					ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography,
					%d
				)
			)`, 
			location.Area.Coordinates.Lng, location.Area.Coordinates.Lat, location.Area.Radius)
	}

	// Добавляем фильтр по характеристикам, если они указаны
	
	var filtersFilter string
	if len(filters) > 0 {
		filtersFilter = `
			AND EXISTS (
				SELECT 1 FROM listing_characteristics lch
				WHERE lch.listing_id = l.id
				AND (`

		filterConditions := []string{}

		for key, _ := range filters {
			if priceFilter, ok := filters.GetPriceFilter(key); ok {
				filterConditions = append(filterConditions, fmt.Sprintf("l.price >= %d",  priceFilter.Min))
				filterConditions = append(filterConditions, fmt.Sprintf("l.price <= %d", priceFilter.Max))
			}
		

		// Обрабатываем фильтр цвета
			if colorFilter, ok := filters.GetColorFilter(key); ok && len(colorFilter.Options) > 0 {
				colorConditions := []string{}
				for _, color := range colorFilter.Options {
					// Проверяем, содержит ли массив цветов заданный цвет
					colorConditions = append(colorConditions, fmt.Sprintf(
						"lch.characteristics -> '%s' ? '%s'", key, color))
				}
				filterConditions = append(filterConditions, "("+strings.Join(colorConditions, " OR ")+")")
			}
		

		// Обрабатываем фильтр выпадающего списка
			if dropdownFilter, ok := filters.GetDropdownFilter(key); ok && len(dropdownFilter) > 0 {
				dropdownConditions := []string{}
				for _, option := range dropdownFilter {
					// Проверяем, содержит ли массив разрешений экрана заданное разрешение
					dropdownConditions = append(dropdownConditions, fmt.Sprintf(
						"lch.characteristics -> '%s' ? '%s'", key, option))
				}

				filterConditions = append(filterConditions, "("+strings.Join(dropdownConditions, " OR ")+")")
			}
		

		// Обрабатываем фильтр чекбокса
			if checkboxFilter, ok := filters.GetCheckboxFilter(key); ok && checkboxFilter != nil {
				filterConditions = append(filterConditions, fmt.Sprintf(
					"(lch.characteristics ->> '%s')::boolean = %t", key, *checkboxFilter))
			}
		

		// Обрабатываем фильтр размеров
			if dimensionFilter, ok := filters.GetDimensionFilter(key); ok {
				filterConditions = append(filterConditions, fmt.Sprintf(
					"(lch.characteristics ->> '%s')::float >= %f", key, float64(dimensionFilter.Min)))
				filterConditions = append(filterConditions, fmt.Sprintf(
					"(lch.characteristics ->> '%s')::float <= %f", key, float64(dimensionFilter.Max)))
			}
		}

		// Если есть условия фильтрации, добавляем их в запрос
		if len(filterConditions) > 0 {
			filtersFilter += strings.Join(filterConditions, " AND ") + ")"
		} else {
			// Если нет условий, просто проверяем наличие записи в таблице характеристик
			filtersFilter += "true)"
		}

		filtersFilter += ")"
	}

	switch searchType {
	case FuzzySearch:
		// Запрос с использованием триграмм (pg_trgm) для нечеткого поиска
		return `
			SELECT ` + listingFields + `
			FROM ` + itemTable + ` l
			WHERE l.deleted_at IS NULL
			AND (
				/* Используем оператор % для поиска с опечатками */
				l.title % $1 OR
				/* similarity возвращает значение от 0 до 1, где 1 означает полное совпадение */
				similarity(l.title, $1) > 0.3 OR
				/* word_similarity сравнивает слова, а не символы */
				word_similarity($1, l.title) > 0.4
			)` + categoryFilter + locationFilter + filtersFilter + `
		`
	case FullTextSearch:
		// Стандартный поиск с использованием полнотекстового индекса
		return `
			SELECT ` + listingFields + `
			FROM ` + itemTable + ` l
			JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE l.deleted_at IS NULL
			AND to_tsquery('russian', $1) @@ lsr.title_vector` + categoryFilter + locationFilter + filtersFilter + `
		`
	case CombinedSearch:
		// Комбинированный поиск, использующий оба метода с ранжированием результатов
		return `
			SELECT ` + listingFields + `,
				/* Вычисляем комбинированную релевантность */
				(
					/* Вес для нечеткого поиска (0.6) */
					0.6 * COALESCE(similarity(l.title, $1), 0) +
					0.4 * COALESCE(word_similarity($1, l.title), 0) +
					/* Вес для полнотекстового поиска (0.4) */
					0.4 * COALESCE(ts_rank(lsr.title_vector, to_tsquery('russian', $2)), 0)
				) AS combined_rank
			FROM ` + itemTable + ` l
			LEFT JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE l.deleted_at IS NULL
			AND (
				/* Нечеткий поиск */
				l.title % $1 OR
				similarity(l.title, $1) > 0.3 OR
				word_similarity($1, l.title) > 0.4 OR
				/* Полнотекстовый поиск */
				to_tsquery('russian', $2) @@ lsr.title_vector
			)` + categoryFilter + locationFilter + filtersFilter + `
		`
	default:
		// По умолчанию используем нечеткий поиск
		return `
			SELECT ` + listingFields + `
			FROM ` + itemTable + ` l
			WHERE l.deleted_at IS NULL
			AND (
				l.title % $1 OR
				similarity(l.title, $1) > 0.3 OR
				word_similarity($1, l.title) > 0.4
			)` + categoryFilter + locationFilter + filtersFilter + `
		`
	}
}

// buildSQLQuery формирует итоговый SQL запрос и набор аргументов для выполнения запроса с пагинацией.
// Параметры:
//
//	baseQuery  - базовый SQL запрос без условий сортировки и пагинации
//	orderExpr  - выражение сортировки, определяющее порядок возвращаемых записей
//	searchQuery- поисковый запрос, используемый в полном текстовом поиске
//	limit      - лимит на количество возвращаемых записей (положительный для следующей страницы, отрицательный для предыдущей)
//	cursor     - объект, представляющий запись-курсор для пагинации
//	searchType - тип поиска (нечеткий, полнотекстовый или комбинированный)
//
// Функция возвращает сформированный SQL запрос и срез аргументов для подстановки в запрос.
func buildSQLQuery(baseQuery, orderExpr, searchQuery string, limit int, cursor *models.Listing, searchType SearchType) (string, []interface{}) {
	var args []interface{}

	var tsQueryVersion string
	switch searchType {
	case FuzzySearch, FullTextSearch:
		// Для нечеткого или полнотекстового поиска используем один параметр
		args = []interface{}{searchQuery}
	case CombinedSearch:
		// Для комбинированного поиска используем два параметра:
		// - оригинальный запрос для нечеткого поиска
		// - обработанный запрос для полнотекстового поиска
		
		// Используем prepareTsQuery для очистки запроса от специальных символов
		tsQueryVersion = prepareTsQuery(searchQuery)
		args = []interface{}{searchQuery, tsQueryVersion}
	}

	if cursor == nil {
		// Если курсор не задан, выборка начинается с начала набора результатов или с конца, в зависимости от знака лимита.
		if limit > 0 {
			// Лимит положительный: выбираем первые limit записей, сортируя результат по заданному orderExpr
			
			// Определяем номер параметра для LIMIT в зависимости от типа поиска
			limitParam := "$2"
			if searchType == CombinedSearch {
				limitParam = "$3"
			}
			
			return baseQuery + `
				ORDER BY ` + orderExpr + `
				LIMIT ` + limitParam + `
			`, append(args, limit)
		} else {
			// Лимит отрицательный: выбираем последние -limit записей.
			// Для этого выполняем следующие шаги:
			// 1. Сортируем базовый запрос в обратном порядке (reverse orderExpr).
			// 2. Ограничиваем выборку до -limit записей.
			// 3. Внешний запрос переворачивает результат для восстановления исходного порядка.
			reverseExpr := getReverseOrderExpression(orderExpr)
			
			// Определяем номер параметра для LIMIT в зависимости от типа поиска
			limitParam := "$2"
			if searchType == CombinedSearch {
				limitParam = "$3"
			}
			
			sql := `
			WITH reversed AS (
				` + baseQuery + `
				ORDER BY ` + reverseExpr + `
				LIMIT ` + limitParam + `
			)
			SELECT ` + listingFields + ` FROM reversed l
			ORDER BY ` + orderExpr + `
			`
			return sql, append(args, -limit)
		}
	} else {
		// Если курсор задан, значит выборка должна быть смещена относительно курсора для пагинации.
		if limit > 0 {
			// Положительный лимит: выбираем записи, следующие за курсором (без включения самой записи-курсор).
			// Функция getCursorCondition формирует условие, исключающее курсор из результата.
			cond := getCursorCondition(orderExpr, cursor, false)
			
			// Определяем номер параметра для LIMIT в зависимости от типа поиска
			limitParam := "$2"
			if searchType == CombinedSearch {
				limitParam = "$3"
			}
			
			return baseQuery + `
				AND ` + cond + `
				ORDER BY ` + orderExpr + `
				LIMIT ` + limitParam + `
			`, append(args, limit)
		} else {
			// Лимит отрицательный: выбираем записи, предшествующие курсору, включая его.
			// В этом случае:
			// 1. Используем обратное сортировочное выражение для формирования условия.
			// 2. Формируем условие, включающее курсор (inclusive = true).
			// 3. Для корректного порядка результатов затем переворачиваем выборку обратно.
			reverseExpr := getReverseOrderExpression(orderExpr)
			cond := getCursorCondition(reverseExpr, cursor, true)
			
			// Определяем номер параметра для LIMIT в зависимости от типа поиска
			limitParam := "$2"
			if searchType == CombinedSearch {
				limitParam = "$3"
			}
			
			sql := `
			WITH reversed AS (
				` + baseQuery + `
				AND ` + cond + `
				ORDER BY ` + reverseExpr + `
				LIMIT ` + limitParam + `
			)
			SELECT ` + listingFields + ` FROM reversed l
			ORDER BY ` + orderExpr + `
			`
			return sql, append(args, -limit)
		}
	}
}

func getSortExpression(sort string, searchType SearchType) string {
	var orderExpr string
	switch sort {
	case models.SORT_PRICE_ASC:
		orderExpr = "l.price ASC"
	case models.SORT_PRICE_DESC:
		orderExpr = "l.price DESC"
	case models.SORT_NEWEST:
		orderExpr = "l.created_at DESC"
	case models.SORT_RELEVANCE:
		// Выбираем выражение для сортировки по релевантности в зависимости от типа поиска
		switch searchType {
		case FuzzySearch:
			// Для нечеткого поиска используем similarity
			orderExpr = "similarity(l.title, $1) DESC"
		case FullTextSearch:
			// Для полнотекстового поиска используем ts_rank
			orderExpr = "ts_rank(lsr.title_vector, to_tsquery('russian', $1)) DESC"
		case CombinedSearch:
			// Для комбинированного поиска используем combined_rank
			orderExpr = "combined_rank DESC"
		default:
			orderExpr = "similarity(l.title, $1) DESC"
		}
	}

	return orderExpr
}

// getReverseOrderExpression возвращает обратный порядок сортировки
func getReverseOrderExpression(orderExpr string) string {
	// Специальная обработка для сортировки по релевантности
	if strings.Contains(orderExpr, "similarity") {
		if strings.Contains(orderExpr, "DESC") {
			return strings.Replace(orderExpr, "DESC", "ASC", 1)
		} else if strings.Contains(orderExpr, "ASC") {
			return strings.Replace(orderExpr, "ASC", "DESC", 1)
		}
	}

	// Обработка стандартных выражений сортировки
	if strings.Contains(orderExpr, "ASC") {
		return strings.Replace(orderExpr, "ASC", "DESC", 1)
	} else if strings.Contains(orderExpr, "DESC") {
		return strings.Replace(orderExpr, "DESC", "ASC", 1)
	}

	return orderExpr
}

// getCursorCondition создает SQL условие для пагинации с курсором
func getCursorCondition(orderExpr string, cursor *models.Listing, inclusive bool) string {
	// Специальная обработка для сортировки по релевантности (similarity)
	if strings.Contains(orderExpr, "similarity") {
		// Для сортировки по релевантности нам не нужно дополнительное условие,
		// так как нельзя сравнивать значения функции similarity в условии WHERE
		// Вместо этого используем OFFSET и LIMIT в buildSQLQuery
		return "TRUE" // Возвращаем условие, которое не ограничит выборку
	}

	// Определяем оператор сравнения на основе направления сортировки и включения курсора
	operator := ">" // По умолчанию для ASC и не включая курсор
	if strings.Contains(orderExpr, "DESC") {
		operator = "<" // Для DESC и не включая курсор
	}

	if inclusive {
		operator += "=" // Добавляем = если нужно включить курсор
	}

	// Определяем поле для сравнения на основе выражения сортировки
	var field string
	var value interface{}

	if strings.Contains(orderExpr, "price") {
		field = "l.price"
		value = cursor.Price
	} else if strings.Contains(orderExpr, "created_at") {
		field = "l.created_at"
		value = cursor.CreatedAt
	} else {
		// По умолчанию используем ID для уникальности
		field = "l.id"
		value = cursor.ID

		// Для ID используем оператор сравнения в зависимости от inclusive
		if inclusive {
			return field + " = '" + cursor.ID.String() + "'" // Включаем текущий элемент
		} else {
			return field + " != '" + cursor.ID.String() + "'" // Исключаем текущий элемент
		}
	}

	// Для числовых и временных полей формируем условие сравнения
	if strings.Contains(field, "price") {
		return field + " " + operator + " " + fmt.Sprintf("%f", value.(float64))
	} else if strings.Contains(field, "created_at") {
		return field + " " + operator + " '" + value.(time.Time).Format(time.RFC3339) + "'"
	}

	return ""
}

// determineSearchType определяет оптимальный тип поиска на основе запроса
func determineSearchType(query string) SearchType {
	query = strings.TrimSpace(query)

	// Если запрос пустой, используем нечеткий поиск по умолчанию
	if query == "" {
		return FuzzySearch
	}

	// Короткие запросы (1-2 слова) лучше искать нечетким поиском
	words := strings.Fields(query)
	if len(words) <= 2 {
		return FuzzySearch
	}

	// Запросы с опечатками или специальными символами лучше искать нечетким поиском
	if containsSpecialChars(query) {
		return FuzzySearch
	}

	// Длинные запросы (более 3 слов) лучше искать полнотекстовым поиском
	if len(words) >= 4 {
		return FullTextSearch
	}

	// Для средних запросов (3 слова) используем комбинированный поиск
	return CombinedSearch
}

// containsSpecialChars проверяет, содержит ли запрос специальные символы
func containsSpecialChars(query string) bool {
	specialChars := []string{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "-", "_", "+", "=", "[", "]", "{", "}", "|", "\\", ":", ";", "'", "\"", "<", ">", ",", ".", "?", "/"}

	for _, char := range specialChars {
		if strings.Contains(query, char) {
			return true
		}
	}

	return false
}

// createSearchQuery создает поисковый запрос в зависимости от типа поиска
func createSearchQuery(query string, searchType SearchType) string {
	// Очищаем и нормализуем запрос
	query = strings.TrimSpace(query)

	if query == "" {
		return ""
	}

	// Для нечеткого поиска возвращаем оригинальный запрос
	if searchType == FuzzySearch {
		return query
	}

	// Для полнотекстового и комбинированного поиска создаем tsquery
	return prepareTsQuery(query)
}

// prepareTsQuery подготавливает запрос для полнотекстового поиска
func prepareTsQuery(query string) string {
	// Очищаем запрос от специальных символов, которые могут нарушить синтаксис to_tsquery
	// Заменяем специальные символы на пробелы
	specialChars := []string{"&", "|", "!", "(", ")", ":", "*", "'", "-", "<", ">"}
	for _, char := range specialChars {
		query = strings.ReplaceAll(query, char, " ")
	}

	// Удаляем лишние пробелы
	query = strings.TrimSpace(query)
	query = strings.Join(strings.Fields(query), " ")

	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}

	if len(words) == 1 {
		// Для одного слова используем префиксный поиск
		return words[0] + ":*"
	}

	// Для нескольких слов обрабатываем каждое слово отдельно
	var queryParts []string
	for i, word := range words {
		if i == len(words)-1 {
			// Последнее слово с префиксным поиском
			queryParts = append(queryParts, word+":*")
		} else {
			// Предыдущие слова ищутся полностью
			queryParts = append(queryParts, word)
		}
	}

	// Соединяем слова оператором &
	return strings.Join(queryParts, " & ")
}
