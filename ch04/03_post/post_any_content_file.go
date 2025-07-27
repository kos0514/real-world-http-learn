// post_any_content_file.go - ファイルを送信するHTTP POSTリクエストの例
// http.Post関数を使用してファイルデータを送信する方法を示します
package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// ファイルの送信
	// ファイルをオープンしてそのコンテンツをPOSTリクエストのボディとして送信
	log.Println("--- ファイル送信の例 ---")
	file, err := os.Open("main.go")
	if err != nil {
		// ファイルオープンに失敗した場合はパニック
		panic(err)
	}
	// 関数終了時にファイルを必ず閉じる
	defer file.Close()

	// ファイルの内容をPOSTリクエストのボディとして送信
	// Content-Type: text/plainを指定
	resp, err := http.Post("http://localhost:18888", "text/plain", file)
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
	log.Println("ファイル送信のStatus:", resp.Status)
}
