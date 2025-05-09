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

// package config
//
// import (
// 	"database/sql"
// 	"fmt"
//
// 	_ "github.com/lib/pq"
// )
//
// func ConnectDB(cfg DBConfig) (*sql.DB, error) {
// 	connStr := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		cfg.Host,
// 		cfg.Port,
// 		cfg.Username,
// 		cfg.Password,
// 		cfg.Database,
// 	)
//
// 	db, err := sql.Open(cfg.Driver, connStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %v", err)
// 	}
//
// 	err = db.Ping()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %v", err)
// 	}
// 	return db, nil
// }
