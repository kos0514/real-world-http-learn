// post_any_content_text.go - テキストデータを送信するHTTP POSTリクエストの例
// http.Post関数を使用してテキストデータを送信する方法を示します
package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	// テキストの送信
	// strings.NewReaderを使用してテキストデータをPOSTリクエストのボディとして送信
	log.Println("--- テキスト送信の例 ---")
	reader := strings.NewReader("テキスト")
	resp, err := http.Post("http://localhost:18888", "text/plain", reader)
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}
	// respがnilでないことを確認してからBodyをクローズ
	if resp != nil {
		defer resp.Body.Close()
	} else {
		log.Fatalf("レスポンスがnilです")
		return
	}

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("テキスト送信のStatus:", resp.Status)
}
