package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kranix-io/kranix-packages/types"
)

func (s *mockServer) seedDemoRunbooks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := "rb-oncall-pagerduty"
	s.runbooks[id] = map[string]interface{}{
		"id":          id,
		"name":        "PagerDuty service-down playbook",
		"description": "Agent-driven triage: acknowledge alert, pull workload logs, optionally restart",
		"category":    "oncall",
		"steps": []interface{}{
			map[string]interface{}{"order": 1, "name": "ack", "action": "pagerduty.acknowledge"},
			map[string]interface{}{"order": 2, "name": "context", "action": "kranix.describe_workload"},
			map[string]interface{}{"order": 3, "name": "logs", "action": "kranix.tail_logs"},
			map[string]interface{}{"order": 4, "name": "mitigate", "action": "kranix.restart_workload"},
		},
	}
}

func (s *mockServer) routeIncident(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodGet && path == "/api/v1/incident/runbooks":
		s.listRunbooks(w, r)
		return
	case r.Method == http.MethodPost && path == "/api/v1/incident/runbooks":
		s.createRunbook(w, r)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/incident/runbooks/"):
		if strings.HasSuffix(path, "/execute") {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(path, "/api/v1/incident/runbooks/")
		s.getRunbook(w, id)
		return
	case r.Method == http.MethodPost && strings.HasSuffix(path, "/execute"):
		base := strings.TrimSuffix(path, "/execute")
		id := strings.TrimPrefix(base, "/api/v1/incident/runbooks/")
		s.executeRunbook(w, r, id)
		return
	case r.Method == http.MethodGet && path == "/api/v1/incident/executions":
		s.listExecutions(w, r)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/incident/executions/"):
		id := strings.TrimPrefix(path, "/api/v1/incident/executions/")
		s.getExecution(w, id)
		return
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/v1/incident/executions/"):
		id := strings.TrimPrefix(path, "/api/v1/incident/executions/")
		s.cancelExecution(w, id)
		return
	}
	http.NotFound(w, r)
}

func (s *mockServer) listRunbooks(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []map[string]interface{}
	for _, rb := range s.runbooks {
		if category != "" {
			if c, ok := rb["category"].(string); !ok || c != category {
				continue
			}
		}
		list = append(list, cloneMap(rb))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"category": category,
		"runbooks": list,
	})
}

func (s *mockServer) getRunbook(w http.ResponseWriter, id string) {
	s.mu.RLock()
	rb, ok := s.runbooks[id]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cloneMap(rb))
}

func (s *mockServer) createRunbook(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.nextRunbook++
	id := fmt.Sprintf("rb-%d", s.nextRunbook)
	body["id"] = id
	s.runbooks[id] = body
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"runbook": body})
}

func (s *mockServer) executeRunbook(w http.ResponseWriter, r *http.Request, runbookID string) {
	s.mu.RLock()
	_, ok := s.runbooks[runbookID]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "runbook not found", http.StatusNotFound)
		return
	}
	var req map[string]interface{}
	raw, _ := io.ReadAll(r.Body)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
	}
	if req == nil {
		req = map[string]interface{}{}
	}

	s.mu.Lock()
	s.nextExec++
	exID := fmt.Sprintf("ex-%d", s.nextExec)
	now := time.Now().UTC()
	ex := map[string]interface{}{
		"id":         exID,
		"runbookId":  runbookID,
		"status":     "succeeded",
		"startedAt":  now,
		"finishedAt": now,
		"source":     "pagerduty_webhook",
		"input":      req,
		"steps": []interface{}{
			map[string]interface{}{"name": "ack", "status": "done"},
			map[string]interface{}{"name": "triaged_via_kranix_api", "status": "done"},
		},
	}
	s.executions[exID] = ex
	s.mu.Unlock()

	s.sse.broadcast("runbook.executed", map[string]interface{}{
		"executionId": exID,
		"runbookId":   runbookID,
	}, nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"runbook_id": runbookID,
		"execution":  ex,
	})
}

