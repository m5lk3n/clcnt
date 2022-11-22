package main

import (
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"lttl.dev/clcnt/models"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var reg *models.Registry

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

func getEntries(c *gin.Context) {
	entries, err := reg.GetEntries()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
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

	err = reg.AddEntry(models.Entry{Timestamp: timestamp, Food: food, Calories: entryCalories})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entry added"})
}

func options(c *gin.Context) {
	o := "HTTP/1.1 200 OK\n" +
		"Allow: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Origin: http://locahost:8080\n" +
		"Access-Control-Allow-Methods: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Headers: Content-Type\n"

	c.String(200, o)
}

func main() {
	var err error
	reg, err = models.NewRegistry()
	checkErr(err)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", getEntries)
		v1.POST("entry/:food/:calories/*timestamp", addEntry)
		v1.OPTIONS("entry", options)
	}

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	r.Run()
}
