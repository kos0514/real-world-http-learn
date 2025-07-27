// header_send.go - HTTPヘッダーフィールドの送信例
// リクエストヘッダーを設定する様々な方法を示します
// ヘッダーフィールドはHTTPリクエストのメタデータを提供します
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	// 標準のHTTPクライアントを作成
	client := &http.Client{}

	// POSTリクエストを作成（ヘッダーを設定するため）
	// 実際のリクエストボディはここでは省略（nilを使用）
	request, err := http.NewRequest("POST", "http://localhost:18888", nil)
	if err != nil {
		// リクエスト作成に失敗した場合はパニック
		panic(err)
	}

	// 方法1: Header.Add()メソッドを使用してヘッダーフィールドを追加
	// Content-Typeヘッダーを設定（curlの -H "Content-Type=image/jpeg" に相当）
	request.Header.Add("Content-Type", "image/jpeg")

	// 方法2: SetBasicAuth()メソッドを使用して認証ヘッダーを設定
	// Basic認証用のAuthorizationヘッダーを追加（curlの --basic -u ユーザー名:パスワード に相当）
	request.SetBasicAuth("username", "password")

	// 方法3: AddCookie()メソッドを使用してCookieを手動で追加
	// サーバーから受け取っていないCookieも自由に送信可能
	cookie := &http.Cookie{
		Name:  "test",
		Value: "value",
	}
	request.AddCookie(cookie)

	// リクエストをダンプして表示（送信前）
	// 設定したヘッダーフィールドが含まれていることを確認
	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		panic(err)
	}
	log.Println("リクエスト：")
	log.Println(string(dump))

	// リクエストを送信
	resp, err := client.Do(request)
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

	// レスポンス全体（ヘッダーとボディ）をダンプ
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// レスポンスのダンプに失敗した場合はパニック
		panic(err)
	}

	// ダンプしたレスポンスを文字列として出力
	log.Println("レスポンス：")
	log.Println(string(respDump))
}
