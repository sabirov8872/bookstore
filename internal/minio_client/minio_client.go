package minio_client

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
	location   string
}

func NewMinioClient(endpoint, accessKeyID, secretAccessKey, bucketName, location string) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
		location:   location,
	}, nil
}

func (m *MinioClient) CreateBucket() {
	err := m.client.MakeBucket(context.Background(), m.bucketName, minio.MakeBucketOptions{
		Region:        m.location,
		ObjectLocking: false,
	})
	if err != nil {
		if exists, _ := m.client.BucketExists(context.Background(), m.bucketName); exists {
			fmt.Println("Bucket already exists.")
		} else {
			log.Fatalln(err)
		}
	} else {
		fmt.Println("Bucket created successfully.")
	}
}

func (m *MinioClient) GetObjectFromMinio(ctx context.Context, objectName string) (*minio.Object, error) {
	fmt.Println(objectName)
	object, err := m.client.GetObject(
		ctx,
		m.bucketName,
		objectName,
		minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (m *MinioClient) PutObjectToTheMinio(ctx context.Context, objectName string, reader io.Reader) error {
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream"})
	return err
}

func (m *MinioClient) DeleteObject(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioClient) ListObjectsInBucket(bucketName string) ([]string, error) {
	var objects []string
	objectCh := m.client.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		objects = append(objects, object.Key)
	}

	return objects, nil
}

func (m *MinioClient) DeleteBucket() error {
	return m.client.RemoveBucket(context.Background(), m.bucketName)
}
