package config

import "net/url"

type FootlockerMonitorCfg struct {
	url              *url.URL
	publishQueueName string
	publishExchange  string
}

func New(rawUrl string, publishQueueName string, publishExchange string) *FootlockerMonitorCfg {
	cfg := FootlockerMonitorCfg{
		publishExchange:  publishExchange,
		publishQueueName: publishQueueName,
	}
	cfg.SetUrl(rawUrl)

	return &cfg
}

func (cfg *FootlockerMonitorCfg) SetUrl(rawUrl string) {
	cfg.url, _ = url.Parse(rawUrl)
}

func (cfg *FootlockerMonitorCfg) GetUrl() string {
	return cfg.url.String()
}

func (cfg *FootlockerMonitorCfg) GetHost() string {
	return cfg.url.Hostname()
}

func (cfg *FootlockerMonitorCfg) GetPublishExchange() string {
	return cfg.publishExchange
}

func (cfg *FootlockerMonitorCfg) GetPublishQueueName() string {
	return cfg.publishQueueName
}
