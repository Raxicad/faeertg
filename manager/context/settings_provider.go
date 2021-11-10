package context

import (
	"context"
	stats "github.com/bandar-monitors/monitors/core/domain/stats"
	"github.com/bandar-monitors/monitors/core/domain/webhook"
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.elastic.co/apm"
	"time"
)

func RefreshSettings(ctx *ManagerContext, context context.Context) error {
	tx := apm.DefaultTracer.StartTransaction("refresh_settings", "balancer")
	traceCtx := apm.ContextWithTransaction(context, tx)
	defer tx.End()
	settingsCache := make(map[string][]string)
	entries := []*balancerSubscriptionEntry{}
	rows, err := ctx.Database.QueryContext(traceCtx, "SELECT slug, discord_webhook_url from public.subscriptions")
	if err != nil {
		e := apm.CaptureError(traceCtx, err)
		e.SetTransaction(tx)
		e.Send()
		return err
	}

	for rows.Next() {
		span, _ := apm.StartSpan(traceCtx, "read_rows", "balancer")
		var slug string
		var url string
		_ = rows.Scan(&slug, &url)

		list, ok := settingsCache[slug]
		if !ok {
			list = []string{}
		}

		settingsCache[slug] = append(list, url)
		entries = append(entries, &balancerSubscriptionEntry{
			Slug:       slug,
			WebhookUrl: url,
		})
		span.End()
	}

	entriesJson, err := jsoniter.Marshal(entries)
	if err != nil {
		e := apm.CaptureError(traceCtx, err)
		e.SetTransaction(tx)
		e.Send()
	}
	componentStats := stats.ComponentStats{
		Stats:         string(entriesJson),
		ComponentType: "balancer",
		ComponentName: "balancer",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	json, err := jsoniter.Marshal(componentStats)
	ctx.Bus.Channel.Publish(webhook.ComponentExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		})
	ctx.SettingsCache = &settingsCache
	return nil
}

type balancerSubscriptionEntry struct {
	Slug       string `json:"slug"`
	WebhookUrl string `json:"webhookUrl"`
}
