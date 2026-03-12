package s3

import (
	"bytes"
	"context"
	"io"

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

func (s *S3Storage) Save(ctx context.Context, key string, file io.Reader) error {

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, file)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buf.Bytes())

	_, err = s.client.PutObject(
		ctx,
		s.bucket,
		key,
		reader,
		int64(reader.Len()),
		minio.PutObjectOptions{},
	)

	return err
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadSeeker, error) {

	obj, err := s.client.GetObject(
		ctx,
		s.bucket,
		key,
		minio.GetObjectOptions{},
	)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {

	err := s.client.RemoveObject(
		ctx,
		s.bucket,
		key,
		minio.RemoveObjectOptions{},
	)

	if err != nil {
		return err
	}

	return nil
}
