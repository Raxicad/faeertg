package core

import (
	"context"
	"sync"
)

type genericWatcherFactory struct {
	httpFactory    HttpClientFactory
	monitorSlug    string
	fetcherFactory ProductStatusFetcherFactory
}

func (wf *genericWatcherFactory) GetTargetsFromUrl(productUrl string) []string {
	targetsFactory, ok := wf.fetcherFactory.(ProductTargetsFactory)
	if !ok {
		return []string{productUrl}
	}

	return targetsFactory.GetTargetsFromUrl(productUrl)
}

func (wf *genericWatcherFactory) CreateWatcher(productUrl string, statuses map[string]ProductStatus, handler StatusChangeHandler,
	wg *sync.WaitGroup, ctx context.Context) Watcher {
	return &genericWatcher{
		statusChangeHandler: handler,
		ctx:                 ctx,
		fetcherFactory:      wf.fetcherFactory,
		monitorSlug:         wf.monitorSlug,
		fetchUrl:            productUrl,
		results:             map[string]*executionResult{},
		statuses:            statuses,
		targetsCount:        len(statuses),
		mu:                  &sync.Mutex{},
	}
}

func CreateGenericWatchersFactory(factory HttpClientFactory, monitorSlug string, fetcherFactory ProductStatusFetcherFactory) WatcherFactory {
	return &genericWatcherFactory{
		httpFactory:    factory,
		fetcherFactory: fetcherFactory,
		monitorSlug:    monitorSlug,
	}
}
