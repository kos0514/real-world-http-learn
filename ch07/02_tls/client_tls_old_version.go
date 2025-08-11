package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
)

// 非準拠クライアント（旧バージョン）
// - サーバは MinVersion=TLS1.2
// - このクライアントは TLS1.1 以下に限定するため、バージョン不一致で拒否される想定
func main() {
	caPEM, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		log.Fatal("failed to append CA")
	}

	tlsConf := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS10,
		MaxVersion: tls.VersionTLS11,
	}

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}

	_, err = client.Get("https://localhost:18443")
	if err != nil {
		// 期待どおり: バージョン不一致による失敗（例: protocol version not supported）
		log.Printf("expected protocol version error: %v", err)
		return
	}
	log.Fatal("UNEXPECTED: request succeeded; server accepted TLS1.1 or lower")
}
