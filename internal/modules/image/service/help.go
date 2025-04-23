package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
)

func (s *Image) getFileName(ctx context.Context, id, size string) string {
	format := "%v-%v.webp"

	return fmt.Sprintf(format, id, size)
}

func (s *Image) GetImageID(ctx context.Context, url string) (string, error) {
	imageFileName, err := s.GetImageFileName(ctx, url)
	if err != nil {
		return "", err
	}

	// Извлекаем UUID из имени файла (UUID-size.webp)
	uuidParts := strings.Split(imageFileName, "-")
	if len(uuidParts) < 2 {
		return "", eris.New("invalid filename format: missing size suffix")
	}

	// Проверка формата UUID
	uuidStr := uuidParts[0]
	if len(uuidStr) != 36 { // Стандартная длина UUID в строковом представлении
		return "", eris.New("invalid UUID length in filename")
	}

	// Возвращаем базовый UUID изображения
	return uuidStr, nil
}

// GetImageFileName извлекает имя файла из URL изображения
func (s *Image) GetImageFileName(ctx context.Context, url string) (string, error) {
	// Валидация URL
	if url == "" {
		return "", eris.New("empty image URL")
	}

	// Проверка на наличие базовых компонентов URL
	if !strings.Contains(url, "/") {
		return "", eris.New("invalid image URL format: no path separator")
	}

	// Извлекаем имя файла из URL
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", eris.New("empty filename in URL")
	}

	// Получаем имя файла из последней части URL
	imageFileName := parts[len(parts)-1]
	if imageFileName == "" {
		return "", eris.New("empty filename in URL")
	}

	// Проверка на наличие расширения файла
	if !strings.Contains(imageFileName, ".") {
		return "", eris.New("invalid filename: no file extension")
	}

	// Возвращаем имя файла
	return imageFileName, nil
}
