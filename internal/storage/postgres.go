package storage

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
)

type PostgresStorage struct {
	pgDB *sql.DB
	log  *logrus.Entry
	cfg  *config.Config
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

	pgDB, err := sql.Open("postgres", storage.cfg.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres")
	}
	storage.pgDB = pgDB

	query := `
    CREATE TABLE IF NOT EXISTS urls (
        uuid SERIAL NOT NULL,
        short_url TEXT NOT NULL,
        original_url TEXT NOT NULL, 
        is_deleted BOOLEAN NOT NULL DEFAULT FALSE
    );`

	_, err = storage.pgDB.Exec(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create urls table")
	}

	return storage, err
}
