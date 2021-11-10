package core

import (
	"context"
	"net/http"
)

//
//type TargetProductStatus struct {
//	ProductTitle string
//	ProductUrl string
//	ProductPic string
//	Available  bool
//	Status     ProductStatus
//	Payload    WebhookPayload
//}

type StatusFetchResult struct {
	Targets     []WatchTarget
	StatusCode  int
	RawResponse []byte
	Error       error
	Request     *http.Request
	Response    *http.Response
	Client      *http.Client
}

type watchTarget struct {
	productTitle string
	productId    string
	productUrl   string
	productPic   string
	available    bool
	payload      WebhookPayload
}

func (w *watchTarget) Available() bool {
	return w.available
}

func (w *watchTarget) GetProductUrl() string {
	return w.productUrl
}

func (w *watchTarget) GetProductPic() string {
	return w.productPic
}

func (w *watchTarget) GetProductTitle() string {
	return w.productTitle
}

func (w *watchTarget) GetTargetId() string {
	return w.productId
}

func (w *watchTarget) GetWebhookPayload() WebhookPayload {
	return w.payload
}

func (f *StatusFetchResult) AddTarget(url, title, pic string, available bool, payload WebhookPayload) {
	f.AddTargetWithId(url, url, title, pic, available, payload)
}

func (f *StatusFetchResult) AddTargetWithId(id, url, title, pic string, available bool, payload WebhookPayload) {
	f.Targets = append(f.Targets, &watchTarget{
		productTitle: title,
		productId:    id,
		productUrl:   url,
		productPic:   pic,
		available:    available,
		payload:      payload,
	})
}

type ProductStatusFetcher interface {
	FetchStatus(ctx context.Context) *StatusFetchResult
}
