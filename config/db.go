package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, *Config, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		panic("gagal terkoneksi")
	}

	return db, cfg, nil
}
