// Gin (https://github.com/gin-gonic/gin) handlers
package handlers

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"lttl.dev/clcnt/models"
	"lttl.dev/clcnt/time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var reg *models.Registry

func init() {
	var err error
	reg, err = models.NewRegistry()
	if err != nil {
		log.Fatal(err)
	}
}

// GetEntriesHandler see Description
//
//	@Summary      handler to retrieve all entries
//	@Description  returns all food entries with calories and timestamp, there's neither filtering nor pagination
//	@Accept       json
//	@Produce      json
//	@Success      200
//	@Failure      400
//	@Router       /api/v1/entry [get]
func GetEntriesHandler(c *gin.Context) {
	entries, err := reg.GetEntries()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

// AddEntryHandler see Description
//
//	@Summary      handler to add an entry with calories
//	@Description  adds an entry with the given food and calories with the current Unix timestamp
//	@Accept       json
//	@Produce      json
//	@Param        food			path  string  true  "single word that describes the food (allowed characters are URL path entries/SQLite strings)"
//	@Param        calories	path	int			true	"amount of calories"
//	@Success      200
//	@Failure      400
//	@Router       /api/v1/entry/{food}/{calories} [post]
func AddEntryHandler(c *gin.Context) {
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

// EntryOptionsHandler shows available access methods for /api/v1/entry
func EntryOptionsHandler(c *gin.Context) {
	o := "HTTP/1.1 200 OK\n" +
		"Allow: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Origin: http[s]://<host>[:<port>]\n" +
		"Access-Control-Allow-Methods: GET,POST,OPTIONS\n" +
		"Access-Control-Allow-Headers: Content-Type\n"

	c.String(200, o)
}

// GetCaloriesHandler see Description
//
//	@Summary      handler to retrieve calories
//	@Description  returns today's sum of calories (if 1 day was specific) or the average for the amount of given days
//	@Accept       json
//	@Produce      json
//	@Param        days    query     int  false  "recent day(s) to retrieve calories for"
//	@Success      200
//	@Failure      400
//	@Router       /api/v1/calories [get]
func GetCaloriesHandler(c *gin.Context) {
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

// CaloriesOptionsHandler shows available access methods for /api/v1/calories
func CaloriesOptionsHandler(c *gin.Context) {
	o := "HTTP/1.1 200 OK\n" +
		"Allow: GET,OPTIONS\n" +
		"Access-Control-Allow-Origin: http[s]://<host>[:port]\n" +
		"Access-Control-Allow-Methods: GET,OPTIONS\n" +
		"Access-Control-Allow-Headers: Content-Type\n"

	c.String(200, o)
}

// NotFoundHandler indicates that requested resource could not be found
func NotFoundHandler(c *gin.Context) {
	// log this event as it could be an attempt to break in...
	log.Infoln("Not found, requested URL path:", c.Request.URL.Path)
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "requested resource not found", "status": http.StatusNotFound})
}

// LivenessHandler see Description
//
//	@Summary      liveness handler
//	@Description  always indicates that the API server is alive, consider using readiness handler instead
//	@Accept       json
//	@Produce      json
//	@Success      200
//	@Router       /healthy [get]
func LivenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "alive", "status": http.StatusOK})
}

// ReadinessHandler see Description
//
//	@Summary      readiness handler
//	@Description  indicates whether or not the API server and its database are ready
//	@Accept       json
//	@Produce      json
//	@Success      200
//	@Failure      503
//	@Router       /ready [get]
func ReadinessHandler(c *gin.Context) {
	if reg != nil && reg.IsReady() {
		c.JSON(http.StatusOK, gin.H{"message": "ready", "status": http.StatusOK})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unavailable", "status": http.StatusServiceUnavailable})
	}
}

// IndexHandler provides the landing page
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "clcnt", "header": "clcnt"})
}
