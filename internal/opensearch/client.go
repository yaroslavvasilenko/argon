package opensearch

import (
	"bytes"
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Client структура для клиента OpenSearch
type Client struct {
	transport *opensearch.Client
}

// NewClient создает новый экземпляр клиента для работы с OpenSearch
func newClient(transport *opensearch.Client) *Client {
	return &Client{transport: transport}
}

// DeleteIndex удаляет индекс
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	req := opensearchapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return errors.Wrap(err, "request execution error")
	}
	defer res.Body.Close()

	return checkResponse(res, []int{404})
}

// CreateIndex создает индекс
func (c *Client) CreateIndex(ctx context.Context, index, settings string) error {
	req := opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(settings),
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return errors.Wrapf(err, "request execution error")
	}
	defer res.Body.Close()

	return checkResponse(res, nil)
}

// GetIndexVersion Возвращает версию индекса. nil, nil - индекс не найден
func (c *Client) GetIndexVersion(ctx context.Context, index string) (*int64, error) {
	raw, err := c.getIndexRawData(ctx, index)
	if err != nil {
		return nil, errors.Wrap(err, "fail get index info data")
	}

	version := gjson.GetBytes(raw, fmt.Sprintf("%s.mappings._meta.version", index)).Int()

	if version == 0 {
		return nil, nil
	}

	return &version, nil
}

// GetIndexCreationDate возвращает дату создания индекса. nil, nil - индекс не найден
func (c *Client) GetIndexCreationDate(ctx context.Context, index string) (*time.Time, error) {
	raw, err := c.getIndexRawData(ctx, index)
	if err != nil {
		return nil, errors.Wrap(err, "fail get index info data")
	}

	dateStr := gjson.GetBytes(raw, fmt.Sprintf("%s.settings.index.creation_date", index)).String()

	if dateStr == "" {
		return nil, errors.New("not found creation_date value")
	}

	i, err := strconv.ParseInt(dateStr, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing date: %s", dateStr)
	}

	date := time.UnixMilli(i)

	return &date, nil
}

// GetIndexRawData возвращает сырые данные индекса в виде байтового массива.
// nil, nil - индекс не найден.
func (c *Client) getIndexRawData(ctx context.Context, index string) ([]byte, error) {
	req := opensearchapi.IndicesGetRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return nil, errors.Wrapf(err, "request execution error")
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := checkResponse(res, nil); err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading response body")
	}

	return raw, nil
}

// Index индексирует документ
func (c *Client) Index(ctx context.Context, index string, id string, body []byte) error {
	req := opensearchapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return errors.Wrapf(err, "request execution error")
	}
	defer res.Body.Close()

	return checkResponse(res, nil)
}

// RequestBulk выполняет пакетную индексацию документов
func (c *Client) RequestBulk(ctx context.Context, index string, buf *bytes.Buffer, expectCodes []int) error {
	req := opensearchapi.BulkRequest{
		Index:   index,
		Body:    buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return errors.Wrap(err, "bulk index operation error")
	}
	defer res.Body.Close()

	return checkResponse(res, expectCodes)
}

// Delete удаляет документ из индекса
func (c *Client) Delete(ctx context.Context, index string, id string) error {
	req := opensearchapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return errors.Wrapf(err, "request execution error")
	}
	defer res.Body.Close()

	return checkResponse(res, []int{404})
}

// Search выполняет поиск документов
func (c *Client) Search(ctx context.Context, index string, body string) (io.ReadCloser, error) {
	req := opensearchapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(body),
	}

	res, err := req.Do(ctx, c.transport)
	if err != nil {
		return nil, errors.Wrap(err, "error getting response")
	}

	// Проверяем ответ, но не закрываем тело
	if err := checkResponse(res, []int{}); err != nil {
		defer res.Body.Close() // Закрываем тело в случае ошибки
		return nil, err
	}

	return res.Body, nil // Передаем ответственность за закрытие вызывающему коду
}

// Проверяет запрос на предмет успешного кода ответа
func checkResponse(res *opensearchapi.Response, expectCodes []int) error {
	if res.IsError() && !slices.Contains(expectCodes, res.StatusCode) {
		bodyBytes, _ := io.ReadAll(res.Body)
		bodyRaw := string(bodyBytes)

		return fmt.Errorf("response code error: %s %s %s",
			gjson.Get(bodyRaw, "error.type").String(),
			gjson.Get(bodyRaw, "error.reason").String(),
			res.Status(),
		)
	}
	return nil
}
