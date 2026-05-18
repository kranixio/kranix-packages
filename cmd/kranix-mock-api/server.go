package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kranix-io/kranix-packages/types"
)

type mockServer struct {
	skipAuth bool

	mu             sync.RWMutex
	workloads      map[string]*types.Workload
	namespaces     map[string]*types.Namespace
	pods           map[string]*types.Pod
	workloadPods   map[string][]string // workloadID -> pod IDs
	sse            *mockSSE
	nextWorkloadID int
	nextPodID      int
}

func newMockServer(skipAuth bool) *mockServer {
	now := time.Now().UTC()
	ns := map[string]*types.Namespace{
		"default": {
			Name:      "default",
			CreatedAt: now,
			UpdatedAt: now,
			Status: types.NamespaceStatus{
				Phase: types.NamespacePhaseActive,
			},
		},
	}
	return &mockServer{
		skipAuth:   skipAuth,
		workloads:  make(map[string]*types.Workload),
		namespaces: ns,
		pods:       make(map[string]*types.Pod),
		workloadPods: make(map[string][]string),
		sse:        newMockSSE(),
	}
}

func (s *mockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.skipAuth && !checkAuth(r) {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	switch {
	case r.Method == http.MethodGet && path == "/health":
		s.health(w)
		return
	case r.Method == http.MethodGet && path == "/openapi.json":
		s.openapi(w)
		return
	case r.Method == http.MethodGet && path == "/api/sse":
		s.sse.handleConnection(w, r)
		return
	case r.Method == http.MethodGet && path == "/api/sse/stats":
		s.sse.stats(w)
		return
	case r.Method == http.MethodPost && path == "/api/sse/broadcast":
		s.sse.handleBroadcast(w, r)
		return
	}

	if strings.HasPrefix(path, "/api/v1/workloads") {
		s.routeWorkloads(w, r)
		return
	}
	if strings.HasPrefix(path, "/api/v1/pods/") {
		s.routePods(w, r)
		return
	}
	if strings.HasPrefix(path, "/api/v1/namespaces") {
		s.routeNamespaces(w, r)
		return
	}

	http.NotFound(w, r)
}

func checkAuth(r *http.Request) bool {
	h := r.Header.Get("Authorization")
	if h == "" {
		return false
	}
	token := strings.TrimPrefix(h, "Bearer ")
	return strings.HasPrefix(token, "krane_")
}

func (s *mockServer) health(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *mockServer) openapi(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	stub := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]string{
			"title":   "Kranix API (mock)",
			"version": "0.0-mock",
		},
		"paths": map[string]any{
			"/api/v1/workloads":      map[string]string{"description": "see kranix-api /openapi.json for full spec"},
			"/api/v1/namespaces":     map[string]string{},
			"/api/sse":               map[string]string{},
			"/health":                map[string]string{},
		},
	}
	_ = json.NewEncoder(w).Encode(stub)
}

