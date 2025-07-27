// simple_get.go - シンプルなHTTP GETリクエストの例
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// ローカルサーバーにHTTP GETリクエストを送信
	resp, err := http.Get("http://localhost:18888")
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}
	// respがnilでないことを確認してからBodyをクローズ
	if resp != nil {
		// 関数終了時にレスポンスボディを必ず閉じる
		defer resp.Body.Close()
	} else {
		log.Fatalf("レスポンスがnilです")
		return
	}

	// レスポンスボディを読み込む
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// ボディの読み込み中にエラーが発生した場合はパニック
		panic(err)
	}

	// レスポンスボディを文字列として出力
	log.Println(string(body))

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("Status:", resp.Status)
	// レスポンスのステータスコード（例: 200）を出力
	log.Println("StatusCode:", resp.StatusCode)
	// レスポンスヘッダーのすべてのフィールドを出力
	log.Println("Fields:", resp.Header)
	// Content-Lengthヘッダーの値を出力
	log.Println("Content-Length:", resp.Header.Get("Content-Length"))
}
