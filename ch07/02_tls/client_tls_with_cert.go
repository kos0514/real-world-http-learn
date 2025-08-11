package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

// TLS バージョンを人が読める文字列に変換（クライアント側）
func tlsVersionNameClient(v uint16) string {
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
		return "unknown"
	}
}

// 自前の CA 証明書を信頼して HTTPS に接続する最小クライアント。
//   - ch07/02_tls/ca/certs/ca.crt を RootCAs に読み込み、サーバー証明書の検証を有効にします。
//   - SNI/ホスト名検証については、tls.Config.ServerName に "localhost" を明示します
//     （http.Client は通常 URL のホスト名を自動設定しますが、学習用に明示）。
func main() {
	// CA 証明書を読み込む（実行ディレクトリは ch07/02_tls を想定）
	cert, err := os.ReadFile("ca/certs/ca.crt")
	if err != nil {
		panic(err)
	}

	// 信頼ストアを作成して、CA 証明書を追加
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	// TLS 設定を構築
	tlsConfig := &tls.Config{
		RootCAs:    certPool,    // この CA を信頼してサーバー証明書を検証する
		ServerName: "localhost", // SNI/ホスト名検証で使う名前（今回は localhost）
	}
	// 注意: tls.Config.BuildNameToCertificate は非推奨です（サーバー用の古い API）。
	// クライアントでは使用しません。

	// クライアントを作成（Transport に TLS 設定を適用）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// 通信を行う
	resp, err := client.Get("https://localhost:18443")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 可視化: 交渉結果の TLS 情報をログ出力
	if resp.TLS != nil {
		log.Printf("[TLS][client] version=%s cipher=%s alpn=%q sni=%q resumed=%v verifiedChains=%d",
			tlsVersionNameClient(resp.TLS.Version),
			tls.CipherSuiteName(resp.TLS.CipherSuite),
			resp.TLS.NegotiatedProtocol,
			resp.TLS.ServerName,
			resp.TLS.DidResume,
			len(resp.TLS.VerifiedChains),
		)
		if len(resp.TLS.PeerCertificates) > 0 {
			log.Printf("[TLS][client] server cert subject=%s", resp.TLS.PeerCertificates[0].Subject.String())
		}
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	log.Println(string(dump))
}
