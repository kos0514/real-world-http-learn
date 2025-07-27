// head_request.go - HTTP HEADリクエストの例
// HEADリクエストはGETリクエストと似ていますが、レスポンスボディを返さずヘッダーのみを取得します
package main

import (
	"log"
	"net/http"
)

func main() {
	// ローカルサーバーにHTTP HEADリクエストを送信
	// HEADリクエストはリソースのメタデータ（ヘッダー）のみを取得する場合に有用です
	resp, err := http.Head("http://localhost:18888")
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}
	// 注意: 本来はrespがnilでないことを確認し、resp.Body.Close()を呼び出すべきです
	// HEADリクエストではボディが空でも、接続を適切に閉じるためにCloseは必要です

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("Status:", resp.Status)
	// レスポンスヘッダーのすべてのフィールドを出力
	log.Println("Headers:", resp.Header)
}
