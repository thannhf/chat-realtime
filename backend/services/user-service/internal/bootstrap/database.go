package bootstrap

import (
	"context"
	"fmt"
	"time"
	"user-service/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Write *gorm.DB
	Read  *gorm.DB
}

func NewDatabase(cfg *config.DBConfig) (*Database, error) {
	writeDB, err := connectWithRetry(buildDSN(cfg.WriteHost, cfg))
	if err != nil {
		return nil, err
	}

	readDB, err := connectWithRetry(buildDSN(cfg.ReadHost, cfg))
	if err != nil {
		return nil, err
	}

	return &Database{
		Write: writeDB,
		Read:  readDB,
	}, nil
}

func buildDSN(host string, cfg *config.DBConfig) string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
	return dsn
}

func connectWithRetry(dsn string) (*gorm.DB, error) {
	const (
		maxRetries   = 5
		initialDelay = time.Second
	)

	delay := initialDelay

	var (
		db  *gorm.DB
		err error
	)

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err == nil {
			sqlDB, err := db.DB()

			if err == nil {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					30*time.Second,
				)
				err = sqlDB.PingContext(ctx)
				cancel()

				if err == nil {
					sqlDB.SetMaxOpenConns(50)
					sqlDB.SetMaxIdleConns(10)
					sqlDB.SetConnMaxLifetime(time.Hour)
					sqlDB.SetConnMaxIdleTime(15 * time.Minute)

					return db, nil
				}
			}
		}

		fmt.Printf("database connection failed (%d/%d): %v\n", i, maxRetries, err)

		if i < maxRetries {
			time.Sleep(delay)
			delay *= 2
		}
	}

	return nil, fmt.Errorf("connect db failed after %d retries: %w", maxRetries, err)
}

func (db *Database) Close() error {
	var errorsRead error
	var errorsWrite error

	if db.Write != nil {
		if sqlDB, err := db.Write.DB(); err == nil {
			if errClose := sqlDB.Close(); errClose != nil {
				errorsWrite = errClose
			}
		} else {
			errorsWrite = err
		}
	}

	if db.Read != nil {
		if sqlDB, err := db.Read.DB(); err == nil {
			if errClose := sqlDB.Close(); errClose != nil {
				errorsRead = errClose
			}
		} else {
			errorsRead = err
		}
	}

	if errorsWrite != nil && errorsRead != nil {
		return fmt.Errorf("write db error %v; read db error %v: ", errorsWrite, errorsRead)
	}

	if errorsWrite != nil {
		return errorsWrite
	}

	if errorsRead != nil {
		return errorsRead
	}

	return nil
}
