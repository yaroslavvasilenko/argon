package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

func (s *Image) DeleteImageSync(stopChan chan struct{}) {
    // Добавляем обработку паники для всей горутины
    defer func() {
        if r := recover(); r != nil {
            s.log.Errorf("Panic in DeleteImageSync: %v", r)
        }
    }()
    
    s.log.Infof("Starting image cleanup cron task")
    
    for {
        select {
        case <-stopChan:
            s.log.Infof("Image cleanup cron task received stop signal")
            return
            
        default:
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        s.log.Errorf("Panic in image cleanup task: %v", r)
                    }
                }()
                
                ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
                defer cancel()
                
                count, err := s.DeleteImage(ctx)
                if err != nil {
                    s.log.Errorf("Failed to delete unused images: %v", err)
                } else {
                    s.log.Infof("Deleted images: %s", count)
                }
            }()
         
            select {
            case <-stopChan:
                s.log.Infof("Image cleanup cron task received stop signal during sleep")
                return
            case <-time.After(100 * time.Second):
            }
        }
    }
}


// DeleteImage удаляет изображения, которые не связаны с объявлениями более часа
func (s *Image) DeleteImage(ctx context.Context) (string, error) {

	// Получаем текущее время
	now := time.Now()

	// Получаем список несвязанных изображений из базы данных
	unlinkedImages, err := s.s.GetUnlinkedImages(ctx, now.Add(-1 * time.Hour))
	if err != nil {
		return "", eris.Wrapf(err, "Failed to get unlinked images: %v", err)
	}

	var deletedCount int
	var errors []string

	// Удаляем каждое несвязанное изображение
	for _, imageName := range unlinkedImages {
		err := s.s.DeleteFile(ctx, imageName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error deleting image %s: %v", imageName, err))
			continue
		}

		deletedCount++
	}

	// Формируем отчет о выполнении
	result := fmt.Sprintf("Deleted %d unused images", deletedCount)
	if len(errors) > 0 {
		result += fmt.Sprintf(". Errors encountered: %d", len(errors))
		return result, eris.New(strings.Join(errors, "; "))
	}

	return result, nil
}


