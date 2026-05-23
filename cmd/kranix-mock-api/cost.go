package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/kranix-io/kranix-packages/cost"
	"github.com/kranix-io/kranix-packages/types"
)

func (s *mockServer) routeCost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if r.Method == http.MethodGet && path == "/api/v1/cost/summary" {
		s.costSummary(w, r)
		return
	}
	if r.Method == http.MethodPost && path == "/api/v1/cost/estimate" {
		s.estimateDeploymentCost(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *mockServer) estimateDeploymentCost(w http.ResponseWriter, r *http.Request) {
	var req types.CostEstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	name := req.Name
	if name == "" {
		name = req.Spec.Name
	}
	namespace := req.Namespace
	if namespace == "" {
		namespace = req.Spec.Namespace
	}
	resp := cost.EstimateFromSpec(name, namespace, req.Spec, req.Duration)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *mockServer) costSummary(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	s.mu.RLock()
	defer s.mu.RUnlock()
	var top []map[string]interface{}
	var total float64
	count := 0
	for id, wl := range s.workloads {
		if ns != "" && wl.Namespace != ns {
			continue
		}
		count++
		cost := s.estimateCostUnlocked(wl)
		total += cost["total_cost"].(float64)
		top = append(top, map[string]interface{}{
			"workload_id":   id,
			"workload_name": wl.Name,
			"namespace":     wl.Namespace,
			"total_cost":    cost["total_cost"],
		})
	}
	if len(top) > 5 {
		top = top[:5]
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"namespace":          ns,
		"total_cost":         math.Round(total*100) / 100,
		"workload_count":     count,
		"average_cost":       safeAvg(total, count),
		"top_cost_workloads": top,
		"message":            "mock cost rollup for kranix-examples cost agent",
	})
}

func safeAvg(total float64, n int) float64 {
	if n == 0 {
		return 0
	}
	return math.Round(total/float64(n)*100) / 100
}

// estimateCostUnlocked returns mock cost + utilization; caller must hold s.mu (RLock or Lock).
func (s *mockServer) estimateCostUnlocked(wl *types.Workload) map[string]interface{} {
	cpuReq := parseCPUMillis(wl.Spec.Resources.CPURequest)
	cpuLim := parseCPUMillis(wl.Spec.Resources.CPULimit)
	if cpuLim <= 0 {
		cpuLim = cpuReq * 4
	}
	if cpuReq <= 0 {
		cpuReq = 100
	}
	util := 15.0 + float64((wl.Spec.Replicas*17)%43)
	if wl.Labels != nil {
		if v := wl.Labels["kranix.io/demo-cpu-utilization"]; v != "" {
			if u, err := strconv.ParseFloat(v, 64); err == nil {
				util = u
			}
		}
	}
	wasteRatio := 1.0 - (util / 100.0)
	if wasteRatio < 0 {
		wasteRatio = 0
	}
	hourly := 0.04 * float64(cpuLim) / 1000.0 * math.Max(1, float64(wl.Spec.Replicas))
	total := hourly * 24 * 30 * (0.5 + wasteRatio)

	compute := total * 0.85
	storage := total * 0.08
	network := total * 0.07

	recReq := cpuReq
	recLim := cpuLim
	reason := "utilization within expected range"
	if util < 25 && cpuLim > cpuReq*2 {
		recReq = max64(50, cpuReq/2)
		recLim = max64(recReq*2, int64(float64(cpuLim)*0.55))
		reason = "low observed utilization vs CPU limit — candidate for rightsizing"
	}

	return map[string]interface{}{
		"workload_name":           wl.Name,
		"namespace":               wl.Namespace,
		"total_cost":              math.Round(total*100) / 100,
		"compute_cost":            math.Round(compute*100) / 100,
		"storage_cost":            math.Round(storage*100) / 100,
		"network_cost":            math.Round(network*100) / 100,
		"utilization_cpu_percent": math.Round(util*10) / 10,
		"rightsizing": map[string]interface{}{
			"recommended_cpu_request": formatCPUm(recReq),
			"recommended_cpu_limit":   formatCPUm(recLim),
			"reason":                  reason,
		},
		"breakdown": []map[string]interface{}{
			{"resource": "CPU", "cost": math.Round(compute * 0.7), "usage": formatCPUm(cpuLim)},
			{"resource": "Memory", "cost": math.Round(compute * 0.15), "usage": wl.Spec.Resources.MemoryLimit},
		},
	}
}

func (s *mockServer) workloadCost(w http.ResponseWriter, r *http.Request, id string) {
	s.mu.RLock()
	wl, ok := s.workloads[id]
	if !ok {
		s.mu.RUnlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	cost := s.estimateCostUnlocked(wl)
	s.mu.RUnlock()
	out := map[string]interface{}{
		"workload_id": id,
		"duration":    r.URL.Query().Get("duration"),
	}
	for k, v := range cost {
		out[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func parseCPUMillis(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "m") {
		v, err := strconv.ParseInt(strings.TrimSuffix(s, "m"), 10, 64)
		if err != nil {
			return 0
		}
		return v
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(v * 1000)
}

func formatCPUm(m int64) string {
	if m <= 0 {
		return "0"
	}
	if m < 1000 {
		return fmt.Sprintf("%dm", m)
	}
	if m%1000 == 0 {
		return fmt.Sprintf("%d", m/1000)
	}
	return fmt.Sprintf("%dm", m)
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
