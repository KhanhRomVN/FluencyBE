package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
)

type Metrics struct {
	logger     *logger.PrettyLogger
	webhookURL string
	cache      cache.Cache
}

type RedisMetrics struct {
	HitRate          float64
	MissRate         float64
	LatencyMs        float64
	MemoryUsageMB    float64
	EvictionRate     float64
	ConnectedClients int64
	BlockedClients   int64
	OpsPerSecond     float64
	NetworkInput     float64
	NetworkOutput    float64
	UsedMemoryPeak   int64
}

func (m *Metrics) collectRedisMetrics() *RedisMetrics {
	if m.cache == nil {
		return nil
	}

	ctx := context.Background()
	info, err := m.cache.Info(ctx).Result()
	if err != nil {
		m.logger.Error("REDIS_METRICS_COLLECTION", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to collect Redis metrics")
		return nil
	}

	metrics := &RedisMetrics{}

	// Parse Redis INFO output
	infoLines := strings.Split(info, "\r\n")
	for _, line := range infoLines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "used_memory_peak":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.UsedMemoryPeak = v
			}
		case "instantaneous_ops_per_sec":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				metrics.OpsPerSecond = v
			}
		case "instantaneous_input_kbps":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				metrics.NetworkInput = v
			}
		case "instantaneous_output_kbps":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				metrics.NetworkOutput = v
			}
		case "connected_clients":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.ConnectedClients = v
			}
		case "blocked_clients":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.BlockedClients = v
			}
		}
	}

	// Get additional metrics from cache client
	cacheMetrics := m.cache.GetMetrics()
	if cacheMetrics != nil {
		metrics.HitRate = float64(cacheMetrics.Hits) / float64(cacheMetrics.Hits+cacheMetrics.Misses) * 100
		metrics.MissRate = float64(cacheMetrics.Misses) / float64(cacheMetrics.Hits+cacheMetrics.Misses) * 100
		metrics.LatencyMs = cacheMetrics.LatencyMs
		metrics.MemoryUsageMB = float64(cacheMetrics.MemoryUsage) / 1024 / 1024
	}

	return metrics
}

func (m *Metrics) getRedisMetricFields(metrics *RedisMetrics) []map[string]interface{} {
	if metrics == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"name":   "📊 Redis Hit Rate",
			"value":  fmt.Sprintf("%.2f%%", metrics.HitRate),
			"inline": true,
		},
		{
			"name":   "📊 Redis Miss Rate",
			"value":  fmt.Sprintf("%.2f%%", metrics.MissRate),
			"inline": true,
		},
		{
			"name":   "⏱️ Redis Latency",
			"value":  fmt.Sprintf("%.2fms", metrics.LatencyMs),
			"inline": true,
		},
		{
			"name":   "💾 Redis Memory",
			"value":  fmt.Sprintf("%.2f MB", metrics.MemoryUsageMB),
			"inline": true,
		},
		{
			"name":   "👥 Redis Clients",
			"value":  fmt.Sprintf("%d", metrics.ConnectedClients),
			"inline": true,
		},
		{
			"name":   "⚡ Redis Ops/sec",
			"value":  fmt.Sprintf("%.2f", metrics.OpsPerSecond),
			"inline": true,
		},
		{
			"name":   "🌐 Redis Network",
			"value":  fmt.Sprintf("%.2f KB/%.2f KB", metrics.NetworkInput, metrics.NetworkOutput),
			"inline": true,
		},
		{
			"name":   "📈 Redis Peak Mem",
			"value":  fmt.Sprintf("%d MB", metrics.UsedMemoryPeak/1024/1024),
			"inline": true,
		},
	}
}

type SystemMetrics struct {
	Timestamp          string  `json:"timestamp"`
	CPUUsage           float64 `json:"cpu_usage"`
	MemoryUsage        uint64  `json:"memory_usage"`
	GoRoutines         int     `json:"goroutines"`
	ThreadCount        int     `json:"thread_count"`
	HeapAlloc          uint64  `json:"heap_alloc"`
	HeapObjects        uint64  `json:"heap_objects"`
	GarbageCollections uint32  `json:"garbage_collections"`
	Uptime             string  `json:"uptime"`
}

