package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yaroslavvasilenko/argon/config"
	"gorm.io/gorm"
)

// Minio представляет клиент для работы с хранилищем MinIO
type Minio struct {
	client     *minio.Client
	bucketName string
}

// NewMinio создает новый клиент MinIO
func NewMinio(ctx context.Context, cfg *config.Config) (*Minio, error) {
	// Инициализация клиента MinIO
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.User, cfg.Minio.Password, ""),
		Secure: false, // Используем HTTP вместо HTTPS
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании клиента MinIO: %w", err)
	}

	// Проверяем существование бакета, если нет - создаем
	exists, err := client.BucketExists(ctx, cfg.Minio.Bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке существования бакета: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка при создании бакета: %w", err)
		}
		log.Printf("Создан новый бакет: %s\n", cfg.Minio.Bucket)
	}

	return &Minio{
		client:     client,
		bucketName: cfg.Minio.Bucket,
	}, nil
}

// UploadImage загружает изображение в MinIO
func (m *Image) UploadImage(ctx context.Context, fileName, contentType string, file io.Reader) (string, error) {
	// Загружаем файл в MinIO
	_, err := m.minio.client.PutObject(ctx, m.minio.bucketName, fileName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке файла в MinIO: %w", err)
	}

	return fileName, nil
}

// GetFileURL возвращает URL для доступа к файлу
func (m *Image) GetImageURL(ctx context.Context, objectName string) (string, error) {
	// Проверяем существование объекта
	_, err := m.minio.client.StatObject(ctx, m.minio.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("ошибка при проверке существования объекта: %w", err)
	}

	// Получаем URL для доступа к объекту
	url, err := m.minio.client.PresignedGetObject(ctx, m.minio.bucketName, objectName, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL: %w", err)
	}

	return url.String(), nil
}

// DeleteFile удаляет файл из MinIO
func (m *Image) DeleteFile(ctx context.Context, objectName string) error {
	err := m.minio.client.RemoveObject(ctx, m.minio.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка при удалении файла: %w", err)
	}

	return nil
}

// GetFile получает файл из MinIO
func (m *Image) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := m.minio.client.GetObject(ctx, m.minio.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении файла: %w", err)
	}

	return obj, nil
}

type Image struct {
	gorm  *gorm.DB
	pool  *pgxpool.Pool
	minio *Minio
}

func NewImage(db *gorm.DB, pool *pgxpool.Pool, minio *Minio) *Image {
	return &Image{gorm: db, pool: pool, minio: minio}
}
