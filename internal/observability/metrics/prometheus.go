package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

var (
	mu       sync.RWMutex
	counters = map[string]float64{}
)

func Init() error { return nil }

func Inc(name string, labels map[string]string) {
	mu.Lock()
	defer mu.Unlock()
	key := metricKey(name, labels)
	counters[key] += 1
}

func Add(name string, labels map[string]string, value float64) {
	mu.Lock()
	defer mu.Unlock()
	key := metricKey(name, labels)
	counters[key] += value
}

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(Snapshot()))
}

func Snapshot() string {
	mu.RLock()
	defer mu.RUnlock()

	keys := make([]string, 0, len(counters))
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("# TYPE seeding_events_total counter\n")
	b.WriteString("# TYPE seeding_quality_filter_rejected_total counter\n")
	b.WriteString("# TYPE seeding_messages_published_total counter\n")
	b.WriteString("# TYPE llm_requests_total counter\n")
	b.WriteString("# TYPE seeding_shadowban_skipped_total counter\n")

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s %g\n", k, counters[k]))
	}
	return b.String()
}

func metricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", sanitize(k), sanitize(v)))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%s{%s}", sanitize(name), strings.Join(parts, ","))
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
