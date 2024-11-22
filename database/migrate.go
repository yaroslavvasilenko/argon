package database

import (
	"database/sql"
	"embed"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

const migrationPath string = "migrations"
const postgresDialect string = "postgres"

func Migrate(url string) error {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return errors.Wrap(err, "pinging db")
	}
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect(postgresDialect); err != nil {
		return err
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return errors.Wrap(err, "getting db version")
	}

	err = goose.Up(db, migrationPath)
	if err != nil {
		if err := goose.DownTo(db, migrationPath, version); err != nil {
			return err
		}

		return errors.Wrap(err, "cannot migrate db")
	}

	return nil
}
