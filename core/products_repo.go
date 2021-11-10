package core

import (
	"context"
	"database/sql"
	"fmt"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmsql"
	_ "go.elastic.co/apm/module/apmsql/sqlite3"
	"os"
)

type ProductsRepo interface {
	Add(ctx context.Context, id string, status ProductStatus) error
	ChangeStatus(ctx context.Context, id string, status ProductStatus) error
	AreStatusChanged(ctx context.Context, id string, status ProductStatus) (bool, error)
	GetRef(ctx context.Context, url string) (*ProductRef, error)
}

type productsRepo struct {
	db *sql.DB
}

func (p *productsRepo) GetRef(ctx context.Context, url string) (*ProductRef, error) {
	stmt, err := p.db.PrepareContext(ctx, "SELECT status FROM products WHERE url=$1")
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	row := stmt.QueryRowContext(ctx, url)
	var status ProductStatus
	err = row.Scan(&status)
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	return &ProductRef{
		url:    url,
		status: status,
	}, nil
}

func (p *productsRepo) Add(ctx context.Context, productUrl string, status ProductStatus) error {
	stmt, err := p.db.PrepareContext(ctx, "INSERT INTO products (url, status) VALUES ($1, $2)")
	if err != nil {
		apm.CaptureError(ctx, err)
		return err
	}

	_, err = stmt.ExecContext(ctx, productUrl, status)
	if err != nil {
		apm.CaptureError(ctx, err)
		return err
	}

	return nil
}

func (p *productsRepo) ChangeStatus(ctx context.Context, productUrl string, status ProductStatus) error {
	stmt, err := p.db.PrepareContext(ctx, "UPDATE products SET status=$1 WHERE url=$2")
	if err != nil {
		apm.CaptureError(ctx, err)
		return err
	}

	_, err = stmt.ExecContext(ctx, status, productUrl)
	if err != nil {
		apm.CaptureError(ctx, err)
		return err
	}

	return nil
}

func (p *productsRepo) AreStatusChanged(ctx context.Context, productUrl string, status ProductStatus) (bool, error) {
	r := p.db.QueryRowContext(ctx, "SELECT COUNT(*) from products WHERE url=$1 AND status=$2", productUrl, status)
	var count int32
	err := r.Scan(&count)
	if err != nil {
		apm.CaptureError(ctx, err)
		return false, err
	}

	return count == 0, nil
}
func dirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
func NewProductsRepo(monitorSlug string, ctx context.Context) (ProductsRepo, error) {
	relativePathToDb := "data"
	if !dirExists(relativePathToDb) {
		err := os.Mkdir(relativePathToDb, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	database, err := apmsql.Open("sqlite3", relativePathToDb+"/"+monitorSlug+".sqlite")
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	createStmt, err := database.PrepareContext(ctx, "CREATE TABLE IF NOT EXISTS products (url varchar(1024) PRIMARY KEY, status int)")
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	_, err = createStmt.ExecContext(ctx)
	if err != nil {
		apm.CaptureError(ctx, err)
		return nil, err
	}

	repo := productsRepo{
		db: database,
	}

	fmt.Println("Established connection with local database")

	return &repo, nil
}
