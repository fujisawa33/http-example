package main

import (
	"math"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
)

type MyTransport struct {
	wrapped http.RoundTripper

	maxRetryCounts int // 最大リトライ数
	retryCounts    int // リトライした数

	maxRequestCounts int    // 単位時間あたりのリクエスト数上限
	perMilliSecond   int64  // 単位時間(ms)
	window           Window // 現在のwindow
}

func NewMyTransport(transport http.RoundTripper, maxRetryCounts int, maxRequestCounts int, perMilliSecond int64) *MyTransport {
	return &MyTransport{
		wrapped:          transport,
		maxRetryCounts:   maxRetryCounts,
		maxRequestCounts: maxRequestCounts,
		perMilliSecond:   perMilliSecond,
		retryCounts:      0,
		window: Window{
			key:           int64(0),
			requestCounts: 0,
		},
	}
}

// 固定期間(window)の構造体
type Window struct {
	key           int64 // windowのキー
	requestCounts int   // window内のリクエスト数
}

func (t *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Fixed Window Counter
	for {
		now := time.Now().UnixMilli()
		cKey := now / t.perMilliSecond

		if t.window.key != cKey {
			t.window = Window{
				key:           cKey,
				requestCounts: 0,
			}
			break
		}

		if t.window.requestCounts < t.maxRequestCounts {
			break
		}

		wait := t.perMilliSecond - now%t.perMilliSecond
		time.Sleep(time.Millisecond * time.Duration(wait))
	}
	t.window.requestCounts++

	var res *http.Response
	var err error
	for {
		res, err = t.wrapped.RoundTrip(req)
		log.Info(res.StatusCode)
		if res != nil && res.StatusCode < http.StatusInternalServerError {
			break
		}

		t.retryCounts++
		if t.retryCounts > t.maxRetryCounts {
			break
		}

		// Exponential BackOff
		time.Sleep(time.Second * time.Duration(math.Pow(2, float64(t.retryCounts))))
	}
	t.retryCounts = 0

	return res, err
}
