package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// VisualizationData matches the structure from the markdown documentation
type VisualizationData struct {
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ResponseTime float64   `json:"response_time"` // in milliseconds
	Error        string    `json:"error"`
	Timestamp    time.Time `json:"timestamp"`
}

// RouteStats holds aggregated statistics for each route
type RouteStats struct {
	Path         string          `json:"path"`
	Method       string          `json:"method"`
	Count        int             `json:"count"`
	LastAccessed time.Time       `json:"last_accessed"`
	ResponseTime []time.Duration `json:"-"`
	AvgResponse  time.Duration   `json:"avg_response"`
	StatusCodes  map[int]int     `json:"status_codes"`
	Errors       []string        `json:"errors"`
}

// StatsManager manages all route statistics
type StatsManager struct {
	routes map[string]*RouteStats
}

// NewStatsManager creates a new statistics manager
func NewStatsManager() *StatsManager {
	return &StatsManager{
		routes: make(map[string]*RouteStats),
	}
}

// RecordRequest records statistics from the visualization data
func (sm *StatsManager) RecordRequest(data VisualizationData) {
	key := fmt.Sprintf("%s %s", data.Method, data.Endpoint)

	if sm.routes[key] == nil {
		sm.routes[key] = &RouteStats{
			Path:        data.Endpoint,
			Method:      data.Method,
			StatusCodes: make(map[int]int),
			Errors:      []string{},
		}
	}

	stats := sm.routes[key]
	stats.Count++
	stats.LastAccessed = time.Now()

	// Convert response time from ms (float64) to time.Duration
	duration := time.Duration(data.ResponseTime * float64(time.Millisecond))
	stats.ResponseTime = append(stats.ResponseTime, duration)

	stats.StatusCodes[data.StatusCode]++

	if data.Error != "" {
		// Limit stored errors to prevent memory exhaustion
		if len(stats.Errors) < 10 {
			stats.Errors = append(stats.Errors, data.Error)
		}
	}

	// Calculate average response time
	var total time.Duration
	for _, rt := range stats.ResponseTime {
		total += rt
	}
	stats.AvgResponse = total / time.Duration(len(stats.ResponseTime))
}

// GetStats returns all route statistics, sorted by access count
func (sm *StatsManager) GetStats() []*RouteStats {
	var stats []*RouteStats
	for _, stat := range sm.routes {
		statCopy := &RouteStats{
			Path:         stat.Path,
			Method:       stat.Method,
			Count:        stat.Count,
			LastAccessed: stat.LastAccessed,
			AvgResponse:  stat.AvgResponse,
			StatusCodes:  make(map[int]int),
			Errors:       append([]string{}, stat.Errors...),
		}
		for k, v := range stat.StatusCodes {
			statCopy.StatusCodes[k] = v
		}
		stats = append(stats, statCopy)
	}

	// Sort by count (most accessed first)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

// RequestRecord is GORM model for stored requests
type RequestRecord struct {
	ID           uint `gorm:"primaryKey"`
	Endpoint     string
	Method       string
	StatusCode   int
	ResponseTime float64
	Error        string
	Timestamp    time.Time
}

func main() {
	// Initialize GORM SQLite
	gdb, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open GORM DB: %v", err)
	}
	// Auto migrate schema
	if err := gdb.AutoMigrate(&RequestRecord{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// Statistics in-memory
	statsManager := NewStatsManager()

	// Gin router
	router := gin.Default()

	// Serve static assets (JS/CSS) from dist/assets
	router.Static("/assets", "app/dist/assets")

	// API routes
	router.POST("/api/visualize", func(c *gin.Context) {
		var data VisualizationData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.String(400, "Invalid request body")
			return
		}
		statsManager.RecordRequest(data)
		// Insert into DB with GORM
		rec := RequestRecord{
			Endpoint:     data.Endpoint,
			Method:       data.Method,
			StatusCode:   data.StatusCode,
			ResponseTime: data.ResponseTime,
			Error:        data.Error,
			Timestamp:    time.Now(),
		}
		if err := gdb.Create(&rec).Error; err != nil {
			log.Printf("GORM insert error: %v", err)
		}
		c.String(200, "Recorded")
	})

	router.GET("/api/stats", func(c *gin.Context) {
		stats := statsManager.GetStats()
		c.JSON(200, stats)
	})

	router.GET("/api/history", func(c *gin.Context) {
		// Support optional month filter (format YYYY-MM)
		if monthStr := c.Query("month"); monthStr != "" {
			// Parse month and fetch full month range
			t, err := time.Parse("2006-01", monthStr)
			if err == nil {
				start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
				end := start.AddDate(0, 1, 0)
				// GORM query
				var recs []RequestRecord
				if err := gdb.Where("timestamp >= ? AND timestamp < ?", start, end).Find(&recs).Error; err != nil {
					c.String(500, "DB query error")
					return
				}
				// Map to VisualizationData and initialize slice to avoid null
				results := make([]VisualizationData, 0, len(recs))
				for _, r := range recs {
					results = append(results, VisualizationData{
						Endpoint:     r.Endpoint,
						Method:       r.Method,
						StatusCode:   r.StatusCode,
						ResponseTime: r.ResponseTime,
						Error:        r.Error,
						Timestamp:    r.Timestamp,
					})
				}
				c.JSON(200, results)
				return
			}
		}
		// Else use timeRange query, capped to max 1 month
		timeRange := c.Query("timeRange")
		dur, err := time.ParseDuration(timeRange)
		if err != nil || dur <= 0 {
			dur = time.Hour
		}
		maxDur := time.Hour * 24 * 30
		if dur > maxDur {
			dur = maxDur
		}
		cutoff := time.Now().Add(-dur)
		// Build GORM query
		q := gdb.Where("timestamp >= ?", cutoff)
		if ep := c.Query("endpoint"); ep != "" && ep != "all" {
			q = q.Where("endpoint = ?", ep)
		}
		if m := c.Query("method"); m != "" && m != "all" {
			q = q.Where("method = ?", m)
		}
		var recs []RequestRecord
		if err := q.Find(&recs).Error; err != nil {
			c.String(500, "DB query error")
			return
		}
		// Map to VisualizationData and initialize slice
		results := make([]VisualizationData, 0, len(recs))
		for _, r := range recs {
			results = append(results, VisualizationData{
				Endpoint:     r.Endpoint,
				Method:       r.Method,
				StatusCode:   r.StatusCode,
				ResponseTime: r.ResponseTime,
				Error:        r.Error,
				Timestamp:    r.Timestamp,
			})
		}
		c.JSON(200, results)
	})

	// Serve SPA index.html for root
	router.GET("/", func(c *gin.Context) {
		c.File("app/dist/index.html")
	})
	// Fallback for any other route (excluding API) using NoRoute to avoid wildcard conflicts
	router.NoRoute(func(c *gin.Context) {
		// If the path starts with /api/, respond 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "Not Found"})
		} else {
			c.File("app/dist/index.html")
		}
	})

	log.Println("Starting server at http://localhost:8088")
	router.Run(":8088")
}
