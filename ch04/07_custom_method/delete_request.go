// delete_request.go - カスタムHTTPメソッド（DELETE）を使用したリクエストの例
// http.NewRequest関数を使用して、GETやPOST以外のHTTPメソッドでリクエストを送信する方法を示します
// DELETEメソッドは、指定されたリソースを削除するために使用されます
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	// 標準のHTTPクライアントを作成
	client := &http.Client{}

	// DELETEリクエストを作成
	// http.NewRequest関数を使用すると、任意のHTTPメソッドを指定できます
	// 第3引数のnilは、リクエストボディがないことを示します
	request, err := http.NewRequest("DELETE", "http://localhost:18888", nil)
	if err != nil {
		// リクエスト作成に失敗した場合はパニック
		panic(err)
	}

	// リクエストを送信
	resp, err := client.Do(request)
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}

	// 注意: 本来はrespがnilでないことを確認し、defer resp.Body.Close()を呼び出すべき
	// 接続リソースをリークさせないために、レスポンスボディは必ず閉じる必要があります

	// レスポンス全体（ヘッダーとボディ）をダンプ
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// レスポンスのダンプに失敗した場合はパニック
		panic(err)
	}

	// ダンプしたレスポンスを文字列として出力
	log.Println(string(dump))
}
