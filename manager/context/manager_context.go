package context

import (
	"context"
	"database/sql"
	"github.com/bandar-monitors/monitors/core/util"
	bus2 "github.com/bandar-monitors/monitors/manager/bus"
	db2 "github.com/bandar-monitors/monitors/manager/db"
	"go.elastic.co/apm"
)

type ManagerContext struct {
	Database      *sql.DB
	Bus           *bus2.DiscordHooksConsumerContext
	SettingsCache *map[string][]string
}

func NewContext(amqpConnStr string, postgresConStr string, runtimeCtx context.Context) *ManagerContext {
	rdbmsSpan, _ := apm.StartSpan(runtimeCtx, "establish_rdms_conn", "balancer")
	dbConnection, err := db2.EstablishDbConnection(postgresConStr, runtimeCtx)
	util.FailOnError(err, "Can't establish db connection")
	rdbmsSpan.End()

	rmqSpan, _ := apm.StartSpan(runtimeCtx, "establish_rmq_conn", "balancer")
	busCtx := bus2.EstablishAmqpConnection(amqpConnStr)
	rmqSpan.End()

	var settingsCache map[string][]string
	ctx := &ManagerContext{
		Database:      dbConnection,
		Bus:           busCtx,
		SettingsCache: &settingsCache,
	}

	_ = RefreshSettings(ctx, runtimeCtx)

	return ctx
}

func (ctx *ManagerContext) Dispose() {
	ctx.Bus.Close()
	err := ctx.Database.Close()
	util.FailOnError(err, "Failed to close database connection")
}
