package main

import (
	"github.com/bandar-monitors/monitors/sites/core"
	_ "github.com/mattn/go-sqlite3"
)

const monitorSlug = "target"

func main() {
	core.EnvProductsDelim = ','
	core.GenericBootstrapMonitorWithDefaults(monitorSlug, CreateTargetStatusFetcherFactory(), func() []*core.ProductUrlSpec {
		return core.EnvProductUrlsMapped("SKUS", func(sku string) string {
			return "https://redsky.target.com/redsky_aggregations/v1/apps/pdp_v2?tcin=" + sku + "&store_id=1234&pricing_store_id=1234&scheduled_delivery_store_id=1234&device_type=android&key=5d546952f5059b16db4aa90913e56d09d3ff2aa4"
		})
	})
}
