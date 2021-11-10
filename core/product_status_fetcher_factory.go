package core

type ProductStatusFetcherFactory interface {
	CreateFetcher(productUrl string) ProductStatusFetcher
	SetHttpClientFactory(factory HttpClientFactory)
}

type ProductTargetsFactory interface {
	GetTargetsFromUrl(productUrl string) []string
}
