// ch07/keep_alive.go - HTTP Keep-Alive（接続再利用）学習用サンプル
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

// このサンプルは、Keep-Alive（接続再利用）を観察するための最小限の HTTP/1.1 クライアントです。
//
// 解説（Keep-Alive とは）:
// - 同じサーバーに複数のリクエストを送る際、既存の TCP 接続を再利用してハンドシェイクのコストを省けます。
// - Go の http.Transport は既定で永続接続を有効にします。レスポンスボディを最後まで読み取り Close された接続のみが、再利用のためにプールへ戻されます。
// - このサンプルは httptrace の GotConn を用いて、Reused / WasIdle / IdleTime などのフィールドを観察します。
//
// 再利用の前提条件:
// - resp.Body を EOF まで読み、その後に Close すること（このコードでは io.ReadAll と defer Close を使用）。
// - 明示的に "Connection: close" を送らないこと。
// - サーバー／クライアントのアイドルタイムアウト設定を適切に揃えること。
//
// 実行方法:
//  1. このリポジトリの簡易サーバーを起動（例: `go run server.go`）
//  2. 別のターミナルで本プログラムを実行（`go run ./ch07`）
//  3. ログを確認し、同じ接続がリクエスト間で再利用されているか観察する。
//
// 環境変数:
//
//	KEEP_ALIVE_URL: 対象 URL（既定は "http://localhost:18888"）
func main() {
	url := os.Getenv("KEEP_ALIVE_URL")
	if url == "" {
		url = "http://localhost:18888"
	}

	// Keep-Alive を有効にした Transport（デフォルトで有効）
	transport := &http.Transport{
		// 同時接続およびアイドル接続の上限を広めに設定
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		// DisableKeepAlives: false, // 既定は false（Keep-Alive 有効）
	}

	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	// 接続再利用を観察するために複数回の GET を送信
	for i := 1; i <= 5; i++ {
		func(i int) {
			// httptrace を使って接続取得時の情報を取得する
			var gotConnInfo httptrace.GotConnInfo
			trace := &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) {
					gotConnInfo = info
				},
			}

			// このリクエスト用のトレースを含む context を作成
			ctx := httptrace.WithClientTrace(context.Background(), trace)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				log.Fatalf("failed to create request: %v", err)
			}

			// リクエストを実行
			resp, err := client.Do(req)
			if err != nil {
				log.Fatalf("http request failed: %v", err)
			}
			// このスコープを抜ける際に必ず Body を Close する
			defer resp.Body.Close()

			// レスポンスボディを最後まで読み込む
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("failed to read response body: %v", err)
			}

			// 先頭 80 バイトのみをスニペットとして表示
			snippet := string(body)
			if len(snippet) > 80 {
				snippet = snippet[:80] + "..."
			}

			fmt.Printf("[%d] status=%s reused=%t was_idle=%t idle_time=%s\n", i, resp.Status, gotConnInfo.Reused, gotConnInfo.WasIdle, gotConnInfo.IdleTime)
			fmt.Printf("[%d] body: %s\n", i, snippet)
		}(i)

		// アイドル期間を作るため、次のリクエスト前に少し待機
		time.Sleep(300 * time.Millisecond)
	}

	log.Println("【まとめ】Keep-Alive の挙動")
	log.Println("- 各リクエストでボディを最後まで読み Close すると、接続はアイドルプールに戻る")
	log.Println("- 後続の同一ホストへのリクエストは接続を再利用し、ハンドシェイクを省いて遅延を削減")
	log.Println("- ループ中は Transport を生かしておき、最後に必要ならアイドル接続を閉じる")
	log.Println("- ループ内で CloseIdleConnections すると次の再利用が起きない")

	// 最後に、必要に応じてアイドル接続を明示的に閉じます。
	transport.CloseIdleConnections()
}