func (s *mockServer) listExecutions(w http.ResponseWriter, r *http.Request) {
	rbFilter := r.URL.Query().Get("runbook_id")
	statusFilter := r.URL.Query().Get("status")
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []map[string]interface{}
	for _, ex := range s.executions {
		if rbFilter != "" && fmt.Sprint(ex["runbookId"]) != rbFilter {
			continue
		}
		if statusFilter != "" && fmt.Sprint(ex["status"]) != statusFilter {
			continue
		}
		list = append(list, cloneMap(ex))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"runbook_id": rbFilter,
		"status":     statusFilter,
		"executions": list,
	})
}

func (s *mockServer) getExecution(w http.ResponseWriter, id string) {
	s.mu.RLock()
	ex, ok := s.executions[id]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cloneMap(ex))
}

func (s *mockServer) cancelExecution(w http.ResponseWriter, id string) {
	s.mu.Lock()
	if ex, ok := s.executions[id]; ok {
		ex["status"] = "cancelled"
		s.executions[id] = ex
	}
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	b, _ := json.Marshal(m)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out
}

// --- analytics (ML latency) ---

func (s *mockServer) routeAnalytics(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodPost && path == "/api/v1/analytics/metrics":
		s.recordMetric(w, r)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/analytics/workloads/"):
		id := strings.TrimPrefix(path, "/api/v1/analytics/workloads/")
		s.workloadMetrics(w, r, id)
		return
	case r.Method == http.MethodGet && path == "/api/v1/analytics/metrics":
		s.queryMetrics(w, r)
		return
	case r.Method == http.MethodGet && path == "/api/v1/analytics/summary":
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "summary",
			"window":  "24h"})
		return
	}
	http.NotFound(w, r)
}

func (s *mockServer) recordMetric(w http.ResponseWriter, r *http.Request) {
	var m types.AnalyticsMetrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if m.ResourceID != "" && (m.MetricType == "latency" || m.Labels["type"] == "inference_latency") {
		s.mu.Lock()
		s.latencySamples[m.ResourceID] = append(s.latencySamples[m.ResourceID], m.Value)
		if len(s.latencySamples[m.ResourceID]) > 500 {
			s.latencySamples[m.ResourceID] = s.latencySamples[m.ResourceID][len(s.latencySamples[m.ResourceID])-500:]
		}
		s.mu.Unlock()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Metric recorded successfully"})
}

func (s *mockServer) workloadMetrics(w http.ResponseWriter, r *http.Request, workloadID string) {
	metricType := r.URL.Query().Get("type")
	s.mu.RLock()
	samples := append([]float64(nil), s.latencySamples[workloadID]...)
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if metricType != "latency" && metricType != "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"workloadId": workloadID,
			"message":    "only type=latency populated in mock",
		})
		return
	}
	p50, p95, p99, avg, max := percentileLatency(samples)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"workloadId": workloadID,
		"latency": map[string]interface{}{
			"p50Ms": p50,
			"p95Ms": p95,
			"p99Ms": p99,
			"avgMs": avg,
			"maxMs": max,
			"samples": len(samples),
		},
	})
}

func (s *mockServer) queryMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": []interface{}{},
		"count":   0,
	})
}

func percentileLatency(samples []float64) (p50, p95, p99, avg, max float64) {
	if len(samples) == 0 {
		return 12, 28, 45, 18, 50
	}
	sorted := append([]float64(nil), samples...)
	sort.Float64s(sorted)
	n := len(sorted)
	max = sorted[n-1]
	var sum float64
	for _, v := range sorted {
		sum += v
	}
	avg = sum / float64(n)
	p50 = sorted[n/2]
	p95i := (n * 95) / 100
	if p95i >= n {
		p95i = n - 1
	}
	p95 = sorted[p95i]
	p99i := (n * 99) / 100
	if p99i >= n {
		p99i = n - 1
	}
	p99 = sorted[p99i]
	return p50, p95, p99, avg, max
}
