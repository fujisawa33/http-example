package main

import (
	"net/http"
)

type MyTransport struct {
	wrapped http.RoundTripper

	maxRetryCounts int // 最大リトライ数
	retryCounts    int // リトライした数
}

func NewLimitedTransport(transport http.RoundTripper) *MyTransport {
	return &MyTransport{
		wrapped: transport,
	}
}

func (t *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.wrapped.RoundTrip(req)
}
