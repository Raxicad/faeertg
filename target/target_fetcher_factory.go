package main

import (
	"github.com/bandar-monitors/monitors/sites/core"
)

type targetFetcherFactory struct {
	httpClientFactory core.HttpClientFactory
}

func (t *targetFetcherFactory) SetHttpClientFactory(factory core.HttpClientFactory) {
	t.httpClientFactory = factory
}

func CreateTargetStatusFetcherFactory() core.ProductStatusFetcherFactory {
	return &targetFetcherFactory{}
}

func (t *targetFetcherFactory) CreateFetcher(productUrl string) core.ProductStatusFetcher {
	return &targetFetcher{
		http:       t.httpClientFactory.CreateClient(),
		productUrl: productUrl,
	}
}
