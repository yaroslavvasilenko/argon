package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"log"
	"time"
)

type ItemIndex struct {
	Index
}

func (idx *ItemIndex) GetName() string {
	return idx.Name
}

func (idx *ItemIndex) GetVersion() int64 {
	return int64(idx.Version)
}

func (idx *ItemIndex) GetCreatedAt() *time.Time {
	return idx.CreatedAt
}

func (idx *ItemIndex) SetCreatedAt(t *time.Time) {
	idx.CreatedAt = t
}

func (idx *ItemIndex) GetSettings() string {
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

func (os *OpenSearch) DeleteItem(ctx context.Context, posterID string) {
	go func() {
		if err := os.deletePoster(ctx, posterID); err != nil {
			log.Printf("error delete item in index %v", err)
		}
	}()
}

// DeleteItem Удаляет индекс объявления
func (os *OpenSearch) deletePoster(ctx context.Context, itemID string) error {
	if err := os.client.Delete(ctx, os.ItemIdx.GetName(), itemID); err != nil {
		return errors.Wrapf(err, "error delete index(%s)", itemID)
	}
	return nil
}

// SearchItems Поиск объявлений
func (os *OpenSearch) SearchItems(ctx context.Context, query string) ([]models.ItemSearch, error) {
	reqBody := fmt.Sprintf(`{
    "query": {
        "prefix": {
            "title": %q
        }
    }
}`, query)

	resBody, err := os.client.Search(ctx, os.ItemIdx.GetName(), reqBody)
	if err != nil {
		return nil, errors.Wrapf(err, "error searching items in index(%s)", os.ItemIdx.GetName())
	}
	defer resBody.Close()

	var res struct {
		Hits struct {
			Hits []struct {
				ID     string            `json:"_id"`
				Source models.ItemSearch `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err = json.NewDecoder(resBody).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "error unmarshal response")
	}

	items := make([]models.ItemSearch, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		itemID, err := uuid.Parse(hit.ID)
		if err != nil {
			//  ToDo: log warn
		}

		hit.Source.ID = itemID // Присваиваем ID из ответа Elasticsearch полю UserID так как этого поля нет в source
		items = append(items, hit.Source)
	}

	return items, nil
}

func (os *OpenSearch) GetChatsIndexCreatedAt(ctx context.Context) (*time.Time, error) {
	idx := os.ItemIdx

	currentCreatedAt, err := os.client.GetIndexCreationDate(ctx, idx.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "could not get index(item) creation date")
	}

	if (idx.GetCreatedAt() != nil || currentCreatedAt != nil) && !(*idx.GetCreatedAt()).Equal(*currentCreatedAt) {
		return nil, errors.New("index(item) creation date does not match")
	}

	return idx.GetCreatedAt(), nil
}

func (os *OpenSearch) IndexItems(ctx context.Context, items []models.Item, expectCodes ...int) {
	go func() {
		if err := os.indexItems(ctx, models.NewItemSearch(items)); err != nil {
			// Логируем ошибку
			log.Printf("error index item %v", err)
		}
	}()
}

func (os *OpenSearch) indexItems(ctx context.Context, items []models.ItemSearch, expectCodes ...int) error {
	var buf bytes.Buffer

	for _, p := range items {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": os.ItemIdx.GetName(),
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

	return os.client.RequestBulk(ctx, os.ItemIdx.GetName(), &buf, expectCodes)
}

func (os *OpenSearch) DeleteItems(ctx context.Context, deleteIDs []string, expectCodes ...int) error {
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

	return os.client.RequestBulk(ctx, os.ItemIdx.GetName(), &buf, expectCodes)
}
