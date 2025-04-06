package main

import (
	"io"
	"log"
	"net/http"
)

// main はプログラムのエントリーポイントです。
// ローカルホストのポート18888にHTTP GETリクエストを送信し、
// レスポンスのボディを読み取ってログに出力します。
// また、HTTPステータスコードとレスポンスヘッダーもログに出力します。
func main() {
	// http.Get を使用して指定されたURLにGETリクエストを送信します。
	resp, err := http.Get("http://localhost:18888")
	if err != nil { // エラーが発生した場合、プログラムをパニック状態にします。
		panic(err)
	}
	defer resp.Body.Close() // 関数が終了する前にレスポンスボディを閉じます。

	// レスポンスボディを読み取ります。
	body, err := io.ReadAll(resp.Body)
	if err != nil { // エラーが発生した場合、プログラムをパニック状態にします。
		panic(err)
	}
	// レスポンスボディの内容をログに出力します。
	log.Println(string(body))
	// 文字列で "200 OK" をログに出力します。
	log.Println("Status:", resp.Status)
	// 数値で 200 をログに出力します。
	log.Println("StatusCode:", resp.StatusCode)
	// レスポンスヘッダーをログに出力します。
	log.Println("Fields:", resp.Header)
	// Content-Length ヘッダーの値をログに出力します。
	log.Println("Content-Length:", resp.Header.Get("Content-Length"))
}
