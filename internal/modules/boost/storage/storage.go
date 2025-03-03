package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
)

type Boost struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

func NewBoost(db *gorm.DB, pool *pgxpool.Pool) *Boost {
	return &Boost{gorm: db, pool: pool}
}

func (s *Boost) GetBoosts(ctx context.Context, id uuid.UUID) ([]models.Boost, error) {
	query := `
		SELECT listing_id, boost_type, commission
		FROM listing_boosts
		WHERE listing_id = $1
	`

	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boosts []models.Boost
	for rows.Next() {
		var boost models.Boost
		var boostTypeStr string

		if err := rows.Scan(&boost.ListingID, &boostTypeStr, &boost.Commission); err != nil {
			return nil, err
		}

		boost.Type = models.BoostType(boostTypeStr)
		boosts = append(boosts, boost)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(boosts) == 0 {
		return []models.Boost{}, nil
	}

	return boosts, nil
}

// UpsertBoost обновляет или добавляет бусты для объявления и удаляет те, которые не были переданы
func (s *Boost) UpsertBoost(ctx context.Context, listingID uuid.UUID, boosts []models.Boost) error {
	// Получаем комиссии для типов бустов
	commissions := models.GetBoostTypesWithCommissions()

	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Используем отложенную функцию для отката транзакции в случае ошибки
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Удаляем все существующие бусты для данного объявления
	_, err = tx.Exec(ctx, "DELETE FROM listing_boosts WHERE listing_id = $1", listingID)
	if err != nil {
		return err
	}

	// Если передан пустой список бустов, просто завершаем транзакцию
	if len(boosts) == 0 {
		return tx.Commit(ctx)
	}

	// Вместо использования batch, будем выполнять запросы напрямую
	insertQuery := `
		INSERT INTO listing_boosts (listing_id, boost_type, commission)
		VALUES ($1, $2, $3)
	`

	// Добавляем каждый буст отдельным запросом
	for _, boost := range boosts {
		// Убеждаемся, что ID объявления совпадает с переданным
		boost.ListingID = listingID

		// Получаем комиссию для данного типа буста
		commission, exists := commissions[boost.Type]
		if !exists {
			// Если комиссия не найдена, используем значение по умолчанию или пропускаем
			continue
		}

		// Выполняем запрос на вставку
		_, err = tx.Exec(ctx, insertQuery, boost.ListingID, boost.Type, commission)
		if err != nil {
			return err
		}
	}

	// Завершаем транзакцию
	return tx.Commit(ctx)
}
