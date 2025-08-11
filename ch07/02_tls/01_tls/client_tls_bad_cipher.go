package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"

	"real-world-http-learn/ch07/02_tls/tlsutil"
)

// 非準拠クライアント（例示用）
// - サーバーは TLS1.2 以上かつ AEAD (GCM/ChaCha20) のみ許可
// - このクライアントは TLS1.2 に固定し、CBC 系のみを提示するため、ハンドシェイクで拒否される想定
func main() {
	// 自前 CA を読み込み（証明書検証は通す）。失敗要因を暗号スイート不一致に限定するため
	caPEM, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		log.Fatal("failed to append CA")
	}

	// TLS1.2 に固定し、非AEAD（CBC）スイートのみ提示
	tlsConf := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			// CBC 系（サーバ側の許可リスト外）
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		},
	}

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}

	resp, err := client.Get("https://localhost:18443")
	if err != nil {
		// 期待どおり: 共通の暗号スイートが見つからずハンドシェイク失敗
		log.Printf("expected handshake failure due to disallowed cipher suites: %v", err)
		return
	}
	defer resp.Body.Close()

	// 成功してしまった場合（想定外）のデバッグ出力
	if resp.TLS != nil {
		tlsutil.LogClientState(resp.TLS)
	}
	log.Println("UNEXPECTED: request succeeded; server accepted one of the proposed CBC suites")
}
