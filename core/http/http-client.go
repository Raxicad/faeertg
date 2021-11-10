package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
)

type HttpClient struct {
	parsedUrl  *url.URL
	request    *fasthttp.Request
	response   *fasthttp.Response
	sharedBody []byte
	client     *fasthttp.Client
}

func NewNotInitialized() *HttpClient {
	req := fasthttp.AcquireRequest()
	// fasthttp does not automatically request a gzipped response.
	// We must explicitly ask for it.
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:79.0) Gecko/20100101 Firefox/79.0")

	resp := fasthttp.AcquireResponse()

	client := HttpClient{
		request:  req,
		response: resp,
		client:   &fasthttp.Client{},
	}

	return &client
}

func (client *HttpClient) GetUrl() string {
	return client.parsedUrl.String()
}

func (client *HttpClient) SetUrl(rawUrl string) {
	client.parsedUrl, _ = url.Parse(rawUrl)
	client.request.SetRequestURI(rawUrl)
}

func (client *HttpClient) SetHost(host string) {
	client.request.Header.Set("Host", host)
}

func (client *HttpClient) SetReferer(url string) {
	client.request.Header.Set("Referer", url)
}

func New(url string, referer string) *HttpClient {
	r := NewNotInitialized()
	r.SetUrl(url)
	r.SetHost(r.parsedUrl.Host)
	if len(referer) > 0 {
		r.SetReferer(referer)
	}

	return r
}

func (client *HttpClient) SetMethod(method string) {
	client.request.Header.SetMethod(method)
}

func (client *HttpClient) SetHeader(key string, value string) {
	client.request.Header.Set(key, value)
}

func (client *HttpClient) SetBody(reader *bytes.Reader) {
	client.request.SetBodyStream(reader, reader.Len())
}

func (client *HttpClient) Release() {
	fasthttp.ReleaseRequest(client.request)
	fasthttp.ReleaseResponse(client.response)
}

func (client *HttpClient) Execute() ([]byte, error) {
	err := client.client.Do(client.request, client.response)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return nil, err
	}

	// Do we need to decompress the response?
	contentEncoding := client.response.Header.Peek("Content-Encoding")
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		//fmt.Println("Unzipping...")
		client.sharedBody, _ = client.response.BodyGunzip()
	} else {
		client.sharedBody = client.response.Body()
	}

	if client.response.StatusCode() >= 400 {
		//fmt.Printf("Non success status code %d\n", client.response.StatusCode())
		bodyStr := ""
		if client.response != nil {
			bodyStr = string(client.response.Body())
		}

		return client.sharedBody, errors.New(fmt.Sprintf("Non success status code %d\n%s", client.response.StatusCode(), bodyStr))
	}

	return client.sharedBody, nil
}
