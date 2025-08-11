package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"real-world-http-learn/ch07/02_tls/tlsutil"
)

// mTLS（相互TLS）クライアントの最小実装。
// - 自身のクライアント証明書/秘密鍵を提示（client/certs/client.crt と client/private/client.key）
// - サーバー検証のため、CA 証明書（ca/certs/ca.crt）を RootCAs に追加
// - ホスト名検証のため ServerName に "localhost" を設定
// 実行は ch07/02_tls をカレントディレクトリにして行う前提です。
func main() {
	// 1) クライアント証明書/秘密鍵を読み込む（無暗号鍵を想定）
	cert, err := tls.LoadX509KeyPair("client/certs/client.crt", "client/private/client.key")
	if err != nil {
		panic(err)
	}

	// 2) サーバー証明書の検証に使う CA を用意（自前 CA）
	caPEM, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		panic(err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caPEM) {
		panic("failed to append CA cert")
	}

	// 3) TLS 設定を構築
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert}, // クライアント証明書を提示
		RootCAs:      rootCAs,                 // サーバー証明書の検証用 CA
		ServerName:   "localhost",             // SNI/ホスト名検証
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConf},
	}

	// 4) 通信を行う
	resp, err := client.Get("https://localhost:18443")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 可視化: 交渉結果の TLS 情報をログ出力
	if resp.TLS != nil {
		tlsutil.LogClientState(resp.TLS)
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	log.Println(string(dump))
}
