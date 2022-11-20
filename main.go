package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

// move into package
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

// move into package
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

func getEntries(c *gin.Context) {

	entries := edb.getEntries()

	if entries == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No entries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func addEntry(c *gin.Context) {

	food := c.Param("food")
	calories := c.Param("calories")

	entryCalories, err := strconv.Atoi(calories)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal parameter"})
		return
	}

	added, _ := edb.addEntry(entry{time.Now().Unix(), food, entryCalories})
	if added {
		c.JSON(http.StatusOK, gin.H{"message": "entry added"})
	} // else TODO
}

func init() {
	const dbFile = "database.db"

	file, err := os.Create(dbFile)
	checkErr(err)
	file.Close()

	database, _ := sql.Open("sqlite3", dbFile)
	edb = EntryDB{database}
	edb.createTable()
}

var edb EntryDB

func main() {

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", getEntries)
		v1.POST("entry/:food/:calories", addEntry)
	}

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	r.Run()
}
