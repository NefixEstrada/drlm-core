// SPDX-License-Identifier: AGPL-3.0-only

package minio_test

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/utils/tests"
)

func (s *TestMinioSuite) TestCreateUser() {
	s.Run("should create the user correctly", func() {
		tests.GenerateCfg(s.T())

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v1/add-user", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		pwd, err := minio.CreateUser("nefix")
		s.NotEqual("", pwd)
		s.Nil(err)
	})

	s.Run("should return an error if there's an error creating the admin client", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.Host = ""
		cfg.Config.Minio.Port = 9443

		pwd, err := minio.CreateUser("nefix")
		s.Equal("", pwd)
		s.EqualError(err, "error connecting to the Minio admin API: Endpoint: :9443 does not follow ip address or domain name standards.")
	})

	s.Run("should return an error if there's an error creating the user", func() {
		tests.GenerateCfg(s.T())

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v1/add-user", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		pwd, err := minio.CreateUser("nefix")
		s.Equal("", pwd)
		s.EqualError(err, "error creating the minio user: Failed to parse server response.")
	})
}

func (s *TestMinioSuite) TestMakeBucketForUser() {
	s.Run("should create the bucket correctly", func() {
		tests.GenerateCfg(s.T())
		minio.Init()

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v1/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/minio/admin/v1/set-user-or-group-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		bName, err := minio.MakeBucketForUser("nefix")
		s.True(strings.HasPrefix(bName, "drlm-"))
		s.Nil(err)
	})

	s.Run("should return an error if there's an error creating the admin client", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.Host = ""
		cfg.Config.Minio.Port = 9443

		bName, err := minio.MakeBucketForUser("nefix")
		s.Equal("", bName)
		s.EqualError(err, "error connecting to the Minio admin API: Endpoint: :9443 does not follow ip address or domain name standards.")
	})

	s.Run("should return an error if there's an error creating the bucket", func() {
		tests.GenerateCfg(s.T())
		minio.Init()

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			s.Fail(r.URL.String())
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		bName, err := minio.MakeBucketForUser("nefix")
		s.Equal("", bName)
		s.EqualError(err, "error creating the storage bucket: The specified bucket does not exist.")
	})

	s.Run("should return an error if there's an error creating the policy", func() {
		tests.GenerateCfg(s.T())
		minio.Init()

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v1/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		bName, err := minio.MakeBucketForUser("nefix")
		s.Equal("", bName)
		s.EqualError(err, "error creating the policy: Failed to parse server response.")
	})

	s.Run("should return an error if there's an error adding the policy to the user", func() {
		tests.GenerateCfg(s.T())
		minio.Init()

		mux := http.NewServeMux()
		mux.HandleFunc("/minio/admin/v1/add-canned-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/minio/admin/v1/set-user-or-group-policy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.String(), "/drlm-") {
				w.WriteHeader(http.StatusOK)
				return
			}

			s.Fail(r.URL.String())
		})

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Config.Minio.Port),
			Handler: mux,
		}

		go srv.ListenAndServe()
		defer srv.Close()

		bName, err := minio.MakeBucketForUser("nefix")
		s.Equal("", bName)
		s.EqualError(err, "error applying the policy to the user: Failed to parse server response.")
	})
}
