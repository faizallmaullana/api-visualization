package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
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

// dbConnections caches database connections for each month.
var dbConnections = make(map[string]*gorm.DB)
var dbMutex = &sync.Mutex{}

// getDbForMonth retrieves or creates a database connection for a specific month.
// The month format should be "YYYY-MM".
func getDbForMonth(month string) (*gorm.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Check cache first
	if db, ok := dbConnections[month]; ok {
		return db, nil
	}

	// Create database directory if it doesn't exist
	dbPath := fmt.Sprintf("database/%s.db", month)
	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open new connection
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM DB for month %s: %w", month, err)
	}

	// Auto migrate schema
	if err := gdb.AutoMigrate(&RequestRecord{}); err != nil {
		return nil, fmt.Errorf("AutoMigrate failed for month %s: %w", month, err)
	}

	// Cache the connection
	dbConnections[month] = gdb
	log.Printf("Successfully connected to and migrated database: %s", dbPath)

	return gdb, nil
}

// getMonthsInRange calculates the list of months (YYYY-MM) a given time range spans.
func getMonthsInRange(start, end time.Time) []string {
	monthsMap := make(map[string]struct{})

	// Normalize start to the beginning of its month
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())

	// Iterate from the start month until we pass the end month
	for !current.After(end) {
		monthStr := current.Format("2006-01")
		monthsMap[monthStr] = struct{}{}
		current = current.AddDate(0, 1, 0)
	}

	// Convert map keys to slice
	result := make([]string, 0, len(monthsMap))
	for month := range monthsMap {
		result = append(result, month)
	}
	return result
}

func main() {
	// Create the main database directory on startup.
	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory on startup: %v", err)
	}

	// Statistics in-memory (this part remains the same)
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

		// Get current month and the corresponding database connection
		currentMonth := time.Now().Format("2006-01")
		gdb, err := getDbForMonth(currentMonth)
		if err != nil {
			log.Printf("Failed to get DB for month %s: %v", currentMonth, err)
			c.String(500, "Internal server error: could not access database")
			return
		}

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
			log.Printf("GORM insert error into %s.db: %v", currentMonth, err)
		}
		c.String(200, "Recorded")
	})

	router.GET("/api/stats", func(c *gin.Context) {
		stats := statsManager.GetStats()
		c.JSON(200, stats)
	})

	router.GET("/api/history", func(c *gin.Context) {
		var startTime, endTime time.Time
		now := time.Now()
		var timeRangeSpecified bool

		// Priority 1: Specific date range from query parameters
		startDateStr := c.Query("startDate")
		endDateStr := c.Query("endDate")
		if startDateStr != "" && endDateStr != "" {
			start, err1 := time.Parse("2006-01-02", startDateStr)
			end, err2 := time.Parse("2006-01-02", endDateStr)

			if err1 == nil && err2 == nil {
				startTime = start
				// Set time to the end of the day for the end date
				endTime = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				timeRangeSpecified = true
			}
		}

		// Priority 2: Relative time range (default behavior)
		if !timeRangeSpecified {
			timeRange := c.Query("timeRange")
			dur, err := time.ParseDuration(timeRange)
			if err != nil || dur <= 0 {
				dur = time.Hour // Default to 1 hour
			}
			// Set a reasonable max duration to prevent abuse, e.g., 30 days
			maxDur := time.Hour * 24 * 30
			if dur > maxDur {
				dur = maxDur
			}
			endTime = now
			startTime = now.Add(-dur)
		}

		// Get all months that the date range spans
		monthsToQuery := getMonthsInRange(startTime, endTime)
		var allResults []VisualizationData

		// Query each relevant monthly database
		for _, month := range monthsToQuery {
			gdb, err := getDbForMonth(month)
			if err != nil {
				log.Printf("Could not get DB for month %s, skipping: %v", month, err)
				continue
			}

			var recs []RequestRecord
			q := gdb.Model(&RequestRecord{}).Where("timestamp BETWEEN ? AND ?", startTime, endTime)

			// Apply other filters
			if ep := c.Query("endpoint"); ep != "" && ep != "all" {
				q = q.Where("endpoint = ?", ep)
			}
			if m := c.Query("method"); m != "" && m != "all" {
				q = q.Where("method = ?", m)
			}

			if err := q.Order("timestamp desc").Find(&recs).Error; err != nil {
				log.Printf("DB query error for month %s: %v", month, err)
				continue
			}

			// Map and append results
			for _, r := range recs {
				allResults = append(allResults, VisualizationData{
					Endpoint:     r.Endpoint,
					Method:       r.Method,
					StatusCode:   r.StatusCode,
					ResponseTime: r.ResponseTime,
					Error:        r.Error,
					Timestamp:    r.Timestamp,
				})
			}
		}

		// Sort final combined results by timestamp descending
		sort.Slice(allResults, func(i, j int) bool {
			return allResults[i].Timestamp.After(allResults[j].Timestamp)
		})

		c.JSON(200, allResults)
	})

	// Serve SPA index.html for root
	router.GET("/", func(c *gin.Context) {
		c.File("app/dist/index.html")
	})
	// Fallback for any other route (excluding API)
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "Not Found"})
		} else {
			c.File("app/dist/index.html")
		}
	})

	log.Println("Starting server at http://localhost:8088")
	router.Run(":8088")
}