func (s *mockServer) routeWorkloads(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if r.Method == http.MethodPost && strings.HasSuffix(path, "/restart") {
		id := segmentBeforeSuffix(path, "/restart")
		s.restartWorkload(w, r, id)
		return
	}
	switch r.Method {
	case http.MethodPost:
		if path == "/api/v1/workloads" {
			s.deployWorkload(w, r)
			return
		}
	case http.MethodGet:
		if path == "/api/v1/workloads" {
			s.listWorkloads(w, r)
			return
		}
		if strings.HasSuffix(path, "/analyze") {
			id := segmentBeforeSuffix(path, "/analyze")
			s.analyzeWorkload(w, r, id)
			return
		}
		if strings.HasSuffix(path, "/pods") {
			id := segmentBeforeSuffix(path, "/pods")
			s.listWorkloadPods(w, r, id)
			return
		}
		s.getWorkload(w, r, path[strings.LastIndex(path, "/")+1:])
		return
	case http.MethodPatch:
		id := path[strings.LastIndex(path, "/")+1:]
		s.updateWorkload(w, r, id)
		return
	case http.MethodDelete:
		id := path[strings.LastIndex(path, "/")+1:]
		s.deleteWorkload(w, r, id)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func segmentBeforeSuffix(full, suffix string) string {
	full = strings.TrimSuffix(full, "/")
	if !strings.HasSuffix(full, suffix) {
		return ""
	}
	base := strings.TrimSuffix(full, suffix)
	idx := strings.LastIndex(base, "/")
	if idx < 0 || idx == len(base)-1 {
		return ""
	}
	return base[idx+1:]
}

func (s *mockServer) routePods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	if strings.HasSuffix(path, "/logs") {
		base := strings.TrimSuffix(path, "/logs")
		id := base[strings.LastIndex(base, "/")+1:]
		s.streamPodLogs(w, r, id)
		return
	}
	http.NotFound(w, r)
}

func (s *mockServer) routeNamespaces(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch r.Method {
	case http.MethodPost:
		if path == "/api/v1/namespaces" {
			s.createNamespace(w, r)
			return
		}
	case http.MethodGet:
		if path == "/api/v1/namespaces" {
			s.listNamespaces(w, r)
			return
		}
		name := strings.TrimPrefix(path, "/api/v1/namespaces/")
		if name != "" {
			s.getNamespace(w, name)
			return
		}
	case http.MethodDelete:
		name := path[strings.LastIndex(path, "/")+1:]
		s.deleteNamespace(w, r, name)
		return
	}
	http.NotFound(w, r)
}

func (s *mockServer) deployWorkload(w http.ResponseWriter, r *http.Request) {
	var spec types.WorkloadSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ns := spec.Namespace
	if ns == "" {
		ns = "default"
	}
	s.ensureNamespace(ns)

	s.mu.Lock()
	s.nextWorkloadID++
	id := fmt.Sprintf("wl-%d", s.nextWorkloadID)
	now := time.Now().UTC()
	wl := &types.Workload{
		ID:        id,
		Name:      spec.Name,
		Namespace: ns,
		Spec:      spec,
		Status: types.WorkloadStatus{
			Phase:         types.WorkloadPhaseRunning,
			ReadyReplicas: spec.Replicas,
			Message:       "mock",
			LastUpdated:   now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.workloads[id] = wl

	s.nextPodID++
	podID := fmt.Sprintf("pod-%d", s.nextPodID)
	pod := &types.Pod{
		ID:         podID,
		Name:       spec.Name + "-0001",
		Namespace:  ns,
		WorkloadID: id,
		Phase:      types.PodPhaseRunning,
		CreatedAt:  now,
	}
	s.pods[podID] = pod
	s.workloadPods[id] = []string{podID}
	s.mu.Unlock()

	change := &types.WorkloadStateChange{
		WorkloadID: id,
		Namespace:  ns,
		OldState:   "",
		NewState:   string(types.WorkloadPhaseRunning),
		Reason:     "mock deploy",
		ChangedAt:  now,
	}
	s.sse.broadcast("workload.changed", change, map[string]string{"namespace": ns})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(wl)
}

func (s *mockServer) listWorkloads(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.Workload
	for _, wl := range s.workloads {
		if ns == "" || wl.Namespace == ns {
			out = append(out, wl)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"namespace": ns,
		"workloads": out,
	})
}

func (s *mockServer) getWorkload(w http.ResponseWriter, _ *http.Request, id string) {
	s.mu.RLock()
	wl, ok := s.workloads[id]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wl)
}

func (s *mockServer) updateWorkload(w http.ResponseWriter, r *http.Request, id string) {
	var spec types.WorkloadSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	wl, ok := s.workloads[id]
	if !ok {
		s.mu.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	wl.Spec = spec
	wl.UpdatedAt = time.Now().UTC()
	wl.Status.LastUpdated = wl.UpdatedAt
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wl)
}

func (s *mockServer) deleteWorkload(w http.ResponseWriter, _ *http.Request, id string) {
	s.mu.Lock()
	if wl, ok := s.workloads[id]; ok {
		ns := wl.Namespace
		delete(s.workloads, id)
		for _, pid := range s.workloadPods[id] {
			delete(s.pods, pid)
		}
		delete(s.workloadPods, id)
		s.mu.Unlock()
		s.sse.broadcast("workload.deleted", map[string]string{"workloadId": id}, map[string]string{"namespace": ns})
	} else {
		s.mu.Unlock()
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *mockServer) restartWorkload(w http.ResponseWriter, _ *http.Request, id string) {
	s.mu.RLock()
	wl, ok := s.workloads[id]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.sse.broadcast("workload.updated", map[string]string{"workloadId": id, "action": "restart"}, map[string]string{"namespace": wl.Namespace})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "restart accepted", "workloadId": id})
}

func (s *mockServer) analyzeWorkload(w http.ResponseWriter, _ *http.Request, id string) {
	s.mu.RLock()
	_, ok := s.workloads[id]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	res := types.AnalysisResult{
		WorkloadID:  id,
		Status:      "ok",
		ProbableFix: "mock: reduce replica count or increase memory limits",
		AnalyzedAt:  time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *mockServer) listWorkloadPods(w http.ResponseWriter, _ *http.Request, id string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.workloads[id]; !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var list []*types.Pod
	for _, pid := range s.workloadPods[id] {
		if p, ok := s.pods[pid]; ok {
			list = append(list, p)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"pods": list})
}

func (s *mockServer) streamPodLogs(w http.ResponseWriter, r *http.Request, podID string) {
	s.mu.RLock()
	_, ok := s.pods[podID]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ctx := r.Context()
	line := 0
	tick := time.NewTicker(800 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			line++
			msg := fmt.Sprintf("[%s] mock log line %d", podID, line)
			fmt.Fprintf(w, "event: log\n")
			fmt.Fprintf(w, "data: %s\n\n", msg)
			fl.Flush()
			if r.URL.Query().Get("follow") != "true" && line >= 3 {
				return
			}
		}
	}
}

func (s *mockServer) createNamespace(w http.ResponseWriter, r *http.Request) {
	var ns types.Namespace
	if err := json.NewDecoder(r.Body).Decode(&ns); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if ns.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	now := time.Now().UTC()
	ns.CreatedAt = now
	ns.UpdatedAt = now
	if ns.Status.Phase == "" {
		ns.Status = types.NamespaceStatus{Phase: types.NamespacePhaseActive}
	}
	s.mu.Lock()
	s.namespaces[ns.Name] = &ns
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ns)
}

