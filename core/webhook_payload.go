package core

type WebhookPayload interface {
	ToJSON() []byte
}
