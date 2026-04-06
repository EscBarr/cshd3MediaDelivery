package s3

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

func New(endpoint, accessKey, secretKey, bucket string) (*S3Storage, error) {

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &S3Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Storage) buildKey(key string, path *string) string {
	if path == nil || *path == "" {
		return key
	}
	return strings.TrimSuffix(*path, "/") + "/" + key
}

func (s *S3Storage) Save(ctx context.Context, key string, path *string, file io.Reader) (string, error) {

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	reader := bytes.NewReader(buf.Bytes())
	finalKey := s.buildKey(key, path)

	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		finalKey,
		reader,
		int64(reader.Len()),
		minio.PutObjectOptions{},
	)

	return finalKey, err
}

func (s *S3Storage) Get(ctx context.Context, key string, path *string) (io.ReadSeeker, error) {

	finalKey := s.buildKey(key, path)

	obj, err := s.client.GetObject(
		ctx,
		s.bucket,
		finalKey,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, obj); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string, path *string) error {

	finalKey := s.buildKey(key, path)

	return s.client.RemoveObject(
		ctx,
		s.bucket,
		finalKey,
		minio.RemoveObjectOptions{},
	)
}
