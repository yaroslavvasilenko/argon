package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/modules/image/service"
	"strings"
)

type Image struct {
	s *service.Image
}

func NewImage(s *service.Image) *Image {
	return &Image{s: s}
}

// isImageContentType проверяет, является ли MIME-тип изображением
func isImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

func (h *Image) UploadImage(c *fiber.Ctx) error {
	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не удалось получить изображение из формы",
			"details": err.Error(),
		})
	}

	// Получаем MIME-тип файла
	contentType := file.Header.Get("Content-Type")
	
	// Проверяем, что это изображение
	if !isImageContentType(contentType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Загруженный файл не является изображением",
		})
	}

	// Сохраняем изображение через сервисный слой
	fileURL, err := h.s.SaveImage(c.UserContext(), file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при сохранении изображения",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"filename": file.Filename,
		"size": file.Size,
		"type": contentType,
		"url": fileURL,
	})
}
