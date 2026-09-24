package s3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequiresBucket(t *testing.T) {
	_, err := New(context.Background(), Config{Region: "eu-central-1"})
	assert.Error(t, err)
}

func TestNewPublicBaseURL(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "aws default (no endpoint)",
			cfg:  Config{Bucket: "my-bucket", Region: "eu-central-1"},
			want: "https://my-bucket.s3.eu-central-1.amazonaws.com",
		},
		{
			name: "custom endpoint, virtual-hosted style",
			cfg:  Config{Bucket: "my-bucket", Region: "eu", Endpoint: "https://s3.example.com"},
			want: "https://my-bucket.s3.example.com",
		},
		{
			name: "custom endpoint, forced path style (MinIO/Hetzner style)",
			cfg:  Config{Bucket: "my-bucket", Region: "eu", Endpoint: "https://minio.internal:9000", ForcePathStyle: true},
			want: "https://minio.internal:9000/my-bucket",
		},
		{
			name: "explicit PublicBaseURL wins over derivation (e.g. a CDN)",
			cfg:  Config{Bucket: "my-bucket", Region: "eu", Endpoint: "https://minio.internal:9000", PublicBaseURL: "https://cdn.example.com/"},
			want: "https://cdn.example.com",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(context.Background(), tc.cfg)
			require.NoError(t, err)
			assert.Equal(t, tc.want, s.publicBaseURL)
		})
	}
}
