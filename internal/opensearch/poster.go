package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/yaroslavvasilenko/argon/internal/entity"
	"log"
	"time"
)

type PosterIndex struct {
	Index
}

func (idx *PosterIndex) GetName() string {
	return idx.Name
}

func (idx *PosterIndex) GetVersion() int64 {
	return int64(idx.Version)
}

func (idx *PosterIndex) GetCreatedAt() *time.Time {
	return idx.CreatedAt
}

func (idx *PosterIndex) SetCreatedAt(t *time.Time) {
	idx.CreatedAt = t
}

func (idx *PosterIndex) GetSettings() string {
	return fmt.Sprintf(`{
        "settings": {
            "analysis": {
                "analyzer": {
                    "standard_analyzer": {
                        "type": "custom",
                        "tokenizer": "standard",
                        "filter": ["lowercase"]
                    }
                }
            }
        },
        "mappings": {
            "properties": {
                "title": {
                    "type": "text",
                    "analyzer": "standard_analyzer"
                }
            },
            "_meta": {
                "version": %d
            }
        }
    }`, idx.Version)

}

func (os *OpenSearch) DeletePoster(ctx context.Context, posterID string) {
	go func() {
		if err := os.deletePoster(ctx, posterID); err != nil {
			log.Printf("Ошибка при удалении постера %v", err)
		}
	}()
}

// DeletePoster Удаляет индекс объявления
func (os *OpenSearch) deletePoster(ctx context.Context, posterID string) error {
	if err := os.client.Delete(ctx, os.PosterIdx.GetName(), posterID); err != nil {
		return errors.Wrapf(err, "error delete poster(%s)", posterID)
	}
	return nil
}

// SearchPosters Поиск чатов
func (os *OpenSearch) SearchPosters(ctx context.Context, query string) ([]entity.PosterSearch, error) {
	reqBody := fmt.Sprintf(`{
    "query": {
        "prefix": {
            "title": %q
        }
    }
}`, query)

	resBody, err := os.client.Search(ctx, os.PosterIdx.GetName(), reqBody)
	if err != nil {
		return nil, errors.Wrapf(err, "error searching chats in index(%s)", os.PosterIdx.GetName())
	}
	defer resBody.Close()

	var res struct {
		Hits struct {
			Hits []struct {
				ID     string              `json:"_id"`
				Source entity.PosterSearch `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err = json.NewDecoder(resBody).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "error unmarshal response")
	}

	posters := make([]entity.PosterSearch, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		posterID, err := uuid.Parse(hit.ID)
		if err != nil {
			//  ToDo: log warn
		}

		hit.Source.ID = posterID // Присваиваем ID из ответа Elasticsearch полю UserID так как этого поля нет в source
		posters = append(posters, hit.Source)
	}

	return posters, nil
}

func (os *OpenSearch) GetChatsIndexCreatedAt(ctx context.Context) (*time.Time, error) {
	idx := os.PosterIdx

	currentCreatedAt, err := os.client.GetIndexCreationDate(ctx, idx.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "could not get index(messages) creation date")
	}

	if (idx.GetCreatedAt() != nil || currentCreatedAt != nil) && !(*idx.GetCreatedAt()).Equal(*currentCreatedAt) {
		return nil, errors.New("index(messages) creation date does not match")
	}

	return idx.GetCreatedAt(), nil
}

func (os *OpenSearch) IndexPosters(ctx context.Context, posters []entity.Poster, expectCodes ...int) {
	go func() {
		if err := os.indexPosters(ctx, entity.NewPosterSearch(posters)); err != nil {
			// Логируем ошибку
			log.Printf("Ошибка при индексировании постера %v", err)
		}

	}()
}

func (os *OpenSearch) indexPosters(ctx context.Context, posters []entity.PosterSearch, expectCodes ...int) error {
	var buf bytes.Buffer

	for _, p := range posters {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": os.PosterIdx.GetName(),
				"_id":    p.ID,
			},
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		msgBytes, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		buf.Write(metaBytes)
		buf.WriteByte('\n')
		buf.Write(msgBytes)
		buf.WriteByte('\n')
	}

	return os.client.RequestBulk(ctx, os.PosterIdx.GetName(), &buf, expectCodes)
}

func (os *OpenSearch) DeleteMessages(ctx context.Context, deleteIDs []string, expectCodes ...int) error {
	var buf bytes.Buffer

	for _, id := range deleteIDs {
		meta := map[string]map[string]string{
			"delete": {
				"_id": id,
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')
	}

	return os.client.RequestBulk(ctx, os.PosterIdx.GetName(), &buf, expectCodes)
}
