// http1.0/server.go は単純なHTTPサーバーを実装します。
// このサーバーは受信したHTTPリクエストの内容をコンソールに表示し、
// シンプルなHTMLレスポンスを返します。
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

// handler はHTTPリクエストを処理する関数です。
// 受信したリクエストの内容をダンプして表示し、
// 単純なHTML応答を返します。
//
// パラメータ:
//   - w: HTTPレスポンスを書き込むためのResponseWriter
//   - r: 処理すべきHTTPリクエスト
func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println("==========リクエスト情報==========")
	fmt.Println(string(dump))
	fmt.Fprintf(w, "<html><body>hello</body></html>\n")
}

// cookieHandler は/cookieパスへのリクエストを処理する関数です。
// POSTとGETの両方のリクエストでCookieを設定・更新します。
// Cookieの有無に基づいて異なるコンテンツを返します。
//
// パラメータ:
//   - w: HTTPレスポンスを書き込むためのResponseWriter
//   - r: 処理すべきHTTPリクエスト
func cookieHandler(w http.ResponseWriter, r *http.Request) {
	// リクエスト情報をダンプして表示
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println("==========リクエスト情報==========")
	fmt.Println(string(dump))

	// POSTとGETの両方のリクエストを処理
	if r.Method == "POST" || r.Method == "GET" {
		// Cookieを確認
		cookie, err := r.Cookie("VISIT")

		if err == http.ErrNoCookie {
			// Cookieがない場合（初回訪問）- 新しいCookieを設定
			expiration := time.Now().Add(24 * time.Hour)
			newCookie := http.Cookie{
				Name:     "VISIT",
				Value:    "1",
				Expires:  expiration,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, &newCookie)

			// JSONレスポンスを返す
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "初めての訪問です - Cookieを設定しました",
				"visits":  "1",
			})
		} else if err != nil {
			// その他のエラー
			http.Error(w, "Cookieの読み込みに失敗しました", http.StatusInternalServerError)
		} else {
			// Cookieがある場合（訪問回数を増やす）
			visitCount, err := strconv.Atoi(cookie.Value)
			if err != nil {
				// 数値変換エラー
				http.Error(w, "Cookie値の変換に失敗しました", http.StatusInternalServerError)
				return
			}
			visitCount++

			// 新しいCookieを設定
			expiration := time.Now().Add(24 * time.Hour)
			newCookie := http.Cookie{
				Name:     "VISIT",
				Value:    strconv.Itoa(visitCount),
				Expires:  expiration,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, &newCookie)

			// JSONレスポンスを返す
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "訪問回数を更新しました",
				"visits":  strconv.Itoa(visitCount),
			})
		}
	} else {
		// その他のHTTPメソッド
		http.Error(w, "サポートされていないHTTPメソッドです", http.StatusMethodNotAllowed)
		return
	}
}

// main はプログラムのエントリーポイントです。
// ポート18888でHTTPサーバーを起動し、
// 各パスに対応するハンドラ関数を登録します。
func main() {
	var httpServer http.Server
	http.HandleFunc("/", handler)
	http.HandleFunc("/cookie", cookieHandler)
	log.Println("start http listening :18888")
	httpServer.Addr = ":18888"
	log.Println(httpServer.ListenAndServe())
}
