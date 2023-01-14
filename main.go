package main

import (
	"embed"
	"flag"
	"html/template"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	log "github.com/sirupsen/logrus"

	"lttl.dev/clcnt/docs"
	"lttl.dev/clcnt/handlers"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var debugFlag *bool

func init() {
	debugFlag = flag.Bool("debug", false, "turn debug output on/off")

	flag.Parse()

	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}
}

//go:embed assets/* templates/*
var f embed.FS

func setupRouter() *gin.Engine {
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

	r.NoRoute(handlers.NotFoundHandler)

	// generic API
	r.GET("/healthy", handlers.LivenessHandler)
	r.GET("/ready", handlers.ReadinessHandler)
	r.GET("/", handlers.IndexHandler)

	// specific API
	v1 := r.Group("/api/v1")
	{
		v1.GET("entry", handlers.GetEntriesHandler)
		v1.POST("entry/:food/:calories/*timestamp", handlers.AddEntryHandler)
		v1.OPTIONS("entry", handlers.EntryOptionsHandler)
		v1.GET("calories", handlers.GetCaloriesHandler)
		v1.OPTIONS("calories", handlers.CaloriesOptionsHandler)
	}

	// API docs
	docs.SwaggerInfo.Schemes = []string{"http"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	return r
}

// @title           clcnt API
// @version         1.0
// @description     This is the clcnt API server.

// @contact.name   clcnt API Support
// @contact.url    https://github.com/m5lk3n/clcnt

// @license.name  MIT
// @license.url   https://github.com/m5lk3n/clcnt/blob/main/LICENSE

// @host      localhost:8080
func main() {
	r := setupRouter()

	log.Info("clcnt server start...")
	defer log.Info("clcnt server shutdown!")

	// set port via PORT environment variable
	r.Run() // default port is 8080
}
