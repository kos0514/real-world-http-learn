// post_form.go - フォームデータを使用したHTTP POSTリクエストの例
// application/x-www-form-urlencodedフォーマットでデータを送信します
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// POSTリクエストで送信するフォームデータを作成
	// url.Values型はmap[string][]stringと同様の構造で、複数の値を持つことができます
	values := url.Values{
		"ampersand":   {"Tom&Jerry"},        // & は %26 に変換される
		"at":          {"user@example.com"}, // @ は %40 に変換される
		"equals":      {"key=value"},        // = は %3D に変換される
		"exclamation": {"Hello!World"},      // ! は %21 に変換される
		"space":       {"hello world"},      // スペースは + に変換される
		"test":        {"value"},            // キー"test"に対して値"value"を設定
	}

	// エンコード前とエンコード後の値を表示（アルファベット順）
	fmt.Println("【変身前】")
	fmt.Println("ampersand: Tom&Jerry")
	fmt.Println("at: user@example.com")
	fmt.Println("equals: key=value")
	fmt.Println("exclamation: Hello!World")
	fmt.Println("space: hello world")
	fmt.Println("test: value")

	fmt.Println("\n【変身後（リクエストボディ）】")
	fmt.Println(values.Encode())

	// ローカルサーバーにHTTP POSTリクエストを送信
	// http.PostForm関数は内部でContent-Type: application/x-www-form-urlencodedヘッダーを設定します
	resp, err := http.PostForm("http://localhost:18888", values)
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はパニック
		panic(err)
	}
	// 注意: 本来はrespがnilでないことを確認し、defer resp.Body.Close()を呼び出すべきです
	// 接続リソースをリークさせないために、レスポンスボディは必ず閉じる必要があります

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("Status:", resp.Status)
}
