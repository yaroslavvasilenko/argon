package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/rotisserie/eris"
)

// UploadImage загружает изображение в MinIO
func (m *Image) UploadImage(ctx context.Context, fileName string, file io.Reader) (string, error) {
	// Загружаем файл в MinIO
	_, err := m.minio.client.PutObject(ctx, m.minio.bucketName, fileName, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", eris.Wrapf(err, "uploading file %s failed", fileName)
	}

	return fileName, nil
}

// DeleteFile удаляет файл из MinIO
func (m *Image) DeleteFile(ctx context.Context, objectName string) error {
	err := m.minio.client.RemoveObject(ctx, m.minio.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return eris.Wrapf(err, "object %s not found, in bucket %s", objectName, m.minio.bucketName)
	}

	return nil
}

// GetFile получает файл из MinIO
func (m *Image) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := m.minio.client.GetObject(ctx, m.minio.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, eris.Wrapf(err, "object %s not found, in bucket %s", objectName, m.minio.bucketName)
	}

	return obj, nil
}
