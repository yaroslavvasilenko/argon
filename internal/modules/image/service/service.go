package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/config"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/image/storage"
)

type Image struct {
	s        *storage.Image
	log      *logger.Glog
	stopChan chan struct{} // Канал для остановки cron-задачи
}

func NewImage(s *storage.Image, logger *logger.Glog) *Image {
	srv := &Image{
		s:        s,
		log:      logger,
		stopChan: make(chan struct{}),
	}

	return srv
}

// SaveImage сохраняет изображение в MinIO и возвращает имя файла
func (s *Image) SaveImage(ctx context.Context, file multipart.File, name, contentType string) ([]string, error) {
	const maxFileSize = 10 * 1024 * 1024 // 10 МБ в байтах

	// Читаем файл в память
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, eris.Wrap(err, "failed to read file")
	}

	// Проверяем размер файла
	if len(fileBytes) > maxFileSize {
		return nil, fiber.NewError(fiber.StatusBadRequest, "size of file is too large (10MB)")
	}

	// Создаем копию для изображения 400px
	im400, err := vips.NewImageFromBuffer(fileBytes)
	if err != nil {
		return nil, eris.Wrap(err, "failed to load image for 400px")
	}
	defer im400.Close()

	// Обрезаем изображение до нужных пропорций
	if err := cropImage(im400); err != nil {
		return nil, eris.Wrap(err, "failed to crop image for 400px")
	}

	// Изменяем размер до 400px
	if err := resizeImage(im400, 400); err != nil {
		return nil, eris.Wrap(err, "failed to resize image to 400px")
	}

	// Создаем копию для изображения 200px
	im200, err := vips.NewImageFromBuffer(fileBytes)
	if err != nil {
		return nil, eris.Wrap(err, "failed to load image for 200px")
	}
	defer im200.Close()

	// Обрезаем изображение до нужных пропорций
	if err := cropImage(im200); err != nil {
		return nil, eris.Wrap(err, "failed to crop image for 200px")
	}

	// Изменяем размер до 200px
	if err := resizeImage(im200, 200); err != nil {
		return nil, eris.Wrap(err, "failed to resize image to 200px")
	}

	// Экспортируем и загружаем изображение 400px
	webpBytes400, err := exportToWebP(im400, true) // Используем сжатие с потерями для 400px
	if err != nil {
		return nil, eris.Wrap(err, "failed to export 400px image to WebP")
	}

	// Upload image MinIO 400px
	imageName400px, err := s.s.UploadImage(ctx, s.getFileName(ctx, uuid.New().String(), "400px"), bytes.NewReader(webpBytes400))
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload 400px image")
	}

	// Экспортируем и загружаем изображение 200px
	webpBytes200, err := exportToWebP(im200, true) // Используем сжатие с потерями для 200px
	if err != nil {
		return nil, eris.Wrap(err, "failed to export 200px image to WebP")
	}

	// Upload image MinIO 200px
	imageName200px, err := s.s.UploadImage(ctx, s.getFileName(ctx, uuid.New().String(), "200px"), bytes.NewReader(webpBytes200))
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload 200px image")
	}

	return []string{createUrlForImage(imageName400px), createUrlForImage(imageName200px)}, nil
}

// cropImage обрезает изображение до нужных пропорций
func cropImage(img *vips.ImageRef) error {
	// Проверяем размеры изображения
	width := img.Width()
	height := img.Height()

	// Определяем соотношение сторон
	ratio := float64(width) / float64(height)

	// Если изображение слишком широкое (соотношение > 3:1), обрезаем его до пропорции 3:1
	if ratio > 3.0 {
		// Вычисляем новую ширину для соотношения 3:1
		newWidth := int(float64(height) * 3.0)
		// Вычисляем, сколько нужно отрезать с каждой стороны
		cropAmount := (width - newWidth) / 2
		// Обрезаем изображение с обеих сторон
		err := img.ExtractArea(cropAmount, 0, newWidth, height)
		if err != nil {
			return eris.Wrapf(err, "cropping wide image failed")
		}
	}

	// Если изображение слишком высокое (соотношение < 1:3), обрезаем его до пропорции 1:3
	if ratio < 1.0/3.0 {
		// Вычисляем новую высоту для соотношения 1:3
		newHeight := int(float64(width) * 3.0)
		// Вычисляем, сколько нужно отрезать с каждой стороны
		cropAmount := (height - newHeight) / 2
		// Обрезаем изображение сверху и снизу
		err := img.ExtractArea(0, cropAmount, width, newHeight)
		if err != nil {
			return eris.Wrapf(err, "cropping tall image failed")
		}
	}

	return nil
}

// resizeImage изменяет размер изображения
func resizeImage(img *vips.ImageRef, scale int) error {
	// Получаем текущие размеры изображения
	width := img.Width()
	height := img.Height()

	// Проверяем, превышает ли высота или ширина заданный масштаб
	if height > scale || width > scale {
		// Вычисляем коэффициенты масштабирования
		scaleHeight := float64(scale) / float64(height)
		scaleWidth := float64(scale) / float64(width)

		// Используем наименьший коэффициент для сохранения пропорций
		scaleComputed := scaleHeight
		if scaleWidth < scaleHeight {
			scaleComputed = scaleWidth
		}

		// Применяем масштабирование
		err := img.Resize(scaleComputed, vips.KernelLanczos3)
		if err != nil {
			return eris.Wrapf(err, "resizing image to %vpx failed", scale)
		}
	}
	return nil
}

// exportToWebP экспортирует изображение в формат WebP
func exportToWebP(img *vips.ImageRef, lossless bool) ([]byte, error) {
	// Параметры экспорта в WebP
	exportParams := vips.NewWebpExportParams()
	exportParams.Quality = 85        // Хорошее качество
	exportParams.Lossless = lossless // Сжатие с потерями по умолчанию
	exportParams.ReductionEffort = 4 // Средний уровень сжатия (0-6)

	// Экспортируем в WebP
	webpBytes, _, err := img.ExportWebp(exportParams)
	if err != nil {
		return nil, eris.Wrapf(err, "exporting image to WebP failed")
	}

	return webpBytes, nil
}

func createUrlForImage(imageName string) string {
	return fmt.Sprintf("%v/api/v1/images/get/%v", config.GetConfig().App.ServerUrl, imageName) // fmt.Sprintf("%v/api/image/%v", config.GetConfig().App.ServerUrl, imageName)
}

func (s *Image) GetImage(ctx context.Context, id string) (io.ReadCloser, error) {
	return s.s.GetFile(ctx, id)
}
