package database

import (
	"net"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"sispat/backend/internal/config"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	dsnConfig := mysqldriver.Config{
		User:      cfg.DBUser,
		Passwd:    cfg.DBPass,
		Net:       "tcp",
		Addr:      net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		DBName:    cfg.DBName,
		ParseTime: true,
		Params: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_0900_ai_ci",
		},
	}

	return gorm.Open(mysql.Open(dsnConfig.FormatDSN()), &gorm.Config{
		TranslateError: true,
	})
}
