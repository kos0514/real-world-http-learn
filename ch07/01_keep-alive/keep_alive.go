// ch07/keep_alive.go - Learning sample for HTTP Keep-Alive (connection reuse)
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

// This sample is a minimal HTTP/1.1 client to observe Keep-Alive (connection reuse).
//
// Explanation (What is Keep-Alive):
// - When sending multiple requests to the same server, you can reuse an existing TCP connection to avoid handshake costs.
// - Go's http.Transport enables persistent connections by default; only connections whose response bodies are fully read and closed are returned to the pool for reuse.
// - This sample uses httptrace's GotConn to observe fields like Reused, WasIdle, and IdleTime.
//
// Preconditions for reuse:
// - Always read resp.Body to EOF and then Close it (this code uses io.ReadAll and defer Close).
// - Do not explicitly send "Connection: close".
// - Ensure the server/client idle timeouts are aligned.
//
// How to run:
//  1. Start the simple server in this repo (e.g., `go run server.go`)
//  2. In another terminal, run this program (`go run ./ch07`)
//  3. Observe the logs to see whether the same connection is reused across requests.
//
// Environment variables:
//
//	KEEP_ALIVE_URL: Target URL (defaults to "http://localhost:18888")
func main() {
	url := os.Getenv("KEEP_ALIVE_URL")
	if url == "" {
		url = "http://localhost:18888"
	}

	// Transport with Keep-Alive enabled (enabled by default)
	transport := &http.Transport{
		// Set generous limits for concurrent and idle connections
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		// DisableKeepAlives: false, // default is false (Keep-Alive enabled)
	}

	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	// Send multiple GETs to observe connection reuse
	for i := 1; i <= 5; i++ {
		func(i int) {
			// Use httptrace to capture connection acquisition info
			var gotConnInfo httptrace.GotConnInfo
			trace := &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) {
					gotConnInfo = info
				},
			}

			// Create a context with the trace for this request
			ctx := httptrace.WithClientTrace(context.Background(), trace)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				log.Fatalf("failed to create request: %v", err)
			}

			// Perform the request
			resp, err := client.Do(req)
			if err != nil {
				log.Fatalf("http request failed: %v", err)
			}
			// Always close the body when leaving this scope
			defer resp.Body.Close()

			// Read the entire response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("failed to read response body: %v", err)
			}

			// Print only the first 80 bytes as a snippet
			snippet := string(body)
			if len(snippet) > 80 {
				snippet = snippet[:80] + "..."
			}

			fmt.Printf("[%d] status=%s reused=%t was_idle=%t idle_time=%s\n", i, resp.Status, gotConnInfo.Reused, gotConnInfo.WasIdle, gotConnInfo.IdleTime)
			fmt.Printf("[%d] body: %s\n", i, snippet)
		}(i)

		// Wait a moment before the next request to create an idle period
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
