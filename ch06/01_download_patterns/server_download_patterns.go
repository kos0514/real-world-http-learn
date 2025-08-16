// ch06/01_download_patterns/server_download_patterns.go
// 1つのサーバーで、ダウンロードに関する代表的なパターンをまとめて体験できるサンプルです。
// 目的:
// - Content-Type / Content-Disposition による「表示 or ダウンロード」挙動の違い
// - attachment / inline の使い分けと、国際化ファイル名（filename*）の扱い
// - <a download> / JavaScript によるダウンロード誘導
// - 「ありがとうページ」→ meta refresh → ダウンロードの流れ
// - 擬似「署名付き URL」（有効期限 + 署名）でのダウンロード
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	shared "real-world-http-learn/ch06/01_download_patterns/shared"
)

func main() {
	mux := http.NewServeMux()

	// トップページ: 各デモへのリンクを表示します。
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>ch06/01_download_patterns</title>
		<h1>ch06/01_download_patterns</h1>
		<ul>
		  <li><a href="/attachment">/attachment</a> … Content-Disposition: attachment（UTF-8日本語ファイル名）</li>
		  <li><a href="/inline">/inline</a> … Content-Disposition: inline（PNGをインライン表示）</li>
		  <li><a href="/public/a_download/">/public/a_download/</a> … &lt;a download&gt; の例</li>
		  <li><a href="/public/js_download/">/public/js_download/</a> … JS で動的にダウンロード</li>
		  <li><a href="/public/thanks/">/public/thanks/</a> … ありがとう → 自動ダウンロード（/download_file）</li>
		  <li><a href="/public/signed_url/">/public/signed_url/</a> … 署名URL → ダウンロードの例</li>
		</ul>
		<p>API: <code>/api/download</code>, <code>/api/sign</code> → <code>/api/file</code></p>
		`)
	})

	// Attachment の例: Content-Disposition: attachment を付与して、
	// ブラウザに「表示ではなく保存」を促します。name クエリでファイル名を指定可能。
	mux.HandleFunc("/attachment", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "レポート 2025-03.xlsx"
		}
		// MIME は実データに合わせる（ここでは Excel を想定）
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		// filename*（UTF-8）+ filename（ASCII フォールバック）の 2本立て
		w.Header().Set("Content-Disposition", shared.BuildAttachmentDisposition(name, "report.xlsx"))
		w.WriteHeader(http.StatusOK)
		// 実データの代わりにダミーのバイト列を返します
		_, _ = w.Write([]byte("DUMMY_EXCEL_BINARY"))
	})

	// Inline 画像の例: 1x1 PNG をメモリから返し、ブラウザ内でインライン表示させます。
	mux.HandleFunc("/inline", func(w http.ResponseWriter, r *http.Request) {
		filename := "サンプル画像.png"
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", shared.BuildInlineDisposition(filename, "sample.png"))
		_, _ = w.Write(oneByOnePNG())
	})

	// 静的配信: public/（HTML/JS）と assets/（テキスト等）
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("ch06/01_download_patterns/public"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("ch06/01_download_patterns/assets"))))

	// JS 用 API: ダウンロード対象を attachment で返す。
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "hello.html"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Disposition", shared.BuildAttachmentDisposition(name, "hello.html"))
		_, _ = fmt.Fprint(w, "<!doctype html><meta charset=utf-8><h1>Hello!</h1>")
	})

	// 署名 URL 発行 API（擬似）: name + 期限(exp) に HMAC を付けた URL を返します。
	mux.HandleFunc("/api/sign", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "secret.txt"
		}
		exp := time.Now().Add(30 * time.Second).Unix() // 30秒有効
		base := name + "|" + strconv.FormatInt(exp, 10)
		token := sign(base)

		q := url.Values{}
		q.Set("name", name)
		q.Set("exp", strconv.FormatInt(exp, 10))
		q.Set("token", token)
		u := "/api/file?" + q.Encode()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"url": u})
	})

	// 署名 URL の実体: token/exp/name を検証し、OK なら attachment でコンテンツを返します。
	mux.HandleFunc("/api/file", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		expStr := r.URL.Query().Get("exp")
		token := r.URL.Query().Get("token")
		exp, _ := strconv.ParseInt(expStr, 10, 64)
		if time.Now().Unix() > exp {
			http.Error(w, "link expired", http.StatusForbidden)
			return
		}
		base := name + "|" + expStr
		if sign(base) != token {
			http.Error(w, "invalid token", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", shared.BuildAttachmentDisposition(name, "secret.txt"))
		_, _ = w.Write([]byte("これは擬似署名URLで配られたコンテンツです。\n"))
	})

	// 「ありがとう」→ 自動ダウンロードの例: attachment を付けた PDF 風バイト列を返します。
	// 実 PDF ではありませんが、attachment の動作確認には十分です。
	mux.HandleFunc("/download_file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", shared.BuildAttachmentDisposition("入門ガイド.pdf", "guide.pdf"))
		// 最小限のバイト列（ダミー）。attachment により保存ダイアログ/自動保存が期待されます。
		_, _ = w.Write([]byte("%PDF-1.4\n% Dummy PDF content for download demo.\n"))
	})

	addr := ":18061"
	srv := &http.Server{Addr: addr, Handler: logging(mux), ReadHeaderTimeout: 5 * time.Second}

	// ブラウザで開くためのローカル URL を分かりやすく表示
	// 例: addr=":18061" → http://localhost:18061/
	host := "localhost"
	port := strings.TrimPrefix(addr, ":")
	if h, p, err := net.SplitHostPort(addr); err == nil {
		if h == "" || h == "0.0.0.0" || h == "::" {
			host = "localhost"
		} else if h == "127.0.0.1" {
			host = "127.0.0.1"
		} else {
			host = h
		}
		port = p
	}
	browseURL := fmt.Sprintf("http://%s:%s/", host, port)
	log.Printf("download-patterns server listening on %s", addr)
	log.Printf("ブラウザで開く: %s", browseURL)

	log.Fatal(srv.ListenAndServe())
}

// oneByOnePNG は、透明な 1x1 PNG を返します（最小のインライン表示デモ用）。
func oneByOnePNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
		0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c,
		0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

// logging は簡易的なアクセスログ用ミドルウェアです。
// 各リクエストのメソッド/パス/処理時間を出力します。
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// HMAC の秘密鍵（デモ用）
var secret = []byte("very-secret")

// sign は "name|exp" のような文字列に HMAC-SHA256 を計算し、hex 文字列で返します。
// 署名 URL の完全性検証に使用します。
func sign(base string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(base))
	return hex.EncodeToString(mac.Sum(nil))
}
