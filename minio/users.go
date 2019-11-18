package minio

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/secret"

	"github.com/rs/xid"
	"github.com/minio/minio/pkg/madmin"
)

func newAdminClient() (*madmin.AdminClient, error) {
	cli, err := madmin.New(conn())
	if err != nil {
		return nil, fmt.Errorf("error connecting to the Minio admin API: %v", err)
	}
	// If the certificate is self signed, add it to the transport certificates pool
	if cfg.Config.Minio.SSL && cfg.Config.Minio.CertPath != "" {
		tr := http.DefaultTransport.(*http.Transport)
		if err := transport(tr); err != nil {
			return nil, err
		}

		cli.SetCustomTransport(tr)
	}

	return cli, nil
}

// CreateUser creates a new user to the Minio server
// TODO: Add S3 compatibility
func CreateUser(usr string) (key string, err error) {
	cli, err := newAdminClient()
	if err != nil {
		return "", err
	}

	pwd, err := secret.New(usr)
	if err != nil {
		return "", fmt.Errorf("error generating the secret key: %v", err)
	}

	if err := cli.AddUser(usr, pwd); err != nil {
		return "", fmt.Errorf("error creating the minio user: %v", err)
	}

	return pwd, nil
}

// MakeBucketForUser creates a new bucket and adds read and write permissions to the user
// TODO: Add S3 compatibility
func MakeBucketForUser(usr string) (string, error) {
	cli, err := newAdminClient()
	if err != nil {
		return "", err
	}

	bName := fmt.Sprintf("drlm-%s", xid.New())

	if err := Minio.MakeBucket(bName, cfg.Config.Minio.Location); err != nil {
		return "", fmt.Errorf("error creating the storage bucket: %v", err)
	}

	if err := cli.AddCannedPolicy(bName, strings.Replace(`{
		"Version": "2012-10-17",
		"Id": "{BUCKET_NAME}",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": [
						"*"
					]
				},
				"Action": [
					"s3:ListBucket",
					"s3:ListBucketMultipartUploads",
					"s3:GetBucketLocation"
				],
				"Resource": [
					"arn:aws:s3:::{BUCKET_NAME}"
				]
			},
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": [
						"*"
					]
				},
				"Action": [
					"s3:DeleteObject",
					"s3:GetObject",
					"s3:ListMultipartUploadParts",
					"s3:PutObject",
					"s3:AbortMultipartUpload"
				],
				"Resource": [
					"arn:aws:s3:::{BUCKET_NAME}/*"
				]
			}
		]
	}`, "{BUCKET_NAME}", bName, -1)); err != nil {
		return "", fmt.Errorf("error creating the policy: %v", err)
	}

	if err := cli.SetPolicy(bName, usr, false); err != nil {
		return "", fmt.Errorf("error applying the policy to the user: %v", err)
	}

	return bName, nil
}
