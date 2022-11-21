package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type entryDB struct {
	db *sql.DB
}

type entry struct {
	Timestamp int64
	Food      string
	Calories  int
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// idempotent
func (edb entryDB) createTableIfNeeded() {
	entriesTable := `CREATE TABLE entries (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"timestamp" INTEGER,
        "food" TEXT,
        "calories" INTEGER);`
	query, err := edb.db.Prepare(entriesTable)
	if err != nil {
		if err.Error() == "table entries already exists" {
			log.Println("table already exists, skip creation")
			return
		} else {
			checkErr(err)
		}
	}

	query.Exec()
	log.Println("table created successfully")
}

// move into package
func (edb entryDB) getEntries() []entry {

	rows, _ := edb.db.Query("SELECT timestamp, food, calories FROM entries")

	defer rows.Close()

	err := rows.Err()
	checkErr(err)

	entries := make([]entry, 0)

	for rows.Next() {
		anEntry := entry{}
		err = rows.Scan(&anEntry.Timestamp, &anEntry.Food, &anEntry.Calories)
		checkErr(err)

		entries = append(entries, anEntry)
	}

	err = rows.Err()
	checkErr(err)

	return entries
}

// move into package
func (edb entryDB) addEntry(newEntry entry) (bool, error) {

	tx, err := edb.db.Begin()
	if err != nil {
		return false, err
	}

	stmt, err := tx.Prepare("INSERT INTO entries (timestamp, food, calories) VALUES (?, ?, ?)")
	if err != nil {
		return false, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(newEntry.Timestamp, newEntry.Food, newEntry.Calories)
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

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

// tries to convert given string into timestamp
// defaults to current Unix epoch time
func defaultTimestamp(s string) int64 {

	s = s[1:] // chop leading /

	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return n
	}

	return time.Now().Unix()
}

func addEntry(c *gin.Context) {

	food := c.Param("food")
	calories := c.Param("calories")
	timestamp := defaultTimestamp(c.Param("timestamp")) // contains at least leading / due to redirect

	entryCalories, err := strconv.Atoi(calories)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal parameter"})
		return
	}

	added, err := edb.addEntry(entry{timestamp, food, entryCalories})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if added {
		c.JSON(http.StatusOK, gin.H{"message": "entry added"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entry not added"})
	}
}

func init() {
	const dbFile = "clcnt.db"

	_, err := os.Stat(dbFile)
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(dbFile) // re-creates!
		checkErr(err)
		file.Close()
		log.Println("database file created")
	} else {
		log.Println("database file already exists")
	}

	database, _ := sql.Open("sqlite3", dbFile)
	edb = entryDB{database}
	edb.createTableIfNeeded()
}

var edb entryDB

func main() {

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", getEntries)
		v1.POST("entry/:food/:calories/*timestamp", addEntry)
	}

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	r.Run()
}
