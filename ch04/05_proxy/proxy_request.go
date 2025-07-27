// proxy_request.go - プロキシを使用したHTTP GETリクエストの例
// プロキシサーバーを経由してHTTPリクエストを送信する方法を示します
// プロキシは、クライアントとサーバーの間に位置し、リクエストを中継する役割を果たします
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// プロキシサーバーのURLを解析
	// この例では、ローカルの18888ポートで動作するサーバーをプロキシとして使用
	proxyUrl, err := url.Parse("http://localhost:18888")
	if err != nil {
		// URL解析に失敗した場合はパニック
		panic(err)
	}

	// プロキシを使用するHTTPクライアントを作成
	// Transport.ProxyフィールドにプロキシURLを設定することで、
	// すべてのリクエストがこのプロキシを経由して送信されます
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	// HTTP GETリクエストを送信（宛先URLはhttp://github.comを指定）
	// 注意: 実際にはこのリクエストはGitHubに直接送信されず、
	// 上記で設定したプロキシサーバー（localhost:18888）に送信され、
	// プロキシサーバーがリクエストを処理します
	resp, err := client.Get("http://github.com")
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}

	// 注意: 本来はrespがnilでないことを確認し、defer resp.Body.Close()を呼び出すべき
	// 接続リソースをリークさせないために、レスポンスボディは必ず閉じる必要があります

	// レスポンス全体（ヘッダーとボディ）をダンプ
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// レスポンスのダンプに失敗した場合はパニック
		panic(err)
	}

	// ダンプしたレスポンスを文字列として出力
	log.Println(string(dump))
}
