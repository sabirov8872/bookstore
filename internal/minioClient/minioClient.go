package minioClient

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(endpoint, accessKeyID, secretAccessKey, bucketName string) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false})
	if err != nil {
		return nil, err
	}
	return &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

func (m *MinioClient) CreateBucket(location string) error {
	exists, _ := m.client.BucketExists(context.Background(), m.bucketName)
	if exists {
		return nil
	}

	return m.client.MakeBucket(context.Background(), m.bucketName, minio.MakeBucketOptions{
		Region:        location,
		ObjectLocking: false})
}

func (m *MinioClient) GetBookFile(ctx context.Context, bookFileName string) (*minio.Object, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, bookFileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (m *MinioClient) PutBookFile(ctx context.Context, bookFileName string, reader io.Reader) error {
	_, err := m.client.PutObject(ctx, m.bucketName, bookFileName, reader, -1,
		minio.PutObjectOptions{ContentType: "application/pdf"})
	return err
}

func (m *MinioClient) DeleteBookFile(ctx context.Context, bookFileName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, bookFileName, minio.RemoveObjectOptions{})
}
