package models

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

type Registry struct {
}

type Entry struct {
	Timestamp int64
	Food      string
	Calories  int
}

var rdb *regDb

type regDb struct {
	db *sql.DB
}

func NewRegistry() (*Registry, error) {
	const fn = "clcnt.db"

	_, err := os.Stat(fn)
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(fn) // re-creates!
		if err != nil {
			return nil, err
		}
		f.Close()
		log.Info("database file created")
	} else {
		log.Info("database file already exists")
	}

	db, err := sql.Open("sqlite3", fn)
	if err != nil {
		return nil, err
	}

	rdb = &regDb{db}
	err = createTableIfNeeded()
	if err != nil {
		return nil, err
	}

	return &Registry{}, nil
}

// AddEntry adds the given Entry
func (*Registry) AddEntry(entry Entry) error {
	tx, err := rdb.db.Begin()
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

// GetEntries retrieves all entries
func (*Registry) GetEntries() ([]Entry, error) {
	r, _ := rdb.db.Query("SELECT timestamp, food, calories FROM entries")
	defer r.Close()

	err := r.Err()
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)

	for r.Next() {
		entry := Entry{}
		err = r.Scan(&entry.Timestamp, &entry.Food, &entry.Calories)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	err = r.Err()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// GetCalories sums up calory entries since given time as Unix timestamp
func (*Registry) GetCalories(t int64) (int, error) {
	stmt, err := rdb.db.Prepare("SELECT SUM(calories) FROM entries WHERE timestamp >= ?")
	if err != nil {
		return -1, err
	}

	var calories int
	sqlErr := stmt.QueryRow(t).Scan(&calories)
	if sqlErr != nil {
		if sqlErr == sql.ErrNoRows || strings.HasPrefix(sqlErr.Error(), "sql: Scan error on column") {
			log.Info("empty result")
			return 0, nil
		}
		return -1, sqlErr
	}

	return calories, nil
}

// idempotent
func createTableIfNeeded() error {
	t := `CREATE TABLE entries (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"timestamp" INTEGER,
        "food" TEXT,
        "calories" INTEGER);`
	q, err := rdb.db.Prepare(t)
	if err != nil {
		if err.Error() == "table entries already exists" {
			log.Info("table already exists, skip creation")
			return nil
		} else {
			return err
		}
	}

	q.Exec()
	log.Info("table created successfully")

	return nil
}
