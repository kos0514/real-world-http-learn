// cookie_client.go - Cookieを自動的に処理するHTTPクライアントの例
// HTTPリクエスト間でCookieを保持し、自動的に送信する方法を示します
// サーバーが最初のレスポンスでCookieを設定すると、次のリクエストで自動的に送信されます
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

func main() {
	// Cookieを保存するためのCookieJarを作成
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	// HTTPクライアントにCookieJarを設定
	client := http.Client{
		Jar: jar,
	}

	// 同じエンドポイントに対して2回GETリクエストを送信
	// 2回目のリクエストでは、1回目のレスポンスで設定されたCookieが自動的に送信されます
	for i := 0; i < 2; i++ {
		fmt.Printf("\n===== %d回目のGETリクエスト =====\n", i+1)

		// GETリクエストを作成
		req, err := http.NewRequest("GET", "http://localhost:18888/cookie", nil)
		if err != nil {
			panic(err)
		}

		// サーバーURLを解析（Cookieの表示に使用）
		serverURL, _ := url.Parse("http://localhost:18888")

		if len(client.Jar.Cookies(serverURL)) == 0 {
			fmt.Println("Cookieはまだ設定されていません")
		}

		// リクエストをダンプして表示（送信前）
		fmt.Printf("\n==========リクエスト情報（%d回目）==========\n", i+1)
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(reqDump))

		// リクエストを送信
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp != nil {
			defer resp.Body.Close()
		} else {
			log.Fatalf("レスポンスがnilです")
			return
		}

		// レスポンス全体をダンプ
		fmt.Printf("\n==========レスポンス情報（%d回目）==========\n", i+1)
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(respDump))
	}
}
