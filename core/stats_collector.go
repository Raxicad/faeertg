package core

import (
	"context"
	"github.com/bandar-monitors/monitors/core/domain/stats"
	jsoniter "github.com/json-iterator/go"
	"go.elastic.co/apm"
	"sync"
	"time"
)

type statsCollector struct {
	syncDelay   time.Duration
	mu          *sync.Mutex
	entries     map[string]*stats.MonitorWatcherStats
	monitorSlug string
	publisher   StatsPublisher
}

func (s *statsCollector) UpdateWatcherEntry(e *stats.MonitorWatcherStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.ProductUrl] = e
}

func StartCollector(monitorSlug string, publisher StatsPublisher) MonitorStatsCollector {
	e := &statsCollector{
		syncDelay:   time.Second,
		mu:          &sync.Mutex{},
		entries:     map[string]*stats.MonitorWatcherStats{},
		monitorSlug: monitorSlug,
		publisher:   publisher,
	}

	go e.startSyncWorker()

	return e
}

func (s *statsCollector) startSyncWorker() {
	for true {
		if len(s.entries) > 0 {
			tx := apm.DefaultTracer.StartTransaction("publish_stats", s.monitorSlug)
			ctx := apm.ContextWithTransaction(context.Background(), tx)

			s.mu.Lock()
			entries := make([]*stats.MonitorWatcherStats, 0, len(s.entries))
			for _, e := range s.entries {
				entries = append(entries, e)
			}

			s.mu.Unlock()

			entriesJson, err := jsoniter.Marshal(entries)
			if err != nil {
				e := apm.CaptureError(ctx, err)
				e.SetTransaction(tx)
				e.Send()
			}

			payload := &stats.ComponentStats{
				Stats:         string(entriesJson),
				ComponentType: "monitor",
				ComponentName: s.monitorSlug,
				Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			}

			err = s.publisher(ctx, payload)
			if err != nil {
				e := apm.CaptureError(ctx, err)
				e.SetTransaction(tx)
				e.Send()
			}
		}

		time.Sleep(s.syncDelay)
	}
}

type MonitorStatsCollector interface {
	UpdateWatcherEntry(e *stats.MonitorWatcherStats)
}
