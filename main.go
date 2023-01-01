package main

import (
	"embed"
	"html/template"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func getCaloriesHandler(c *gin.Context) {
	d := c.DefaultQuery("days", "1")

	days, err := strconv.Atoi(d)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal parameter"})
		return
	}

	if days < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal parameter"})
		return
	}

	t := getDaysAgoAsUnix(days)

	calories, err := reg.GetCalories(t)
	if err != nil {
		log.Infof(":::%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	avg := calories / days

	c.JSON(http.StatusOK, gin.H{"days": d, "avg_calories": avg})
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
	if reg != nil && reg.IsReady() {
		c.JSON(http.StatusOK, gin.H{"message": "ready", "status": http.StatusOK})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unavailable", "status": http.StatusServiceUnavailable})
	}
}

// indexHandler provides the landing page
func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "clcnt", "header": "clcnt"})
}

//go:embed assets/* templates/*
var f embed.FS

// SetupRouter is published here to allow setup of tests
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil) // https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies

	// https://github.com/gin-gonic/gin#build-a-single-binary-with-templates
	t := template.Must(template.New("").ParseFS(f, "templates/*.tmpl"))
	r.SetHTMLTemplate(t) // do not use the following with SetHTMLTemplate: r.LoadHTMLGlob("templates/*")
	// conflicts with getting /favicon.ico: r.StaticFS("/", http.FS(f))

	r.GET("favicon.ico", func(c *gin.Context) {
		file, _ := f.ReadFile("assets/favicon.ico")
		c.Data(
			http.StatusOK,
			"image/x-icon",
			file,
		)
	})

	// to debug: r.Use(gindump.Dump())

	r.Use(gin.Recovery()) // "recover from any panics", write 500 if any

	r.NoRoute(notFoundHandler)

	// generic API
	r.GET("/healthy", livenessHandler)
	r.GET("/ready", readinessHandler)
	r.GET("/", indexHandler)

	// specific API
	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", getEntriesHandler)
		v1.POST("entry/:food/:calories/*timestamp", addEntryHandler)
		v1.OPTIONS("entry", entryOptionsHandler)
		v1.GET("calories", getCaloriesHandler)
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

// TODO: move to class
func getStartOfTodayAsUnix() int64 {
	t := time.Now()
	m := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return m.Unix()
}

// TODO: move to class
func getDaysAgoAsUnix(d int) int64 {
	t := getStartOfTodayAsUnix()

	return t - int64((d-1)*24*60*60)
}
