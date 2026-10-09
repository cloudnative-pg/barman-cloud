/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package restorer

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

const walSegmentBytes = 16 * 1024 * 1024

var walSegmentName = regexp.MustCompile(`^[0-9A-F]{24}$`)

// PersistentConnection is the object-store location used when WAL restore
// reuses one S3 client instead of starting barman-cloud-wal-restore.
type PersistentConnection struct {
	EndpointURL     string
	DestinationPath string
	ServerName      string
	Compression     string
	AddressingStyle string
}

type persistentClient struct {
	bucket      string
	prefix      string
	serverName  string
	compression string
	client      *s3.Client
}

type clientCacheKey struct {
	endpoint, bucket, accessKey, secretSum, sessionSum, region, style, ca string
}

var (
	clientCacheMu sync.Mutex
	clientCache   = map[clientCacheKey]*s3.Client{}
)

// EnablePersistentConnection makes later restores of regular WAL segments reuse
// one S3 client built from the credentials already stored in the restorer environment.
func (restorer *WALRestorer) EnablePersistentConnection(spec PersistentConnection) error {
	bucket, prefix, err := splitS3URL(spec.DestinationPath)
	if err != nil {
		return err
	}
	if spec.ServerName == "" {
		return fmt.Errorf("server name is empty")
	}
	accessKey := envValue(restorer.env, "AWS_ACCESS_KEY_ID")
	secretKey := envValue(restorer.env, "AWS_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		return fmt.Errorf("AWS access key is not in the restore environment")
	}

	sessionToken := envValue(restorer.env, "AWS_SESSION_TOKEN")
	region := envValue(restorer.env, "AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}
	caFile := envValue(restorer.env, "REQUESTS_CA_BUNDLE")
	style := spec.AddressingStyle
	if style == "" {
		style = "auto"
	}

	client, err := cachedS3Client(clientCacheKey{
		endpoint:   spec.EndpointURL,
		bucket:     bucket,
		accessKey:  accessKey,
		secretSum:  shortSum(secretKey),
		sessionSum: shortSum(sessionToken),
		region:     region,
		style:      style,
		ca:         caFile,
	}, spec.EndpointURL, region, accessKey, secretKey, sessionToken, caFile, style)
	if err != nil {
		return err
	}

	restorer.persistent = &persistentClient{
		bucket:      bucket,
		prefix:      prefix,
		serverName:  spec.ServerName,
		compression: spec.Compression,
		client:      client,
	}
	return nil
}

func (client *persistentClient) restore(walName, destinationPath string) error {
	key := walObjectKey(client.prefix, client.serverName, walName, client.compression)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := client.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(client.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("object storage or file not found %s: %w", walName, ErrWALNotFound)
		}
		return fmt.Errorf("connectivity failure while reading %s: %w", walName, ErrConnectivity)
	}
	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("connectivity failure while reading %s: %w", walName, ErrConnectivity)
	}
	plain, err := decodeWAL(body, client.compression)
	if err != nil {
		return err
	}
	if len(plain) != walSegmentBytes {
		return fmt.Errorf("WAL %s decoded to %d bytes", walName, len(plain))
	}
	return writeAtomically(destinationPath, plain)
}

func isWALSegment(name string) bool {
	return walSegmentName.MatchString(name)
}

func walObjectKey(prefix, serverName, walName, compression string) string {
	extension := ""
	if compression == "gzip" {
		extension = ".gz"
	}
	return path.Join(prefix, serverName, "wals", walName[:16], walName+extension)
}

func splitS3URL(raw string) (bucket, prefix string, err error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "s3" || parsed.Host == "" {
		return "", "", fmt.Errorf("destinationPath %q is not an s3 URL", raw)
	}
	return parsed.Host, strings.Trim(parsed.Path, "/"), nil
}

func decodeWAL(body []byte, compression string) ([]byte, error) {
	if compression != "gzip" {
		return body, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("while reading gzip WAL: %w", err)
	}
	defer reader.Close()
	plain, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("while decompressing gzip WAL: %w", err)
	}
	return plain, nil
}

func writeAtomically(destinationPath string, plain []byte) error {
	if err := os.MkdirAll(path.Dir(destinationPath), 0o750); err != nil {
		return fmt.Errorf("while creating %s: %w", path.Dir(destinationPath), err)
	}
	partial := destinationPath + ".partial"
	if err := os.WriteFile(partial, plain, 0o640); err != nil {
		return fmt.Errorf("while writing %s: %w", partial, err)
	}
	if err := os.Rename(partial, destinationPath); err != nil {
		return fmt.Errorf("while renaming %s: %w", destinationPath, err)
	}
	return nil
}

func cachedS3Client(
	key clientCacheKey,
	endpoint, region, accessKey, secretKey, sessionToken, caFile, style string,
) (*s3.Client, error) {
	clientCacheMu.Lock()
	defer clientCacheMu.Unlock()
	if client, ok := clientCache[key]; ok {
		return client, nil
	}

	transport := &http.Transport{
		MaxIdleConns:        32,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
	}
	if caFile != "" && strings.HasPrefix(endpoint, "https://") {
		pool, err := certPoolFromFile(caFile)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    pool,
		}
	}

	options := s3.Options{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			sessionToken,
		),
		UsePathStyle: style == "path" || (style != "virtual" && endpoint != ""),
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
	if endpoint != "" {
		options.BaseEndpoint = aws.String(endpoint)
	}
	client := s3.New(options)
	clientCache[key] = client
	return client, nil
}

func certPoolFromFile(caFile string) (*x509.CertPool, error) {
	pemBytes, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("while reading %s: %w", caFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("while parsing certificates in %s", caFile)
	}
	return pool, nil
}

func isNotFound(err error) bool {
	var response *awshttp.ResponseError
	if errors.As(err, &response) && response.HTTPStatusCode() == http.StatusNotFound {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey", "404":
			return true
		}
	}
	return false
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func shortSum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:8])
}
