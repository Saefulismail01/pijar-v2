package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Driver   string
}

type APIConfig struct {
	ApiPort string
}

type Config struct {
	DBConfig
	APIConfig
}

func (c *Config) readConfig() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	c.DBConfig = DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
		Driver:   os.Getenv("DB_DRIVER"),
	}

	c.APIConfig = APIConfig{
		ApiPort: os.Getenv("API_PORT"),
	}

	if c.Host == "" || c.Port == "" || c.User == "" || c.Password == "" || c.DBName == "" || c.ApiPort == "" {
		return fmt.Errorf("required config")
	}
	return nil
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := cfg.readConfig(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// package config
//
// import (
// 	"fmt"
// 	"os"
// )
//
// type DBConfig struct {
// 	Host     string
// 	Port     string
// 	Database string
// 	Username string
// 	Password string
// 	Driver   string
// }
//
// type APIConfig struct {
// 	APIHost string
// 	APIPort string
// }
//
// type Config struct {
// 	DBConfig
// 	APIConfig
// }
//
// func NewConfig() (*Config, error) {
// 	cfg := &Config{}
// 	err := cfg.readConfig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cfg, nil
// }
//
// func (c *Config) readConfig() error {
// 	c.DBConfig = DBConfig{
// 		Host:     os.Getenv("DB_HOST"),
// 		Port:     os.Getenv("DB_PORT"),
// 		Database: os.Getenv("DB_NAME"),
// 		Username: os.Getenv("DB_USER"),
// 		Password: os.Getenv("DB_PASS"),
// 		Driver:   os.Getenv("DB_DRIVER"),
// 	}
//
// 	c.APIConfig = APIConfig{
// 		APIHost: os.Getenv("API_HOST"),
// 		APIPort: os.Getenv("API_PORT"),
// 	}
//
// 	if c.Host == "" ||
// 		c.Port == "" ||
// 		c.Database == "" ||
// 		c.Username == "" ||
// 		c.Password == "" ||
// 		c.Driver == "" ||
// 		c.APIHost == "" ||
// 		c.APIPort == "" {
// 		return fmt.Errorf("required config")
// 	}
// 	return nil
// }
