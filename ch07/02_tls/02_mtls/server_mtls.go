package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"real-world-http-learn/ch07/02_tls/tlsutil"
)

// mTLS（相互TLS）対応サーバー。
// - クライアント証明書の提示と検証を必須化（ClientAuth: RequireAndVerifyClientCert）
// - クライアント証明書の検証用に、自前 CA（ca/certs/ca.crt）を ClientCAs に設定
// - サーバーの証明書/秘密鍵は server/certs/server.crt と server/private/server.key
// 実行は ch07/02_tls をカレントディレクトリにして行う前提です。

// handlerMTLS は受け取った HTTP リクエストをダンプし、簡単な HTML を返します。
// 併せて r.TLS から TLS 交渉結果（バージョン/暗号/ALPN/SNI/再開/クライアント証明書）をログ出力します。
func handlerMTLS(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(dump))

	if r.TLS != nil {
		tlsutil.LogServerState(r.TLS)
	}

	fmt.Fprintf(w, "<html><body>hello (mTLS)</body></html>\n")
}

func main() {
	// クライアント証明書の検証に使用する CA（自前 CA）を読み込む
	caPEM, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	clientCAPool := x509.NewCertPool()
	if !clientCAPool.AppendCertsFromPEM(caPEM) {
		log.Fatal("failed to append client CA cert")
	}

	tlsConf := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert, // クライアント証明書を必須に
		ClientCAs:  clientCAPool,                   // 検証用 CA
		MinVersion: tls.VersionTLS12,               // 最低 TLS 1.2
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		NextProtos:       []string{"h2", "http/1.1"}, // ALPN で HTTP/2 と HTTP/1.1 を広告
	}

	server := &http.Server{
		Addr:      ":18443",
		TLSConfig: tlsConf,
		Handler:   http.DefaultServeMux,
	}

	http.HandleFunc("/", handlerMTLS)

	log.Println("start https listening :18443 (mTLS)")
	if err := server.ListenAndServeTLS("server/certs/server.crt", "server/private/server.key"); err != nil {
		log.Println(err)
	}
}
