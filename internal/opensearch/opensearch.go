package opensearch

import (
	"context"
	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
	"time"
)

type Indexer interface {
	GetName() string
	GetVersion() int64
	GetSettings() string
	SetCreatedAt(t *time.Time)
	GetCreatedAt() *time.Time
}

type OpenSearch struct {
	client  *Client
	ItemIdx Indexer
}

// NewOpenSearch создает новый экземпляр OpenSearch
func NewOpenSearch(addresses []string, login, password, posterIdxName string) (*OpenSearch, error) {
	cfg := opensearch.Config{
		Addresses: addresses,
		Username:  login,
		Password:  password,
		//Logger:    NewSlogLogger(logger),
	}
	osClient, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error creating opensearch client")
	}

	posterIdx := NewItemIndex(posterIdxName)
	return &OpenSearch{
		client:  newClient(osClient),
		ItemIdx: posterIdx,
	}, nil
}

type Index struct {
	Name      string
	Version   int
	CreatedAt *time.Time
}

func (os *OpenSearch) CreateIndex() {

}

func NewItemIndex(name string) Indexer {
	return &ItemIndex{
		Index: Index{
			Name:      name,
			Version:   1,
			CreatedAt: nil,
		},
	}
}

func (os *OpenSearch) createOrUpdateIndex(ctx context.Context, idx Indexer) error {
	index := idx.GetName()

	// получаем информацию
	version, err := os.client.GetIndexVersion(ctx, index)
	if err != nil {
		return errors.Wrapf(err, "error getting index(%s) version", index)
	}

	// создаем индекс, если версия не вернулась
	if version == nil {
		//  если там просто не было версии, иначе при создании будет ошибка
		if err := os.client.DeleteIndex(ctx, index); err != nil {
			return errors.Wrapf(err, "error delete index(%s)", index)
		}

		if err := os.client.CreateIndex(ctx, index, idx.GetSettings()); err != nil {
			return errors.Wrapf(err, "error create index(%s)", index)
		}
	} else if *version != idx.GetVersion() {
		// сверяем версию версии индекса с системой, если есть разница, удаляем индекс

		if err := os.client.DeleteIndex(ctx, index); err != nil {
			return errors.Wrapf(err, "error delete index(%s)", index)
		}

		if err := os.client.CreateIndex(ctx, index, idx.GetSettings()); err != nil {
			return errors.Wrapf(err, "error create index(%s)", index)
		}
	}

	// получаем дату создания индекса
	date, err := os.client.GetIndexCreationDate(ctx, index)
	if err != nil {
		return errors.Wrapf(err, "fail get creation time for index(%s)", index)
	}
	if date == nil {
		return errors.Wrapf(err, "creation time for index(%s) is empty", index)
	}

	idx.SetCreatedAt(date)

	return nil
}
