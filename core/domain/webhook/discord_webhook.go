package webhook

import "time"

type DiscordWebhookPayload struct {
	Content   string         `json:"content"`
	Embeds    []DiscordEmbed `json:"embeds"`
	Username  string         `json:"username"`
	AvatarUrl string         `json:"avatar_url"`
}
type DiscordEmbed struct {
	Title     string           `json:"title,omitempty"`
	Url       string           `json:"url"`
	Color     int              `json:"color,omitempty"`
	Fields    []DiscordField   `json:"fields,omitempty"`
	Author    DiscordAuthor    `json:"author,omitempty"`
	Footer    DiscordFooter    `json:"footer,omitempty"`
	Timestamp time.Time        `json:"timestamp,omitempty"`
	Image     DiscordImage     `json:"image"`
	Thumbnail DiscordThumbnail `json:"thumbnail,omitempty"`
}
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}
type DiscordAuthor struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	IconUrl string `json:"icon_url"`
}
type DiscordFooter struct {
	Text    string `json:"text"`
	IconUrl string `json:"icon_url"`
}
type DiscordImage struct {
	Url string `json:"url"`
}
type DiscordThumbnail struct {
	Url string `json:"url"`
}
