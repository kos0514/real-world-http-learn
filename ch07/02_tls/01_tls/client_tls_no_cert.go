package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// OS の既定の信頼ストア（システム CA）だけを使って HTTPS にアクセスするサンプル。
// このリポジトリのサーバー証明書は「自前の CA」で署名されているため、
// OS にその CA を登録していない場合は TLS 検証エラーになります。
//
// 動かし方の選択肢:
//   - 推奨: ch07/02_tls/ca/certs/ca.crt を OS の信頼ストアにインポートしてから実行する。
//   - もう一つのサンプル（client_tls_with_cert.go）では、CA を手動で読み込んで検証できます。
//   - 学習用途のみ: Transport の TLS 設定で InsecureSkipVerify=true を使う方法もありますが、
//     実運用では非推奨です（検証を無効化するためMITM等に弱くなります）。
func main() {
	// 既定の http.Client を使用します（環境が対応していれば HTTP/2 が自動で利用されます）
	resp, err := http.Get("https://localhost:18443")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	log.Println(string(dump))
}
