// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"fmt"
	"strings"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/secret"

	"github.com/rs/xid"
)

// CreateUser creates a new user to the Minio server
// TODO: Add S3 compatibility
func CreateUser(ctx *context.Context, usr string) (key string, err error) {
	pwd, err := secret.New(usr)
	if err != nil {
		return "", fmt.Errorf("error generating the secret key: %v", err)
	}

	if err := ctx.MinioAdminCli.AddUser(usr, pwd); err != nil {
		return "", fmt.Errorf("error creating the minio user: %v", err)
	}

	return pwd, nil
}

// MakeBucketForUser creates a new bucket and adds read and write permissions to the user
// TODO: Add S3 compatibility
func MakeBucketForUser(ctx *context.Context, usr string, name ...string) (string, error) {
	var bName string
	if len(name) == 0 {
		bName = fmt.Sprintf("drlm-%s", xid.New())
	} else {
		bName = name[0]
	}

	if err := ctx.MinioCli.MakeBucket(bName, ctx.Cfg.Minio.Location); err != nil {
		return "", fmt.Errorf("error creating the storage bucket: %v", err)
	}

	if err := ctx.MinioAdminCli.AddCannedPolicy(bName, strings.Replace(`{
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

	if err := ctx.MinioAdminCli.SetPolicy(bName, usr, false); err != nil {
		return "", fmt.Errorf("error applying the policy to the user: %v", err)
	}

	return bName, nil
}
