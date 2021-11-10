package core

import (
	"context"
	"fmt"
	"github.com/bandar-monitors/monitors/core/domain/stats"
	"go.elastic.co/apm"
	"sync"
	"time"
)

func CreateWatchersManager(slug string, productRepo ProductsRepo, publisher ProductStatusChangePublisher,
	watcherFactory WatcherFactory, collector MonitorStatsCollector) WatchersManager {
	//ctx, cancelFunc := context.WithCancel(context.Background())
	manager := Manager{
		watchers:       []Watcher{},
		repo:           productRepo,
		publisher:      publisher,
		monitorSlug:    slug,
		watcherFactory: watcherFactory,
		mu:             &sync.Mutex{},
		collector:      collector,
		//ctx:                ctx,
		//cancel:             cancelFunc,
	}
	return &manager
}

type Manager struct {
	watchers       []Watcher
	repo           ProductsRepo
	publisher      ProductStatusChangePublisher
	watcherFactory WatcherFactory
	mu             *sync.Mutex
	monitorSlug    string
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
	collector      MonitorStatsCollector
	//ctx                context.Context
}

func (m *Manager) SpawnWatcher(ctx context.Context, instancesCount int, productUrl string, status ProductStatus) {
	rootSpan, _ := apm.StartSpan(ctx, "spawn_watcher", m.monitorSlug)
	rootSpan.Context.SetTag("status", fmt.Sprintf("%d", status))
	rootSpan.Context.SetTag("url", productUrl)
	defer rootSpan.End()

	targets := m.watcherFactory.GetTargetsFromUrl(productUrl)
	statuses := map[string]ProductStatus{}
	for _, target := range targets {
		ref, _ := m.repo.GetRef(ctx, target)
		if ref == nil {
			m.repo.Add(ctx, target, status)
		} else {
			status = ref.status
		}

		statuses[target] = status
	}

	watcher := m.watcherFactory.CreateWatcher(productUrl, statuses, m.handleTargetStatusChange, m.wg, ctx)
	m.watchers = append(m.watchers, watcher)

	watcher.Spawn(instancesCount)
}

/*func (m *Manager) handleStatusChange(ctx context.Context, watcher Watcher, prevStatus ProductStatus,
	currStatus ProductStatus) error {

	rootSpan, _ := apm.StartSpan(ctx, "status_change_handle", m.monitorSlug)
	defer rootSpan.End()
	rootSpan.Context.SetTag("curr_status", fmt.Sprintf("%d", currStatus))
	rootSpan.Context.SetTag("prev_status", fmt.Sprintf("%d", prevStatus))

	sendStatsSpan, _ := apm.StartSpan(ctx, "stats", m.monitorSlug)
	statsPayload := &stats.MonitorWatcherStats{
		ProductUrl: watcher.GetProductUrl(),
		ProductPic: watcher.GetProductPic(),
		Title: watcher.GetProductTitle(),
		//Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	m.collector.UpdateWatcherEntry(statsPayload)
	sendStatsSpan.End()

	if prevStatus == currStatus {
		rootSpan.Context.SetTag("nothing_change", "true")
		rootSpan.End()

		delaySpan, _ := apm.StartSpan(ctx, "delay", m.monitorSlug)
		defer delaySpan.End()
		if currStatus == Available {
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	areStatusChanged, err := m.repo.AreStatusChanged(ctx, watcher.GetProductUrl(), currStatus)
	if !areStatusChanged {
		rootSpan.Context.SetTag("stored_with_same_status", "true")
		return nil
	}

	processSpan, _ := apm.StartSpan(ctx, "processing", m.monitorSlug)
	defer processSpan.End()
	err = m.publisher.Publish(ctx, watcher.GetProductUrl(), currStatus, watcher.GetWebhookPayload)
	if err != nil {
		e := apm.CaptureError(ctx, err)
		e.SetSpan(processSpan)
		e.Send()
		fmt.Println("Can't publish payload. reason: " + err.Error())
		return err
	}

	err = m.repo.ChangeStatus(ctx, watcher.GetProductUrl(), currStatus)

	return err
}*/

func (m *Manager) handleTargetStatusChange(ctx context.Context, target WatchTarget, prevStatus ProductStatus, currStatus ProductStatus) error {
	rootSpan, _ := apm.StartSpan(ctx, "status_change_handle", m.monitorSlug)
	defer rootSpan.End()
	rootSpan.Context.SetTag("curr_status", fmt.Sprintf("%d", currStatus))
	rootSpan.Context.SetTag("prev_status", fmt.Sprintf("%d", prevStatus))

	sendStatsSpan, _ := apm.StartSpan(ctx, "stats", m.monitorSlug)
	statsPayload := &stats.MonitorWatcherStats{
		ProductUrl: target.GetTargetId(),
		ProductPic: target.GetProductPic(),
		Title:      target.GetProductTitle(),
		//Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	m.collector.UpdateWatcherEntry(statsPayload)
	sendStatsSpan.End()

	if prevStatus == currStatus {
		rootSpan.Context.SetTag("nothing_change", "true")
		rootSpan.End()

		delaySpan, _ := apm.StartSpan(ctx, "delay", m.monitorSlug)
		defer delaySpan.End()
		if currStatus == Available {
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	areStatusChanged, err := m.repo.AreStatusChanged(ctx, target.GetTargetId(), currStatus)
	if !areStatusChanged {
		rootSpan.Context.SetTag("stored_with_same_status", "true")
		return nil
	}

	processSpan, _ := apm.StartSpan(ctx, "processing", m.monitorSlug)
	defer processSpan.End()
	processSpan.Context.SetLabel("title", target.GetProductTitle())
	processSpan.Context.SetLabel("id", target.GetTargetId())
	processSpan.Context.SetLabel("pic", target.GetProductPic())
	processSpan.Context.SetLabel("url", target.GetProductUrl())
	processSpan.Context.SetLabel("is_available", target.Available())
	err = m.publisher.Publish(ctx, target.GetTargetId(), currStatus, target.GetWebhookPayload())
	if err != nil {
		e := apm.CaptureError(ctx, err)
		e.SetSpan(processSpan)
		e.Send()
		fmt.Println("Can't publish payload. reason: " + err.Error())
		return err
	}

	err = m.repo.ChangeStatus(ctx, target.GetTargetId(), currStatus)

	return err
}

func (m *Manager) Dispose() {
	for _, w := range m.watchers {
		w.Dispose()
	}
}

type WatchersManager interface {
	SpawnWatcher(ctx context.Context, instancesCount int, productUrl string, status ProductStatus)
	Dispose()
}
