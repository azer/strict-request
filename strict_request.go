package strictrequest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Options struct {
	AllowRedirects bool
	Body           io.Reader
	BodyBytes      []byte
	Headers        map[string]string
	MaxSizeMb      float32
	TimeoutMs      int
}

func StrictRequest(method, url string, options Options) (*http.Response, error) {
	body := options.Body
	if options.BodyBytes != nil {
		body = bytes.NewReader(options.BodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if options.MaxSizeMb > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=0-%d", int(options.MaxSizeMb*1000000)))
	}

	if options.Headers != nil {
		for k, v := range options.Headers {
			req.Header.Add(k, v)
		}
	}

	client := http.Client{}

	if options.TimeoutMs > 0 {
		client.Timeout = time.Duration(time.Duration(options.TimeoutMs) * time.Millisecond)
	}

	if !options.AllowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client.Do(req)
}

func Get(url string, options Options) (*http.Response, error) {
	return StrictRequest("GET", url, options)
}

func Post(url string, options Options) (*http.Response, error) {
	return StrictRequest("POST", url, options)
}

func Put(url string, options Options) (*http.Response, error) {
	return StrictRequest("PUT", url, options)
}

func Delete(url string, options Options) (*http.Response, error) {
	return StrictRequest("DELETE", url, options)
}
