package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type EntryDB struct {
	db *sql.DB
}

type entry struct {
	timestamp int64
	food      string
	calories  int
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// idempotent
func (edb EntryDB) createTable() {
	entriesTable := `CREATE TABLE entries (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"timestamp" INTEGER,
        "food" TEXT,
        "calories" INTEGER);`
	query, err := edb.db.Prepare(entriesTable)
	checkErr(err)

	query.Exec()
	fmt.Println("Table created successfully!")
}

func (edb EntryDB) getEntries() []entry {

	rows, _ := edb.db.Query("SELECT timestamp, food, calories FROM entries")

	defer rows.Close()

	err := rows.Err()
	checkErr(err)

	entries := make([]entry, 0)

	for rows.Next() {
		anEntry := entry{}
		err = rows.Scan(&anEntry.timestamp, &anEntry.food, &anEntry.calories)
		checkErr(err)

		entries = append(entries, anEntry)
	}

	err = rows.Err()
	checkErr(err)

	return entries
}

func (edb EntryDB) addEntry(newEntry entry) (bool, error) {

	tx, err := edb.db.Begin()
	if err != nil {
		return false, err
	}

	stmt, err := tx.Prepare("INSERT INTO entries (timestamp, food, calories) VALUES (?, ?, ?)")
	if err != nil {
		return false, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(newEntry.timestamp, newEntry.food, newEntry.calories)
	if err != nil {
		return false, err
	}

	tx.Commit()

	return true, nil
}

func main() {

	const dbFile = "database.db"

	file, err := os.Create(dbFile)
	checkErr(err)
	file.Close()

	database, _ := sql.Open("sqlite3", dbFile)
	edb := EntryDB{database}
	edb.createTable()
	added, _ := edb.addEntry(entry{time.Now().Unix(), "Bretzel", 300})
	if added {
		fmt.Println("Entry added successfully")
		jf := edb.getEntries()[0]
		timestamp := time.Unix(jf.timestamp, 0)
		fmt.Printf("%v: %s %d\n", timestamp, jf.food, jf.calories)
	}
}
