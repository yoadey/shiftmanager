// Package s3 implements port.MediaStorage against an S3-compatible object
// store (AWS S3, MinIO, Hetzner Object Storage, …) — T-013. It is a drop-in
// alternative to internal/adapter/storage/local for deployments that can't
// mount a ReadWriteMany PersistentVolumeClaim for uploaded media.
//
// Uploaded objects are served at the same "public, unguessable URL, no auth
// check" model as the local adapter's /uploads/* route (event attachments
// and the club logo are meant to be viewable by anyone who has the link,
// including e.g. in an email). This package deliberately does not set an
// object ACL on Put (many buckets have ACLs disabled entirely under the
// "bucket owner enforced" object-ownership setting, which makes any
// ACL-bearing PutObject call fail outright) — instead, the **bucket itself
// must be configured for public read** (a bucket policy, or a public/CDN
// origin in front of it) before switching MEDIA_STORAGE to s3. See
// docs/openspec/08-technik.md (T-013) and .env.example.
package s3

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yoadey/shiftmanager/internal/port"
)

// Config configures the S3 adapter. Endpoint and ForcePathStyle are only
// needed for non-AWS S3-compatible providers (MinIO, Hetzner, …); leave
// Endpoint empty to talk to AWS S3 itself.
type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
	// PublicBaseURL, if set, overrides the URL objects are served from (e.g.
	// a CDN in front of the bucket, or a public bucket endpoint that differs
	// from Endpoint). Defaults to "<Endpoint>/<Bucket>" (or the AWS S3
	// virtual-hosted URL when Endpoint is empty).
	PublicBaseURL string
}

// Storage implements port.MediaStorage against an S3-compatible bucket.
type Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

var _ port.MediaStorage = (*Storage)(nil)

// New creates a Storage for the given configuration, resolving credentials
// and endpoint eagerly so misconfiguration (e.g. an unreachable endpoint's
// region resolution) surfaces at startup rather than on the first upload.
func New(ctx context.Context, cfg Config) (*Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 storage: bucket is required")
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		loadOpts = append(loadOpts, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("s3 storage: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			endpoint := cfg.Endpoint
			o.BaseEndpoint = &endpoint
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})

	publicBase := cfg.PublicBaseURL
	if publicBase == "" {
		switch {
		case cfg.Endpoint != "" && cfg.ForcePathStyle:
			publicBase = strings.TrimRight(cfg.Endpoint, "/") + "/" + cfg.Bucket
		case cfg.Endpoint != "":
			// Virtual-hosted-style: "<scheme>://<bucket>.<host>"
			scheme, host, found := strings.Cut(cfg.Endpoint, "://")
			if !found {
				scheme, host = "https", cfg.Endpoint
			}
			publicBase = scheme + "://" + cfg.Bucket + "." + strings.TrimRight(host, "/")
		default:
			// Real AWS S3: use the region the SDK actually resolved (it may
			// come from AWS_REGION/a shared profile rather than cfg.Region),
			// not cfg.Region directly — an empty region here would silently
			// produce a malformed "bucket.s3..amazonaws.com" host.
			if awsCfg.Region == "" {
				return nil, fmt.Errorf("s3 storage: no AWS region configured (set S3_REGION, or AWS_REGION/a shared AWS profile) and no S3_PUBLIC_BASE_URL override given")
			}
			publicBase = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, awsCfg.Region)
		}
	}

	return &Storage{client: client, bucket: cfg.Bucket, publicBaseURL: strings.TrimRight(publicBase, "/")}, nil
}

// Put uploads r as a single object (event attachments and logos are small —
// a handful of MB, well under S3's single-PUT limit — so no multipart
// upload manager is needed).
func (s *Storage) Put(ctx context.Context, name string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucket,
		Key:           &name,
		Body:          r,
		ContentLength: &size,
		ContentType:   &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}
	return s.publicBaseURL + "/" + name, nil
}

// Delete removes the object a previous Put returned as url. Only the last
// path segment (the object key this adapter itself generated) is used —
// path.Base collapses any ".."/"." segments, mirroring the local adapter's
// use of filepath.Base (path.Base rather than filepath.Base since an S3 key
// is always "/"-separated regardless of the host OS) — so an arbitrary or
// crafted URL never causes a delete outside what this adapter wrote.
func (s *Storage) Delete(ctx context.Context, url string) error {
	key := path.Base(url)
	if key == "" || key == "." || key == "/" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("s3 delete object: %w", err)
	}
	return nil
}
