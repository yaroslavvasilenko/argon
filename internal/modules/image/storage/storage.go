package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rotisserie/eris"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
)

// Minio представляет клиент для работы с хранилищем MinIO
type Minio struct {
	client     *minio.Client
	bucketName string
}

// NewMinio создает новый клиент MinIO
func NewMinio(ctx context.Context, cfg config.Config) (*Minio, error) {
	// Инициализация клиента MinIO
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.User, cfg.Minio.Password, ""),
		Secure: false, // Используем HTTP вместо HTTPS
	})
	if err != nil {
		return nil, eris.Wrapf(err, "creating client for %s failed", cfg.Minio.Endpoint)
	}

	// Проверяем существование бакета, если нет - создаем
	exists, err := client.BucketExists(ctx, cfg.Minio.Bucket)
	if err != nil {
		return nil, eris.Wrapf(err, "checking bucket %s failed", cfg.Minio.Bucket)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, eris.Wrapf(err, "creating bucket %s failed", cfg.Minio.Bucket)
		}
		slog.Info("Bucket created", "bucket", cfg.Minio.Bucket)
	}

	return &Minio{
		client:     client,
		bucketName: cfg.Minio.Bucket,
	}, nil
}

type Image struct {
	gorm  *gorm.DB
	pool  *pgxpool.Pool
	minio *Minio
}

func NewImage(db *gorm.DB, pool *pgxpool.Pool, minio *Minio) *Image {
	return &Image{gorm: db, pool: pool, minio: minio}
}

// LinkImageToListing создает запись о связи изображения с объявлением
func (m *Image) LinkImageToListing(ctx context.Context, imageName string, listingID string) error {
	// Преобразуем строковый ID в UUID
	listingUUID, err := uuid.Parse(listingID)
	if err != nil {
		return eris.Wrapf(err, "invalid listing ID: %s", listingID)
	}

	// Создаем запись о связи
	imageLink := models.ImageLink{
		ListingID: listingUUID,
		NameImage:  imageName,
		Linked:    true,
		UpdatedAt: time.Now(),
	}

	// Сохраняем в базу данных
	result := m.gorm.WithContext(ctx).Create(&imageLink)
	if result.Error != nil {
		return eris.Wrapf(result.Error, "failed to create image link for %s", imageName)
	}

	return nil
}

// UnlinkImageFromListing помечает изображение как не связанное с объявлением
func (m *Image) UnlinkImageFromListing(ctx context.Context, imageName string, listingID string) error {
	// Преобразуем строковый ID в UUID
	listingUUID, err := uuid.Parse(listingID)
	if err != nil {
		return eris.Wrapf(err, "invalid listing ID: %s", listingID)
	}

	// Обновляем запись о связи
	result := m.gorm.WithContext(ctx).Model(&models.ImageLink{}).Where(
		"listing_id = ? AND name_image = ?", listingUUID, imageName,
	).Updates(map[string]interface{}{
		"linked":     false,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return eris.Wrapf(result.Error, "failed to unlink image %s", imageName)
	}

	return nil
}

// GetUnlinkedImages возвращает список изображений, которые не связаны с объявлениями и старше указанного времени
func (m *Image) GetUnlinkedImages(ctx context.Context, olderThan time.Time) ([]string, error) {
	var imageLinks []models.ImageLink
	result := m.gorm.WithContext(ctx).Model(&models.ImageLink{}).Where(
		"linked = false AND updated_at < ?", olderThan,
	).Select("name_image").Find(&imageLinks)

	if result.Error != nil {
		return nil, eris.Wrapf(result.Error, "failed to get unlinked images")
	}

	// Преобразуем результат в список имен файлов
	imageNames := make([]string, len(imageLinks))
	for i, link := range imageLinks {
		imageNames[i] = link.NameImage
	}

	return imageNames, nil
}
