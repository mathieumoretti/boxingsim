package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mormm/boxing/internal/db"
)

const migrationsDir = "db/migrations"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	var conn *sql.DB
	var err error

	switch command {
	case "up":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.MigrateUp(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "up-one":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.MigrateUpOne(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "down":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.MigrateDownToZero(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "down-one":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.MigrateDownOne(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "reset":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.ResetDatabase(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Reset failed: %v", err)
		}

	case "status":
		conn, err = connectDB()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer closeConn(conn)

		err = db.ShowStatus(conn, migrationsDir)
		if err != nil {
			log.Fatalf("Status failed: %v", err)
		}

	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Migration name required. Usage: migrate create migration_name")
		}
		name := os.Args[2]
		err = db.CreateMigration(name, migrationsDir)
		if err != nil {
			log.Fatalf("Create migration failed: %v", err)
		}

	case "help":
	default:
		printUsage()
		os.Exit(1)
	}
}

func connectDB() (*sql.DB, error) {
	dsn := buildDSN()
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB connection: %w", err)
	}
	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	log.Println("Connected to database")
	return conn, nil
}

func closeConn(conn *sql.DB) {
	if conn != nil {
		conn.Close()
		fmt.Println("\nClosed connection")
	}
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DATABASE_NAME", "boxingsimdb")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

func printUsage() {
	fmt.Println(`Boxing Sim Migrations Tool`)
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("  migrate up       Run all pending migrations")
	fmt.Println("  migrate up-one   Run the next pending migration only")
	fmt.Println("  migrate down     Rollback all applied migrations")
	fmt.Println("  migrate down-one Rollback the last applied migration")
	fmt.Println("  migrate reset    Drop and recreate database from zero")
	fmt.Println("  migrate status   Show current migration status (applied/pending)")
	fmt.Println("  create <name>    Create a new named migration file with up/down blocks")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
