// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"fmt"
	"strings"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/secret"

	"github.com/rs/xid"
)

// CreateUser creates a new user to the Minio server
// TODO: Add S3 compatibility
func CreateUser(usr string) (key string, err error) {
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
