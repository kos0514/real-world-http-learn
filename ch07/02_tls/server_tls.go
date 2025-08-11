// TLS 対応の最小 HTTP サーバーのサンプル
// ポート 18443 で HTTPS を待ち受け、受け取ったリクエストをダンプしてから簡単な HTML を返します。
// 証明書は server/certs/server.crt、秘密鍵は server/private/server.key を使用します（自己署名 CA で署名済みを想定）。
// 動作確認例:
//
//	curl -k https://localhost:18443/                                 // 検証をスキップ（CA 未登録時）
//	curl --cacert ca/certs/ca.crt https://localhost:18443/            // 付属の CA を使って検証
//	ブラウザで警告が出る場合は CA を OS に信頼登録するか、curl -k を使用してください。
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"real-world-http-learn/ch07/02_tls/tlsutil"
)

// handler は受け取った HTTP リクエストを標準出力にダンプし、固定の HTML を返します。
// DumpRequest の第 2 引数 true は「ボディも含めてダンプする」指定です。
func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		// ダンプに失敗した場合は 500 を返す
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	// 受信したリクエスト全体をコンソールに表示
	fmt.Println(string(dump))

	// 可視化: この接続で交渉された TLS 情報をログ出力
	if r.TLS != nil {
		tlsutil.LogServerState(r.TLS)
	}

	// クライアントへ簡単な HTML を返す
	fmt.Fprintf(w, "<html><body>hello</body></html>\n")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	// TLS をハードニングしつつ、ALPN では HTTP/2 と HTTP/1.1 の両方を広告
	tlsConf := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		NextProtos:       []string{"h2", "http/1.1"},
	}
	server := &http.Server{
		Addr:      ":18443",
		Handler:   mux,
		TLSConfig: tlsConf,
	}

	log.Println("start https listening :18443 (ALPN: h2/http1.1)")
	if err := server.ListenAndServeTLS("./server/certs/server.crt", "./server/private/server.key"); err != nil {
		// サーバーが終了した（または起動失敗した）場合のエラーログ
		log.Println(err)
	}
}
