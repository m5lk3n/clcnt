package main

import (
	"embed"
	"flag"
	"html/template"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"lttl.dev/clcnt/models"
	"lttl.dev/clcnt/time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var reg *models.Registry
var debugFlag *bool

func init() {
	debugFlag = flag.Bool("debug", false, "turn debug output on/off")

	flag.Parse()

	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
	timestamp := time.DefaultTimestamp(c.Param("timestamp")) // contains at least leading / due to redirect

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
		"Access-Control-Allow-Origin: http[s]://<host>[:<port>]\n" +
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

	t := time.GetDaysAgoAsUnix(days)

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
		"Access-Control-Allow-Origin: http[s]://<host>[:port]\n" +
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
	if !*debugFlag {
		gin.SetMode(gin.ReleaseMode)
	}

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
