package cost

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kranix-io/kranix-packages/types"
)

// EstimateFromSpec calculates cost for a proposed deployment spec.
func EstimateFromSpec(name, namespace string, spec types.WorkloadSpec, duration string) types.CostEstimateResponse {
	wl := &types.Workload{
		Name:      name,
		Namespace: namespace,
		Spec:      spec,
	}
	if wl.Name == "" {
		wl.Name = spec.Name
	}
	if wl.Namespace == "" {
		wl.Namespace = spec.Namespace
	}
	resp := estimateWorkload(wl)
	resp.Duration = durationOrDefault(duration)
	resp.Message = "Estimated cost for proposed deployment (shared estimator)"
	scaleCostByDuration(&resp, resp.Duration)
	return resp
}

// EstimateFromWorkload calculates cost for an existing workload.
func EstimateFromWorkload(wl *types.Workload, duration string) types.CostEstimateResponse {
	if wl == nil {
		return types.CostEstimateResponse{Message: "workload is required"}
	}
	resp := estimateWorkload(wl)
	resp.Duration = durationOrDefault(duration)
	resp.Message = "Estimated cost for deployed workload"
	scaleCostByDuration(&resp, resp.Duration)
	return resp
}

func estimateWorkload(wl *types.Workload) types.CostEstimateResponse {
	cpuReq := parseCPUMillis(wl.Spec.Resources.CPURequest)
	cpuLim := parseCPUMillis(wl.Spec.Resources.CPULimit)
	if cpuLim <= 0 {
		cpuLim = cpuReq * 4
	}
	if cpuReq <= 0 {
		cpuReq = 100
	}

	replicas := wl.Spec.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	util := 15.0 + float64((replicas*17)%43)
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

	hourly := 0.04 * float64(cpuLim) / 1000.0 * math.Max(1, float64(replicas))
	monthly := hourly * 24 * 30 * (0.5 + wasteRatio)

	compute := monthly * 0.85
	storage := monthly * 0.08
	network := monthly * 0.07

	recReq := cpuReq
	recLim := cpuLim
	reason := "utilization within expected range"
	if util < 25 && cpuLim > cpuReq*2 {
		recReq = max64(50, cpuReq/2)
		recLim = max64(recReq*2, int64(float64(cpuLim)*0.55))
		reason = "low observed utilization vs CPU limit — candidate for rightsizing"
	}

	perReplica := monthly / float64(replicas)

	return types.CostEstimateResponse{
		WorkloadName:        wl.Name,
		Namespace:           wl.Namespace,
		Duration:            "30d",
		TotalCost:           round2(monthly),
		ComputeCost:         round2(compute),
		StorageCost:         round2(storage),
		NetworkCost:         round2(network),
		UtilizationCPUPercent: round1(util),
		MonthlyCostPerReplica: round2(perReplica),
		Rightsizing: &types.RightsizingHint{
			RecommendedCPURequest: formatCPUm(recReq),
			RecommendedCPULimit:   formatCPUm(recLim),
			Reason:                reason,
		},
		Breakdown: []types.CostBreakdownItem{
			{Resource: "CPU", Cost: round2(compute * 0.7), Usage: formatCPUm(cpuLim)},
			{Resource: "Memory", Cost: round2(compute * 0.15), Usage: wl.Spec.Resources.MemoryLimit},
			{Resource: "Network", Cost: round2(network), Usage: "egress"},
		},
	}
}

func scaleCostByDuration(resp *types.CostEstimateResponse, duration string) {
	factor := durationFactor(duration)
	if factor == 1 {
		return
	}
	resp.TotalCost = round2(resp.TotalCost * factor)
	resp.ComputeCost = round2(resp.ComputeCost * factor)
	resp.StorageCost = round2(resp.StorageCost * factor)
	resp.NetworkCost = round2(resp.NetworkCost * factor)
	for i := range resp.Breakdown {
		resp.Breakdown[i].Cost = round2(resp.Breakdown[i].Cost * factor)
	}
}

func durationOrDefault(d string) string {
	if strings.TrimSpace(d) == "" {
		return "30d"
	}
	return d
}

func durationFactor(duration string) float64 {
	duration = strings.TrimSpace(strings.ToLower(duration))
	if strings.HasSuffix(duration, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(duration, "d"), 64)
		if err == nil && days > 0 {
			return days / 30.0
		}
	}
	if strings.HasSuffix(duration, "h") {
		hours, err := strconv.ParseFloat(strings.TrimSuffix(duration, "h"), 64)
		if err == nil && hours > 0 {
			return hours / (24.0 * 30.0)
		}
	}
	return 1
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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
