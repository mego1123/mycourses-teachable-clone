package models

import (
	"github.com/google/uuid"
	"time"

)

// Node status constants
type NodeStatus string

const (
	NodeStatusActive NodeStatus = "active"
	NodeStatusStale  NodeStatus = "stale"
)

// SystemNode represents a registered server instance.
type SystemNode struct {
	ID        uuid.UUID `json:"id"`
	MachineID string             `json:"machineId"`
	Hostname  string             `json:"hostname"`
	Status    NodeStatus         `json:"status"`
	StartedAt time.Time          `json:"startedAt"`
	LastSeen  time.Time          `json:"lastSeen"`
	Version   string             `json:"version"`
	GoVersion string             `json:"goVersion"`
}

// SystemMetric represents a point-in-time metrics snapshot from a single node.
type SystemMetric struct {
	ID        uuid.UUID `json:"id"`
	NodeID    string             `json:"nodeId"`
	Timestamp time.Time          `json:"timestamp"`
	CPU       CPUMetrics         `json:"cpu"`
	Memory    MemoryMetrics      `json:"memory"`
	Disk      DiskMetrics        `json:"disk"`
	Network   NetworkMetrics     `json:"network"`
	HTTP      HTTPMetrics        `json:"http"`
	Mongo     MongoMetrics       `json:"mongo"`
	GoRuntime    GoRuntimeMetrics        `json:"goRuntime"`
	Integrations IntegrationCountMetrics `json:"integrations"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usagePercent"`
	NumCPU       int     `json:"numCpu"`
}

type MemoryMetrics struct {
	UsedBytes   uint64  `json:"usedBytes"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskMetrics struct {
	UsedBytes   uint64  `json:"usedBytes"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type NetworkMetrics struct {
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
}

type HTTPMetrics struct {
	RequestCount int64            `json:"requestCount"`
	LatencyP50   float64          `json:"latencyP50"`
	LatencyP95   float64          `json:"latencyP95"`
	LatencyP99   float64          `json:"latencyP99"`
	StatusCodes  map[string]int64 `json:"statusCodes"`
	ErrorRate4xx float64          `json:"errorRate4xx"`
	ErrorRate5xx float64          `json:"errorRate5xx"`
}

type MongoMetrics struct {
	CurrentConnections   int32            `json:"currentConnections"`
	AvailableConnections int32            `json:"availableConnections"`
	DataSizeBytes        int64            `json:"dataSizeBytes"`
	IndexSizeBytes       int64            `json:"indexSizeBytes"`
	Collections          int32            `json:"collections"`
	OpCounters           map[string]int64 `json:"opCounters"`
}

type GoRuntimeMetrics struct {
	NumGoroutine int    `json:"numGoroutine"`
	HeapAlloc    uint64 `json:"heapAlloc"`
	HeapSys      uint64 `json:"heapSys"`
	GCPauseNs    uint64 `json:"gcPauseNs"`
	NumGC        uint32 `json:"numGC"`
}

type IntegrationCountMetrics struct {
	StripeAPICalls  int64 `json:"stripeApiCalls"`
	ResendEmails    int64 `json:"resendEmails"`
	DataDogAPICalls int64 `json:"datadogApiCalls"`
}

// Integration health check types (in-memory only, no BSON persistence)

type IntegrationStatus string

const (
	IntegrationHealthy       IntegrationStatus = "healthy"
	IntegrationUnhealthy     IntegrationStatus = "unhealthy"
	IntegrationNotConfigured IntegrationStatus = "not_configured"
)

type IntegrationCheck struct {
	Name       string            `json:"name"`
	Status     IntegrationStatus `json:"status"`
	Message    string            `json:"message"`
	LastCheck  time.Time         `json:"lastCheck"`
	ResponseMs int64             `json:"responseMs"`
	Calls24h   int64             `json:"calls24h"`
}
