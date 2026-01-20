//go:build load

package load

import (
	"testing"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"subscription-aggregator-service/internal/api/controllers"
	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/service"
	"subscription-aggregator-service/internal/storage"
	"subscription-aggregator-service/tests/testutils"
)

type LoadTestResult struct {
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	TotalDuration  time.Duration
	AvgLatency     time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	RequestsPerSec float64
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
}

func (r LoadTestResult) String() string {
	return fmt.Sprintf(`
Load Test Results:
  Total Requests:    %d
  Successful:        %d
  Errors:            %d
  Duration:          %v
  Requests/sec:      %.2f
  Avg Latency:       %v
  Min Latency:       %v
  Max Latency:       %v
  P50 Latency:       %v
  P95 Latency:       %v
  P99 Latency:       %v
`,
		r.TotalRequests, r.SuccessCount, r.ErrorCount,
		r.TotalDuration, r.RequestsPerSec,
		r.AvgLatency, r.MinLatency, r.MaxLatency,
		r.P50Latency, r.P95Latency, r.P99Latency)
}

// TestLoad_CreateSubscriptions tests concurrent subscription creation
func TestLoad_CreateSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	container, err := testutils.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	err = container.RunMigrations(ctx)
	require.NoError(t, err)

	st := storage.NewSubscriptionsStorage(container.DB)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)

	router := gin.New()
	router.POST("/subscriptions", ctrl.CreateSubscription)

	server := httptest.NewServer(router)
	defer server.Close()

	// Load test parameters
	concurrency := 10
	requestsPerWorker := 100
	totalRequests := concurrency * requestsPerWorker

	var (
		successCount int64
		errorCount   int64
		latencies    = make([]time.Duration, 0, totalRequests)
		latencyMu    sync.Mutex
	)

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				req := apiModels.CreateSubscriptionRequest{
					ServiceName: fmt.Sprintf("Service-%d-%d", workerID, j),
					Price:       100 + j,
					UserID:      uuid.New().String(),
					StartDate:   "01-2024",
				}
				body, _ := json.Marshal(req)

				reqStart := time.Now()
				resp, err := http.Post(server.URL+"/subscriptions", "application/json", bytes.NewBuffer(body))
				latency := time.Since(reqStart)

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()

				if err != nil || resp.StatusCode != http.StatusCreated {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				if resp != nil {
					resp.Body.Close()
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	result := calculateResults(latencies, successCount, errorCount, totalDuration)
	t.Log(result.String())

	// Assertions
	errorRate := float64(errorCount) / float64(totalRequests) * 100
	if errorRate > 1 {
		t.Errorf("Error rate too high: %.2f%% (expected < 1%%)", errorRate)
	}

	if result.P95Latency > 500*time.Millisecond {
		t.Errorf("P95 latency too high: %v (expected < 500ms)", result.P95Latency)
	}
}

// TestLoad_ListSubscriptions tests concurrent list operations
func TestLoad_ListSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	container, err := testutils.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	err = container.RunMigrations(ctx)
	require.NoError(t, err)

	st := storage.NewSubscriptionsStorage(container.DB)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)

	router := gin.New()
	router.POST("/subscriptions", ctrl.CreateSubscription)
	router.GET("/subscriptions", ctrl.ListSubscriptions)

	server := httptest.NewServer(router)
	defer server.Close()

	// Pre-populate with data
	userID := uuid.New().String()
	for i := 0; i < 100; i++ {
		req := apiModels.CreateSubscriptionRequest{
			ServiceName: fmt.Sprintf("Service-%d", i),
			Price:       100 + i,
			UserID:      userID,
			StartDate:   "01-2024",
		}
		body, _ := json.Marshal(req)
		resp, _ := http.Post(server.URL+"/subscriptions", "application/json", bytes.NewBuffer(body))
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Load test list endpoint
	concurrency := 20
	requestsPerWorker := 50
	totalRequests := concurrency * requestsPerWorker

	var (
		successCount int64
		errorCount   int64
		latencies    = make([]time.Duration, 0, totalRequests)
		latencyMu    sync.Mutex
	)

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()
				resp, err := http.Get(server.URL + "/subscriptions?user_id=" + userID)
				latency := time.Since(reqStart)

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()

				if err != nil || resp.StatusCode != http.StatusOK {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	result := calculateResults(latencies, successCount, errorCount, totalDuration)
	t.Log(result.String())

	// Assertions
	errorRate := float64(errorCount) / float64(totalRequests) * 100
	if errorRate > 1 {
		t.Errorf("Error rate too high: %.2f%% (expected < 1%%)", errorRate)
	}
}

// TestLoad_MixedOperations tests a mix of CRUD operations
func TestLoad_MixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	container, err := testutils.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	err = container.RunMigrations(ctx)
	require.NoError(t, err)

	st := storage.NewSubscriptionsStorage(container.DB)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)

	router := gin.New()
	router.POST("/subscriptions", ctrl.CreateSubscription)
	router.GET("/subscriptions/:id", ctrl.GetSubscriptionByID)
	router.GET("/subscriptions", ctrl.ListSubscriptions)
	router.GET("/subscriptions/total", ctrl.TotalSubscriptionsCost)

	server := httptest.NewServer(router)
	defer server.Close()

	// Pre-populate
	userID := uuid.New().String()
	var createdIDs []string
	var idsMu sync.Mutex

	for i := 0; i < 50; i++ {
		req := apiModels.CreateSubscriptionRequest{
			ServiceName: fmt.Sprintf("Service-%d", i),
			Price:       100,
			UserID:      userID,
			StartDate:   "01-2024",
			EndDate:     strPtr("12-2024"),
		}
		body, _ := json.Marshal(req)
		resp, _ := http.Post(server.URL+"/subscriptions", "application/json", bytes.NewBuffer(body))
		if resp != nil && resp.StatusCode == http.StatusCreated {
			var sub struct {
				ID string `json:"id"`
			}
			json.NewDecoder(resp.Body).Decode(&sub)
			createdIDs = append(createdIDs, sub.ID)
			resp.Body.Close()
		}
	}

	// Mixed operations load test
	concurrency := 10
	duration := 10 * time.Second

	var (
		successCount int64
		errorCount   int64
		latencies    = make([]time.Duration, 0, 10000)
		latencyMu    sync.Mutex
	)

	start := time.Now()
	done := make(chan bool)

	// Start workers
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			client := &http.Client{Timeout: 5 * time.Second}

			for {
				select {
				case <-done:
					return
				default:
					var reqStart time.Time
					var resp *http.Response
					var err error

					// Random operation mix: 40% list, 30% get, 20% create, 10% total
					op := workerID % 10
					switch {
					case op < 4: // List
						reqStart = time.Now()
						resp, err = client.Get(server.URL + "/subscriptions?user_id=" + userID)

					case op < 7: // Get
						if len(createdIDs) > 0 {
							idsMu.Lock()
							id := createdIDs[workerID%len(createdIDs)]
							idsMu.Unlock()
							reqStart = time.Now()
							resp, err = client.Get(server.URL + "/subscriptions/" + id)
						}

					case op < 9: // Create
						req := apiModels.CreateSubscriptionRequest{
							ServiceName: fmt.Sprintf("LoadTest-%d", time.Now().UnixNano()),
							Price:       100,
							UserID:      userID,
							StartDate:   "01-2024",
						}
						body, _ := json.Marshal(req)
						reqStart = time.Now()
						resp, err = client.Post(server.URL+"/subscriptions", "application/json", bytes.NewBuffer(body))

					default: // Total cost
						reqStart = time.Now()
						resp, err = client.Get(fmt.Sprintf("%s/subscriptions/total?user_id=%s&start_date=01-2024&end_date=12-2024",
							server.URL, userID))
					}

					if !reqStart.IsZero() {
						latency := time.Since(reqStart)
						latencyMu.Lock()
						latencies = append(latencies, latency)
						latencyMu.Unlock()

						if err != nil || (resp != nil && resp.StatusCode >= 400) {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}

					if resp != nil {
						resp.Body.Close()
					}
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(done)
	time.Sleep(100 * time.Millisecond) // Let workers finish

	totalDuration := time.Since(start)
	result := calculateResults(latencies, successCount, errorCount, totalDuration)
	t.Log(result.String())

	// Assertions
	totalRequests := successCount + errorCount
	if totalRequests == 0 {
		t.Error("No requests were made")
	}

	errorRate := float64(errorCount) / float64(totalRequests) * 100
	if errorRate > 5 {
		t.Errorf("Error rate too high: %.2f%% (expected < 5%%)", errorRate)
	}
}

func calculateResults(latencies []time.Duration, successCount, errorCount int64, totalDuration time.Duration) LoadTestResult {
	if len(latencies) == 0 {
		return LoadTestResult{}
	}

	// Sort for percentiles
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	sortDurations(sortedLatencies)

	var totalLatency time.Duration
	minLatency := sortedLatencies[0]
	maxLatency := sortedLatencies[0]

	for _, l := range sortedLatencies {
		totalLatency += l
		if l < minLatency {
			minLatency = l
		}
		if l > maxLatency {
			maxLatency = l
		}
	}

	totalRequests := int64(len(latencies))
	avgLatency := totalLatency / time.Duration(totalRequests)

	return LoadTestResult{
		TotalRequests:  totalRequests,
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		TotalDuration:  totalDuration,
		AvgLatency:     avgLatency,
		MinLatency:     minLatency,
		MaxLatency:     maxLatency,
		RequestsPerSec: float64(totalRequests) / totalDuration.Seconds(),
		P50Latency:     percentile(sortedLatencies, 50),
		P95Latency:     percentile(sortedLatencies, 95),
		P99Latency:     percentile(sortedLatencies, 99),
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) - 1) * p / 100
	return sorted[idx]
}

func sortDurations(d []time.Duration) {
	for i := 0; i < len(d); i++ {
		for j := i + 1; j < len(d); j++ {
			if d[j] < d[i] {
				d[i], d[j] = d[j], d[i]
			}
		}
	}
}

func strPtr(s string) *string {
	return &s
}
