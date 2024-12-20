package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client     *minio.Client
	bucketName string
}

type IClient interface {
	GetFile(ctx context.Context, filename string) (*minio.Object, error)
	PutFile(ctx context.Context, filename string, reader io.Reader) error
	DeleteFile(ctx context.Context, filename string) error
}

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Bucket   string `yaml:"bucket"`
	Location string `yaml:"location"`
}

func NewClient(cfg Config) (IClient, error) {
	client, err := minio.New(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.User, cfg.Password, ""),
		Secure: false})
	if err != nil {
		return nil, err
	}

	exists, _ := client.BucketExists(context.Background(), cfg.Bucket)
	if exists {
		return nil, err
	}

	err = client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
		Region:        cfg.Location,
		ObjectLocking: false})

	return &Client{
		client:     client,
		bucketName: cfg.Bucket,
	}, nil
}

func (m *Client) GetFile(ctx context.Context, filename string) (*minio.Object, error) {
	file, err := m.client.GetObject(ctx, m.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (m *Client) PutFile(ctx context.Context, filename string, reader io.Reader) error {
	_, err := m.client.PutObject(ctx, m.bucketName, filename, reader, -1, minio.PutObjectOptions{ContentType: "application/pdf"})
	return err
}

func (m *Client) DeleteFile(ctx context.Context, filename string) error {
	return m.client.RemoveObject(ctx, m.bucketName, filename, minio.RemoveObjectOptions{})
}
