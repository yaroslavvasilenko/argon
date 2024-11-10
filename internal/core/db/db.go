package db

import (
	"fmt"
	"github.com/yaroslavvasilenko/argon/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewSqlDB(dbConf config.Config) (*gorm.DB, error) {
	dsnRaw := "user=%s password=%s host=%s port=%s dbname=%s sslmode=disable"

	dsn := fmt.Sprintf(dsnRaw,
		dbConf.DB.User, dbConf.DB.Pass, dbConf.DB.Host, dbConf.DB.Port, dbConf.DB.Name)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
