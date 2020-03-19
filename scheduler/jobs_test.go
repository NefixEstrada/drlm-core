// SPDX-License-Identifier: AGPL-3.0-only

package scheduler_test

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/scheduler"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type TestJobsSuite struct {
	suite.Suite
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestJobsSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
	tests.GenerateCfg(s.T(), s.ctx)
}

func TestJobs(t *testing.T) {
	suite.Run(t, &TestJobsSuite{})
}

func (s *TestJobsSuite) TestAddJob() {
	s.Run("should add the job correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).AddRow(1, "192.168.1.61"))
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name"}).AddRow(1, "default", "tar"))
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		minio.Init(s.ctx)

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

		ts := tests.GenerateMinio(s.ctx, mux)
		defer ts.Close()

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.NoError(err)
	})

	s.Run("should return an error if there's an error loading the agent from the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.EqualError(err, "error loading the agent from the DB: testing error")
	})

	s.Run("should return an error if there's an error loading the plugins of the agent from the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).AddRow(1, "192.168.1.61"))
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnError(errors.New("testing error"))

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.EqualError(err, "error getting the plugins list: testing error")
	})

	s.Run("should return an error if the agent doesn't have the plugin requested", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).AddRow(1, "192.168.1.61"))
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name"}))

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.EqualError(err, "plugin for the job not found in the agent")
	})

	s.Run("should return an error if there's an error creating the minio bucket", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).AddRow(1, "192.168.1.61"))
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name"}).AddRow(1, "default", "tar"))

		minio.Init(s.ctx)

		ts := tests.GenerateMinio(s.ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.EqualError(err, "error adding the job: error creating the storage bucket: The specified bucket does not exist.")
	})

	s.Run("should return an error if there's an error adding the job to the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).AddRow(1, "192.168.1.61"))
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name"}).AddRow(1, "default", "tar"))
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnError(errors.New("testing error"))

		minio.Init(s.ctx)

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

		ts := tests.GenerateMinio(s.ctx, mux)
		defer ts.Close()

		err := scheduler.AddJob(s.ctx, "192.168.1.61", "default/tar", "", time.Now())

		s.EqualError(err, "error adding the job: error adding the job to the DB: testing error")
	})
}
