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

func getEntriesHandler(c *gin.Context) {
	entries, err := reg.GetEntries()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func addEntryHandler(c *gin.Context) {
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

func entryOptionsHandler(c *gin.Context) {
	o := "HTTP/1.1 200 OK\n" +
		"Allow: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Origin: http://locahost:8080\n" + // TODO
		"Access-Control-Allow-Methods: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Headers: Content-Type\n"

	c.String(200, o)
}

func getTodaysCaloriesHandler(c *gin.Context) {
	calories, err := reg.GetTodaysCalories()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"timespan": "today", "calories": calories})
}

func caloriesOptionsHandler(c *gin.Context) {
	o := "HTTP/1.1 200 OK\n" +
		"Allow: GET,OPTIONS\n" +
		"Access-Control-Allow-Origin: http://locahost:8080\n" + // TODO
		"Access-Control-Allow-Methods: GET,OPTIONS\n" +
		"Access-Control-Allow-Headers: Content-Type\n"

	c.String(200, o)
}

// NotFoundHandler to indicate that requested resource could not be found
func notFoundHandler(c *gin.Context) {
	// log this event as it could be an attempt to break in...
	log.Infoln("Not found, requested URL path:", c.Request.URL.Path)
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "requested resource not found", "status": http.StatusNotFound})
}

// LivenessHandler always returns HTTP 200, consider using ReadinessHandler instead
func livenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "alive", "status": http.StatusOK})
}

// ReadinessHandler indicates HTTP 200 if ready, otherwise HTTP 503
func readinessHandler(c *gin.Context) {
	if true { // TODO
		c.JSON(http.StatusOK, gin.H{"message": "ready", "status": http.StatusOK})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unavailable", "status": http.StatusServiceUnavailable})
	}
}

// SetupRouter is published here to allow setup of tests
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// to debug: r.Use(gindump.Dump())

	r.Use(gin.Recovery()) // "recover from any panics", write 500 if any

	// r.Use(static.Serve("/", static.LocalFile("./static", true)))

	r.NoRoute(notFoundHandler)

	// generic API
	r.GET("/healthy", livenessHandler)
	r.GET("/ready", readinessHandler)

	// specific API
	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", getEntriesHandler)
		v1.POST("entry/:food/:calories/*timestamp", addEntryHandler)
		v1.OPTIONS("entry", entryOptionsHandler)
		v1.GET("calories/24", getTodaysCaloriesHandler) // TODO: 24 ok?
		v1.OPTIONS("calories", caloriesOptionsHandler)
	}

	return r
}

func main() {
	var err error
	reg, err = models.NewRegistry()
	checkErr(err)

	r := SetupRouter()

	log.Info("clcnt server start...")
	defer log.Info("clcnt server shutdown!")

	// set port via PORT environment variable
	r.Run() // default port is 8080
}
