package db

import (
	"context"
	"database/sql"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmsql"
	_ "go.elastic.co/apm/module/apmsql/pq"
)

func EstablishDbConnection(connectionString string, ctx context.Context) (*sql.DB, error) {
	var db *sql.DB
	var err error
	db, err = connect(connectionString, ctx)
	return db, err
}

func connect(connectionString string, ctx context.Context) (*sql.DB, error) {
	db, err := apmsql.Open("postgres", connectionString)

	createDbStatement, err := db.PrepareContext(ctx, "CREATE TABLE IF NOT EXISTS public.subscriptions (id SERIAL PRIMARY KEY, slug VARCHAR(255) NOT NULL, discord_webhook_url VARCHAR(4096) NOT NULL)")
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	_, err = createDbStatement.ExecContext(ctx)
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	//createPublishersDbStatement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS public.publishers (id INT PRIMARY KEY, url VARCHAR(255) NOT NULL")
	//createPublishersDbStatement.Exec()

	return db, err
}
