package database

import (
   
    "fmt"
    "time"

    _ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
)

type PostgresDB struct {
    DB *sqlx.DB
}

func NewPostgresConnection(databaseURL string) (*PostgresDB, error) {
    db, err := sqlx.Connect("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to postgres: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping postgres: %w", err)
    }

    return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
    return p.DB.Close()
}

func (p *PostgresDB) Ping() error {
    return p.DB.Ping()
}