var startTime = time.Now()

func NewMetrics(logger *logger.PrettyLogger, cache cache.Cache) *Metrics {
	return &Metrics{
		logger:     logger,
		webhookURL: os.Getenv("DISCORD_METRIC_WEBHOOK_URL"),
		cache:      cache,
	}
}

func (m *Metrics) StartMetricsCollection() {
	// Log initial start
	m.logger.Info("METRICS_START", map[string]interface{}{
		"interval": "10s",
	}, "Starting metrics collection")

	// Collect metrics immediately on start
	m.collectAndSendMetrics()

	// Then start the ticker for subsequent collections
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			m.collectAndSendMetrics()
		}
	}()
}

func (m *Metrics) collectAndSendMetrics() {
	// Check if metrics should be sent to Discord
	if os.Getenv("IS_METRIC_BACKEND_TO_DISCORD") != "YES" {
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(startTime).Round(time.Second)

	metrics := SystemMetrics{
		Timestamp:          time.Now().Format("2006-01-02 15:04:05"),
		CPUUsage:           float64(runtime.NumCPU()),
		MemoryUsage:        memStats.Alloc,
		GoRoutines:         runtime.NumGoroutine(),
		ThreadCount:        runtime.GOMAXPROCS(0),
		HeapAlloc:          memStats.HeapAlloc,
		HeapObjects:        memStats.HeapObjects,
		GarbageCollections: memStats.NumGC,
		Uptime:             uptime.String(),
	}

	// Get Redis metrics
	var redisMetrics *RedisMetrics
	if m.cache != nil {
		redisMetrics = m.collectRedisMetrics()
	}

	embed := map[string]interface{}{
		"title":       "🖥️ System & Redis Metrics Report",
		"description": fmt.Sprintf("Server Uptime: **%s**", metrics.Uptime),
		"color":       0x00ff00,
		"fields": append([]map[string]interface{}{
			{
				"name":   "💻 CPU Usage",
				"value":  fmt.Sprintf("%.2f%%", metrics.CPUUsage),
				"inline": true,
			},
			{
				"name":   "💾 Memory Usage",
				"value":  fmt.Sprintf("%.2f MB", float64(metrics.MemoryUsage)/1024/1024),
				"inline": true,
			},
			{
				"name":   "🔄 Goroutines",
				"value":  fmt.Sprintf("%d", metrics.GoRoutines),
				"inline": true,
			},
			{
				"name":   "🧵 Thread Count",
				"value":  fmt.Sprintf("%d", metrics.ThreadCount),
				"inline": true,
			},
			{
				"name":   "📊 Heap Allocation",
				"value":  fmt.Sprintf("%.2f MB", float64(metrics.HeapAlloc)/1024/1024),
				"inline": true,
			},
			{
				"name":   "🔹 Heap Objects",
				"value":  fmt.Sprintf("%d", metrics.HeapObjects),
				"inline": true,
			},
			{
				"name":   "♻️ GC Cycles",
				"value":  fmt.Sprintf("%d", metrics.GarbageCollections),
				"inline": true,
			},
		}, m.getRedisMetricFields(redisMetrics)...),
		"footer": map[string]interface{}{
			"text": fmt.Sprintf("Last Updated: %s", metrics.Timestamp),
		},
	}

	payload := map[string]interface{}{
		"embeds": []interface{}{embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		m.logger.Error("METRICS_MARSHAL", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to marshal metrics data")
		return
	}

	resp, err := http.Post(m.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		m.logger.Error("METRICS_SEND", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to send metrics to Discord")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		m.logger.Error("METRICS_RESPONSE", map[string]interface{}{
			"status_code": resp.StatusCode,
		}, "Unexpected response status from Discord webhook")
	}
}
