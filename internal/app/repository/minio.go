package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func generateAspirinStageFileName(
	prefix string,
	originalName string,
) (string, error) {

	randomBytes := make([]byte, 4)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	randomPart := hex.EncodeToString(randomBytes)

	extension := strings.ToLower(
		filepath.Ext(originalName),
	)

	return fmt.Sprintf(
		"%s_%d_%s%s",
		prefix,
		time.Now().UnixMilli(),
		randomPart,
		extension,
	), nil
}

func (r *Repository) UploadAspirinStageFile(
	header *multipart.FileHeader,
	prefix string,
	contentType string,
) (string, error) {

	filename, err :=
		generateAspirinStageFileName(
			prefix,
			header.Filename,
		)

	if err != nil {
		return "", fmt.Errorf(
			"ошибка генерации имени файла: %w",
			err,
		)
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf(
			"ошибка открытия файла: %w",
			err,
		)
	}

	defer file.Close()

	_, err = r.minioClient.PutObject(
		context.Background(),
		r.minioBucketName,
		filename,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		return "", fmt.Errorf(
			"ошибка загрузки файла в MinIO: %w",
			err,
		)
	}

	return filename, nil
}

func (r *Repository) DeleteAspirinStageFile(
	filename string,
) error {

	if filename == "" {
		return nil
	}

	return r.minioClient.RemoveObject(
		context.Background(),
		r.minioBucketName,
		filename,
		minio.RemoveObjectOptions{},
	)
}

// В БД хранится имя файла.
// Наружу возвращаем полноценный URL MinIO.
// Старые http/https и локальные /static пути не ломаем.
func (r *Repository) AspirinStageMediaURL(
	storedValue *string,
) *string {

	if storedValue == nil ||
		strings.TrimSpace(*storedValue) == "" {
		return nil
	}

	value := strings.TrimSpace(*storedValue)

	if strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") ||
		strings.HasPrefix(value, "/") {

		return &value
	}

	url := fmt.Sprintf(
		"http://%s/%s/%s",
		r.minioEndpoint,
		r.minioBucketName,
		value,
	)

	return &url
}