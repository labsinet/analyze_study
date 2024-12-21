Migration Setup Code

// migrations/migrations.go
package migrations

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func SetupMigrations(db *sql.DB, migrationsPath string) error {
    driver, err := mysql.WithInstance(db, &mysql.Config{})
    if err != nil {
        return fmt.Errorf("could not create the mysql driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        "mysql",
        driver,
    )
    if err != nil {
        return fmt.Errorf("could not create the migration instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("could not run up migrations: %v", err)
    }

    log.Println("Successfully ran migrations")
    return nil
}