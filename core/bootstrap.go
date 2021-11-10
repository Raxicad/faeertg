package core

import (
	"context"
	"fmt"
	"github.com/bandar-monitors/monitors/core/util"
	"github.com/joho/godotenv"
	"go.elastic.co/apm"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

type ProductUrlSpec struct {
	Url           string
	WatchersCount int
}

type GenericBootstrapArguments struct {
	ProductUrls    []*ProductUrlSpec
	ProxiesPool    []string
	UserAgentsPool []string
	AmqpConn       string
	MonitorSlug    string
	FetcherFactory ProductStatusFetcherFactory
}

const ProductUrlSpecDelim = '|'

func ParseProductUrlSpec(value string) (*ProductUrlSpec, error) {
	tokens := strings.FieldsFunc(value, func(r rune) bool {
		return r == ProductUrlSpecDelim
	})

	if len(tokens) == 1 {
		return &ProductUrlSpec{
			Url:           strings.TrimSpace(tokens[0]),
			WatchersCount: 1,
		}, nil
	}

	c, err := strconv.Atoi(tokens[0])
	if err != nil {
		return nil, err
	}

	return &ProductUrlSpec{
		Url:           strings.TrimSpace(tokens[1]),
		WatchersCount: c,
	}, nil
}

func EnvAmqpConn() string {
	amqpConn := os.Getenv("CONNECTION_STRINGS_RABBITMQ")
	if len(amqpConn) == 0 {
		panic("no amqp conn provided")
	}

	return amqpConn
}
func EnvProductUrls(paramName string) []*ProductUrlSpec {
	return EnvProductUrlsMapped(paramName, nil)
}

var EnvProductsDelim = '\n'

func EnvProductUrlsMapped(paramName string, mapValue func(v string) string) []*ProductUrlSpec {
	productUrls := os.Getenv(paramName)
	if len(productUrls) == 0 {
		panic(paramName + "is missing")
	}

	var productsToTrack []*ProductUrlSpec
	skus := util.SplitUniqueNonEmptyEntries(productUrls, EnvProductsDelim)
	for _, sku := range skus {
		spec, err := ParseProductUrlSpec(sku)
		if err != nil {
			panic(err)
		}

		if mapValue != nil {
			spec.Url = mapValue(spec.Url)
		}

		productsToTrack = append(productsToTrack, spec)
	}

	return productsToTrack
}

func EnvProxiesPool() []string {
	proxiesPoolStr := os.Getenv("PROXIES_POOL")
	return util.SplitUniqueNonEmptyEntries(proxiesPoolStr, '\n')
}
func EnvUserAgentsPool() []string {
	// stub for now
	return []string{}
}

func InitEnv(monitorSlug string) {
	_ = godotenv.Load(path.Join("sites", monitorSlug, ".env"), ".env")
}
func GenericBootstrapMonitorWithDefaults(monitorSlug string, fetcherFactory ProductStatusFetcherFactory,
	productsToTrackProvider func() []*ProductUrlSpec) {
	InitEnv(monitorSlug)
	args := &GenericBootstrapArguments{
		ProductUrls:    productsToTrackProvider(),
		ProxiesPool:    EnvProxiesPool(),
		UserAgentsPool: EnvUserAgentsPool(),
		AmqpConn:       EnvAmqpConn(),
		MonitorSlug:    monitorSlug,
		FetcherFactory: fetcherFactory,
	}

	GenericBootstrapMonitor(args)
}
func GenericBootstrapMonitor(args *GenericBootstrapArguments) {
	runtimeCtx := context.Background()
	tx := apm.DefaultTracer.StartTransaction("startup", args.MonitorSlug)
	tx.Context.SetLabel("products_count", len(args.ProductUrls))
	tracingCtx := apm.ContextWithTransaction(runtimeCtx, tx)

	factory := CreateHttpClientFactory(args.ProxiesPool, args.UserAgentsPool)
	args.FetcherFactory.SetHttpClientFactory(factory)
	publisher, err := CreatePublisher(args.MonitorSlug, args.AmqpConn, tracingCtx)
	if err != nil {
		util.FailOnError(err, "can't create publisher")
	}

	collector := StartCollector(args.MonitorSlug, publisher.PublishStats)
	repo, err := NewProductsRepo(args.MonitorSlug, tracingCtx)
	if err != nil {
		util.FailOnError(err, "can't create products repo")
	}

	watchersFactory := CreateGenericWatchersFactory(factory, args.MonitorSlug, args.FetcherFactory)
	manager := CreateWatchersManager(args.MonitorSlug, repo, publisher /* new(fakePublisher)*/, watchersFactory, collector)
	defer manager.Dispose()

	for _, product := range args.ProductUrls {
		manager.SpawnWatcher(tracingCtx, int(math.Max(1, float64(product.WatchersCount))), product.Url, Available)
	}

	fmt.Println("Spawned watchers for all products")
	tx.End()
	util.WaitForShutdown()
}
