package domain

type NotificationPayload struct {
	Slug    string `json:"slug"`
	Payload string `json:"payload"`

	// UTC RFC3339Nano formatted timestamp
	Timestamp string `json:"timestamp"`
}
