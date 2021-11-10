package core

type WatchTarget interface {
	Available() bool
	GetProductUrl() string
	GetProductPic() string
	GetProductTitle() string
	GetTargetId() string
	GetWebhookPayload() WebhookPayload
}
