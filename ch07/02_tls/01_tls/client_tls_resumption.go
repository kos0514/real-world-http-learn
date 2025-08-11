package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"real-world-http-learn/ch07/02_tls/tlsutil"
)

// セッション再開（1-RTT）を明示的に観測するクライアント。
// - tls.Config.ClientSessionCache を有効化
// - DisableKeepAlives=true で毎回新規 TCP/TLS ハンドシェイクを実行
// - 同一サーバーに複数回接続し、2 回目以降で resp.TLS.DidResume==true を観測
func main() {
	// 自前 CA を読み込む（実行時の作業ディレクトリは ch07/02_tls を想定）
	caPEM, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		log.Fatal("failed to append CA")
	}

	// TLS 設定: MinVersion を TLS1.2 に、セッションキャッシュを有効化
	tlsConf := &tls.Config{
		RootCAs:            pool,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
		ClientSessionCache: tls.NewLRUClientSessionCache(128),
	}

	// 毎回新規接続を張る（各リクエストで新規ハンドシェイクを発生させる）
	tr := &http.Transport{
		TLSClientConfig:   tlsConf,
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}

	for i := 1; i <= 3; i++ {
		resp, err := client.Get("https://localhost:18443")
		if err != nil {
			log.Fatalf("request #%d failed: %v", i, err)
		}
		if resp.TLS != nil {
			log.Printf("[#%d] DidResume=%v", i, resp.TLS.DidResume)
			tlsutil.LogClientState(resp.TLS)
		}
		dump, _ := httputil.DumpResponse(resp, true)
		log.Printf("[#%d] %s", i, string(dump))
		resp.Body.Close()

		// TLS 1.3 の NewSessionTicket が非同期で届く場合に備えて、短い待機を入れる
		time.Sleep(200 * time.Millisecond)
	}
}
