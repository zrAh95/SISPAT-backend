package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"sispat/backend/internal/config"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_0900_ai_ci&parseTime=true&multiStatements=true",
		cfg.DBUser,
		cfg.DBPass,
		net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var migrationErr error
	switch command {
	case "up":
		migrationErr = migrator.Up()
	case "down":
		migrationErr = migrator.Steps(-1)
	case "version":
		version, dirty, versionErr := migrator.Version()
		if versionErr != nil {
			log.Fatal(versionErr)
		}
		fmt.Printf("versi=%d dirty=%t\n", version, dirty)
		return
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("gunakan: go run ./cmd/migrate force <versi>")
		}
		version, parseErr := strconv.Atoi(os.Args[2])
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		if forceErr := migrator.Force(version); forceErr != nil {
			log.Fatal(forceErr)
		}
		fmt.Printf("versi migration dipaksa menjadi %d\n", version)
		return
	default:
		log.Fatalf("perintah tidak dikenal: %s (gunakan up, down, version, atau force)", command)
	}

	if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNoChange) {
		log.Fatal(migrationErr)
	}

	fmt.Println("migration selesai")
}
