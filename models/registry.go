package models

import (
	"database/sql"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var edb *entryDB

type entryDB struct {
	db *sql.DB
}

type Registry struct {
}

type Entry struct {
	Timestamp int64
	Food      string
	Calories  int
}

// idempotent
func createTableIfNeeded() error {
	entriesTable := `CREATE TABLE entries (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"timestamp" INTEGER,
        "food" TEXT,
        "calories" INTEGER);`
	query, err := edb.db.Prepare(entriesTable)
	if err != nil {
		if err.Error() == "table entries already exists" {
			log.Println("table already exists, skip creation")
			return nil
		} else {
			return err
		}
	}

	query.Exec()
	log.Println("table created successfully")
	return nil
}

// GetEntries retrieves all entries
func (*Registry) GetEntries() ([]Entry, error) {

	rows, _ := edb.db.Query("SELECT timestamp, food, calories FROM entries")

	defer rows.Close()

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)

	for rows.Next() {
		anEntry := Entry{}
		err = rows.Scan(&anEntry.Timestamp, &anEntry.Food, &anEntry.Calories)
		if err != nil {
			return nil, err
		}

		entries = append(entries, anEntry)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// AddEntry adds the given Entry
func (*Registry) AddEntry(entry Entry) error {

	tx, err := edb.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO entries (timestamp, food, calories) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(entry.Timestamp, entry.Food, entry.Calories)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func NewRegistry() (*Registry, error) {
	const dbFile = "clcnt.db"

	_, err := os.Stat(dbFile)
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(dbFile) // re-creates!
		if err != nil {
			return nil, err
		}
		file.Close()
		log.Println("database file created")
	} else {
		log.Println("database file already exists")
	}

	database, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	edb = &entryDB{database}
	err = createTableIfNeeded()
	if err != nil {
		return nil, err
	}

	return &Registry{}, nil
}
