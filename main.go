package main

import (
	"net/http"
)

func main() {
	client := http.Client{
		Transport: NewMyTransport(
			http.DefaultTransport,
			5,    // 最大リトライ数
			3,    // 単位時間あたりのリクエスト数上限
			4500, // 単位時間(ms)
		),
	}

	{
		url := "https://ozuma.sakura.ne.jp/httpstatus/500"
		resp, _ := client.Get(url)
		defer resp.Body.Close()
	}

	for i := 0; i < 20; i++ {
		url := "https://ozuma.sakura.ne.jp/httpstatus/200"
		resp, _ := client.Get(url)
		defer resp.Body.Close()
	}

}
