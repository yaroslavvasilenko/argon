package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/image/storage"
)

type Image struct {
	s      *storage.Image
	logger *logger.Glog
}

func NewImage(s *storage.Image, pool *pgxpool.Pool, logger *logger.Glog) *Image {
	srv := &Image{
		s:      s,
		logger: logger,
	}

	return srv
}

// SaveImage сохраняет изображение в MinIO и возвращает имя файла
func (s *Image) SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка при открытии файла: %w", err)
	}
	defer src.Close()

	// Генерируем уникальное имя файла
	extension := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)

	if !isImageContentType(file.Header.Get("Content-Type")) {
		return "", fmt.Errorf("ошибка при загрузке изображения: %w", err)
	}

	// Загружаем файл в MinIO
	objectName, err = s.s.UploadImage(ctx, objectName, file.Header.Get("Content-Type"), src)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке изображения: %w", err)
	}

	// Получаем URL для доступа к файлу
	fileURL, err := s.s.GetImageURL(ctx, objectName)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL изображения: %w", err)
	}

	return fileURL, nil
}

// isImageContentType проверяет, является ли MIME-тип изображением
func isImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}
