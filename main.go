package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type PersonDB struct {
	db *sql.DB
}

type person struct {
	first_name string
	last_name  string
}

// idempotent
func (pdb PersonDB) createTable() {
	persons_table := `CREATE TABLE persons (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "first_name" TEXT,
        "last_name" TEXT);`
	query, err := pdb.db.Prepare(persons_table)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
	fmt.Println("Table created successfully!")
}

func (pdb PersonDB) getPersons() []person {

	rows, _ := pdb.db.Query("SELECT first_name, last_name FROM persons")

	defer rows.Close()

	err := rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	people := make([]person, 0)

	for rows.Next() {
		ourPerson := person{}
		err = rows.Scan(&ourPerson.first_name, &ourPerson.last_name)
		if err != nil {
			log.Fatal(err)
		}

		people = append(people, ourPerson)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return people
}

func (pdb PersonDB) addPerson(newPerson person) (bool, error) {

	tx, err := pdb.db.Begin()
	if err != nil {
		return false, err
	}

	stmt, err := tx.Prepare("INSERT INTO persons (first_name, last_name) VALUES (?, ?)")

	if err != nil {
		return false, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(newPerson.first_name, newPerson.last_name)

	if err != nil {
		return false, err
	}

	tx.Commit()

	return true, nil
}

func main() {

	const dbFile = "database.db"

	file, err := os.Create(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	database, _ := sql.Open("sqlite3", dbFile)
	pdb := PersonDB{database}
	pdb.createTable()
	added, _ := pdb.addPerson(person{"Joe", "Fool"})
	if added {
		fmt.Println("Person added successfully")
		jf := pdb.getPersons()[0]
		fmt.Println(jf.first_name + " " + jf.last_name)
	}
}
