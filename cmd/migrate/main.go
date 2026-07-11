package main

import (
	"errors"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/harryho/nw-api-gogin/internal/db"
)

func main() {
	action := flag.String("action", "up", "migration action: up, down, seed, drop")
	steps := flag.Int("steps", 0, "number of steps for down")
	flag.Parse()

	cfg := db.LoadConfig()
	sqlDB, err := db.OpenSQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	source, err := iofs.New(db.Migrations, db.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithInstance("iofs", source, cfg.Name, driver)
	if err != nil {
		log.Fatal(err)
	}
	defer closeMigration(m)

	if err := run(*action, *steps, m); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}
}

func run(action string, steps int, m *migrate.Migrate) error {
	switch action {
	case "up":
		return m.Up()
	case "seed":
		return m.Up()
	case "down":
		if steps > 0 {
			return m.Steps(-steps)
		}
		return m.Down()
	case "drop":
		return m.Drop()
	default:
		return errors.New("unsupported migration action")
	}
}

func closeMigration(m *migrate.Migrate) {
	if sourceErr, dbErr := m.Close(); sourceErr != nil {
		log.Print(sourceErr)
	} else if dbErr != nil {
		log.Print(dbErr)
	}
}
