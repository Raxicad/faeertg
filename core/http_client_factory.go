package core

import (
	"encoding/base64"
	"fmt"
	"go.elastic.co/apm/module/apmhttp"
	"net/http"
	"net/url"
	"sync"
)

//import "github.com/bandar-monitors/monitors/core/http"

type HttpClientFactory interface {
	CreateClient() *http.Client
}

type httpClientFactory struct {
	proxiesPool          []*url.URL
	userAgentsRotatePool []string
	lastProxyUrlIdx      int
	mu                   *sync.Mutex
}

func (s *httpClientFactory) CreateClient() *http.Client {
	h := http.Client{}
	proxyURL := s.getNextProxy()
	if proxyURL != nil {
		hdr := http.Header{}
		if proxyURL.User != nil {
			pwd, pwdSet := proxyURL.User.Password()
			if pwdSet {
				auth := fmt.Sprintf("%s:%s", proxyURL.User.Username(), pwd)
				basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
				hdr.Add("Proxy-Authorization", basicAuth)
			}
		}
		transport := &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			ProxyConnectHeader: hdr,
		}

		h.Transport = transport
	}

	return apmhttp.WrapClient(&h)
}

func CreateHttpClientFactory(proxiesPool []string, userAgentsRotatePool []string) HttpClientFactory {
	var proxyUrlsPool []*url.URL
	for ix := 0; ix < len(proxiesPool); ix++ {
		rawUrl := proxiesPool[ix]
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			ix--
			proxiesPool = append(proxiesPool[:ix], proxiesPool[ix+1:]...)
		} else {
			proxyUrlsPool = append(proxyUrlsPool, parsedUrl)
		}
	}
	factory := httpClientFactory{
		proxiesPool:          proxyUrlsPool,
		userAgentsRotatePool: userAgentsRotatePool,
		mu:                   &sync.Mutex{},
		lastProxyUrlIdx:      -1,
	}

	return &factory
}

func (s *httpClientFactory) getNextProxy() *url.URL {
	if len(s.proxiesPool) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	nextIx := s.lastProxyUrlIdx + 1
	if nextIx >= len(s.proxiesPool) {
		nextIx = 0
	}

	nextProxyUrl := s.proxiesPool[nextIx]

	s.lastProxyUrlIdx = nextIx
	return nextProxyUrl
}
