package core

import (
	"context"
	"sync"
)

type StatusChangeHandler = func(ctx context.Context, target WatchTarget, prevStatus ProductStatus, currStatus ProductStatus) error
type WatcherFactory interface {
	CreateWatcher(productUrl string, statuses map[string]ProductStatus, handler StatusChangeHandler, wg *sync.WaitGroup, ctx context.Context) Watcher
	GetTargetsFromUrl(productUrl string) []string
}

type Watcher interface {
	Dispose()
	Spawn(instancesCount int)
	Stop()
	//GetProductUrl() string
	//GetProductPic() string
	//GetProductTitle() string
	//GetWebhookPayload() WebhookPayload
}