func (s *mockServer) getNamespace(w http.ResponseWriter, name string) {
	s.mu.RLock()
	ns, ok := s.namespaces[name]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ns)
}

func (s *mockServer) listNamespaces(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*types.Namespace
	for _, n := range s.namespaces {
		list = append(list, n)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"namespaces": list})
}

func (s *mockServer) deleteNamespace(w http.ResponseWriter, _ *http.Request, name string) {
	s.mu.Lock()
	delete(s.namespaces, name)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *mockServer) ensureNamespace(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.namespaces[name]; ok {
		return
	}
	now := time.Now().UTC()
	s.namespaces[name] = &types.Namespace{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Status:    types.NamespaceStatus{Phase: types.NamespacePhaseActive},
	}
}

// --- minimal SSE hub (mirrors kranix-api /api/sse behaviour) ---

type mockSSE struct {
	mu      sync.RWMutex
	clients map[string]*mockSSEClient
	reg     chan *mockSSEClient
	unreg   chan *mockSSEClient
	bcast   chan *types.BroadcastMessage
}

type mockSSEClient struct {
	id             string
	subscriptions  map[string]bool
	send           chan *types.SSEEvent
	connectionData map[string]string
}

func newMockSSE() *mockSSE {
	s := &mockSSE{
		clients: make(map[string]*mockSSEClient),
		reg:     make(chan *mockSSEClient, 16),
		unreg:   make(chan *mockSSEClient, 16),
		bcast:   make(chan *types.BroadcastMessage, 256),
	}
	go s.loop()
	return s
}

func (s *mockSSE) loop() {
	for {
		select {
		case c := <-s.reg:
			s.mu.Lock()
			s.clients[c.id] = c
			s.mu.Unlock()
		case c := <-s.unreg:
			s.mu.Lock()
			if _, ok := s.clients[c.id]; ok {
				delete(s.clients, c.id)
				close(c.send)
			}
			s.mu.Unlock()
		case msg := <-s.bcast:
			s.dispatch(msg)
		}
	}
}

func (s *mockSSE) dispatch(msg *types.BroadcastMessage) {
	ev := &types.SSEEvent{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Event:     msg.Event,
		Data:      msg.Data,
		Timestamp: time.Now().UTC(),
		Retry:     3000,
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		if mockShouldSend(c, msg.Filter) {
			select {
			case c.send <- ev:
			default:
			}
		}
	}
}

func mockShouldSend(c *mockSSEClient, filter map[string]string) bool {
	if filter == nil {
		return true
	}
	if ns, ok := filter["namespace"]; ok {
		if !c.subscriptions["*"] && !c.subscriptions[ns] {
			return false
		}
	}
	return true
}

func (s *mockSSE) handleConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("c-%d", time.Now().UnixNano())
	}
	namespaces := r.URL.Query()["namespace"]
	if len(namespaces) == 0 {
		namespaces = []string{"*"}
	}
	cid := fmt.Sprintf("%s-%d", clientID, time.Now().UnixNano())
	client := &mockSSEClient{
		id:            cid,
		subscriptions: map[string]bool{},
		send:          make(chan *types.SSEEvent, 64),
		connectionData: map[string]string{
			"clientID": cid,
		},
	}
	for _, ns := range namespaces {
		client.subscriptions[ns] = true
	}
	s.reg <- client
	client.send <- &types.SSEEvent{
		ID:        "connection",
		Event:     "connected",
		Data:      client.connectionData,
		Timestamp: time.Now().UTC(),
	}

	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "sse unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	go func() {
		<-ctx.Done()
		s.unreg <- client
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-client.send:
			if !ok {
				return
			}
			_ = writeSSE(w, ev)
			fl.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, event *types.SSEEvent) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if event.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", event.ID)
	}
	if event.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", event.Event)
	}
	if event.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", event.Retry)
	}
	fmt.Fprintf(&buf, "data: %s\n\n", data)
	_, err = w.Write(buf.Bytes())
	return err
}

func (s *mockSSE) stats(w http.ResponseWriter) {
	s.mu.RLock()
	n := len(s.clients)
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"connectedClients": n})
}

func (s *mockSSE) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Event  string            `json:"event"`
		Data   any               `json:"data"`
		Filter map[string]string `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	s.bcast <- &types.BroadcastMessage{
		Event:  body.Event,
		Data:   body.Data,
		Filter: body.Filter,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Event broadcasted"})
}

func (s *mockSSE) broadcast(event string, data any, filter map[string]string) {
	select {
	case s.bcast <- &types.BroadcastMessage{Event: event, Data: data, Filter: filter}:
	default:
	}
}
