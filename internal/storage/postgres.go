package storage

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

type PostgresStorage struct {
	Postgres *sql.DB
	Queries  *db.Queries
	log      *logrus.Entry
	cfg      *config.Config
}

func NewPostgresStorage(i do.Injector) (*PostgresStorage, error) {
	storage, err := do.InvokeStruct[PostgresStorage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "postgres")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	storage.log = log
	storage.cfg = cfg

	pgDB, err := sql.Open("postgres", storage.cfg.Database.URI)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres")
	}

	storage.Postgres = pgDB

	err = storage.Migrate()
	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	storage.Queries = db.New(pgDB)

	return storage, err
}

func (s *PostgresStorage) Close() error {
	return s.Postgres.Close()
}

func (s *PostgresStorage) HealthCheck() error {
	return s.Postgres.Ping()
}

func (s *PostgresStorage) Migrate() error {
	//create tables in db
	//open schema file
	query, err := os.ReadFile("./internal/storage/schema/schema.sql")
	if err != nil {
		return errors.Wrap(err, "read schema")
	}

	//exec query
	_, err = s.Postgres.Exec(string(query))
	if err != nil {
		//TODO: fix migration
		return nil
	}

	return nil
}
