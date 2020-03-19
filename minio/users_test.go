// SPDX-License-Identifier: AGPL-3.0-only

package minio_test

import (
	"net/http"
	"strings"

	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/utils/tests"
)

func (s *TestMinioSuite) TestCreateUser() {
	s.Run("should create the user correctly", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)

		ts := tests.GenerateMinio(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		pwd, err := minio.CreateUser(ctx, "nefix")
		s.NotEqual("", pwd)
		s.Nil(err)
	})

	s.Run("should return an error if there's an error creating the user", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)

		ts := tests.GenerateMinio(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		pwd, err := minio.CreateUser(ctx, "nefix")
		s.Equal("", pwd)
		s.EqualError(err, "error creating the minio user: Failed to parse server response.")
	})
}

func (s *TestMinioSuite) TestMakeBucketForUser() {
	s.Run("should create the bucket correctly", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		minio.Init(ctx)

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v2/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/minio/admin/v2/set-user-or-group-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		ts := tests.GenerateMinio(ctx, mux)
		defer ts.Close()

		bName, err := minio.MakeBucketForUser(ctx, "nefix")
		s.True(strings.HasPrefix(bName, "drlm-"))
		s.Nil(err)
	})

	s.Run("should return an error if there's an error creating the bucket", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		minio.Init(ctx)

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			s.Fail(r.URL.String())
		})

		ts := tests.GenerateMinio(ctx, mux)
		defer ts.Close()

		bName, err := minio.MakeBucketForUser(ctx, "nefix")
		s.Equal("", bName)
		s.EqualError(err, "error creating the storage bucket: The specified bucket does not exist.")
	})

	s.Run("should return an error if there's an error creating the policy", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		minio.Init(ctx)

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v2/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		ts := tests.GenerateMinio(ctx, mux)
		defer ts.Close()

		bName, err := minio.MakeBucketForUser(ctx, "nefix")
		s.Equal("", bName)
		s.EqualError(err, "error creating the policy: Failed to parse server response.")
	})

	s.Run("should return an error if there's an error adding the policy to the user", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		minio.Init(ctx)

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v2/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/minio/admin/v2/set-user-or-group-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		ts := tests.GenerateMinio(ctx, mux)
		defer ts.Close()

		bName, err := minio.MakeBucketForUser(ctx, "nefix")
		s.Equal("", bName)
		s.EqualError(err, "error applying the policy to the user: Failed to parse server response.")
	})
}
