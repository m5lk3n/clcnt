package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// https://github.com/golang/go/issues/31859#issuecomment-489889428
var _ = func() bool {
	testing.Init()
	return true
}()

const API_V1 = "/api/v1"

type Response struct {
	AvgCalories int    `json:"avg_calories"`
	Days        string `json:"days"`
}

func TestReadiness(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "ready"))
}

func TestEmptyEntries(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", API_V1+"/entry", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"entries\":[]}", w.Body.String())
}

func TestEmptyCalories(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", API_V1+"/calories", nil)
	r.ServeHTTP(w, req)
	resp := Response{}
	json.Unmarshal([]byte(w.Body.String()), &resp)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, resp.AvgCalories)
	assert.Equal(t, "1", resp.Days)
}

func TestAddEntries(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", API_V1+"/entry/Breakfast/500/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", API_V1+"/entry/Dinner/500/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", API_V1+"/entry/Lunch/500/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", API_V1+"/calories", nil)
	r.ServeHTTP(w, req)

	resp := Response{}
	json.Unmarshal([]byte(w.Body.String()), &resp)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 1500, resp.AvgCalories)
	assert.Equal(t, "1", resp.Days)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", API_V1+"/calories?days=5", nil)
	r.ServeHTTP(w, req)

	json.Unmarshal([]byte(w.Body.String()), &resp)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 300, resp.AvgCalories)
	assert.Equal(t, "5", resp.Days)
}
