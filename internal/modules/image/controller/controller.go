package controller

import (
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/modules/image/service"
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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Не удалось получить изображение из формы",
			"details": err.Error(),
		})
	}

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Не удалось открыть файл",
			"details": err.Error(),
		})
	}
	defer file.Close()

	// Читаем первые 512 байт для определения типа файла
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Не удалось прочитать содержимое файла",
			"details": err.Error(),
		})
	}
	file.Seek(0, 0)

	// Определяем MIME-тип
	contentType := http.DetectContentType(buffer)

	// Проверяем, что это изображение
	if !isImageContentType(contentType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Загруженный файл не является изображением",
		})
	}

	// Сохраняем изображение через сервисный слой
	fileURL, err := h.s.SaveImage(c.UserContext(), file, fileHeader.Filename, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Не удалось сохранить изображение",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"url_full": fileURL[0],
		"url":      fileURL[1],
	})
}

func (h *Image) GetImage(c *fiber.Ctx) error {
	req := struct {
		Id string `params:"image_id"`
	}{}

	if err := parser.ParamParser(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	image, err := h.s.GetImage(c.UserContext(), req.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get image",
		})
	}

	// Set appropriate content type header
	// You may want to determine the content type based on the file extension or metadata
	c.Set("Content-Type", "image/webp") // Adjust content type as needed

	// Send the image stream to the client
	return c.SendStream(image)
}
