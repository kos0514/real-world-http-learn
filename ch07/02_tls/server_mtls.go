package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

// TLS バージョンを人が読める文字列に変換（サーバー側・mTLS用）
func tlsVersionNameServer(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS1.3"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS10:
		return "TLS1.0"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

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
		log.Printf("[TLS][server] version=%s cipher=%s alpn=%q sni=%q resumed=%v peerCerts=%d",
			tlsVersionNameServer(r.TLS.Version),
			tls.CipherSuiteName(r.TLS.CipherSuite),
			r.TLS.NegotiatedProtocol,
			r.TLS.ServerName,
			r.TLS.DidResume,
			len(r.TLS.PeerCertificates),
		)
		if len(r.TLS.PeerCertificates) > 0 {
			log.Printf("[TLS][server] client cert subject=%s", r.TLS.PeerCertificates[0].Subject.String())
		} else {
			log.Printf("[TLS][server] no client certificate presented")
		}
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
