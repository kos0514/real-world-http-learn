// get_with_query.go - パーセントエンコーディングの実例を示すHTTP GETリクエストの例
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// 変身の魔法一覧：特殊文字とそのエンコード後の形式
	// | 危険な文字 | 変身後 | 理由 |
	// |-----------|--------|------|
	// | スペース   | +      | application/x-www-form-urlencodedでは常にプラス記号に変身！ |
	// | &         | %26    | 項目区切りと混同防止 |
	// | =         | %3D    | キー・値結合子と混同防止 |
	// | @         | %40    | メールアドレス等で使用 |
	// | !         | %21    | 特殊記号の安全な変身 |

	// 特殊文字を含むクエリパラメータを作成（アルファベット順）
	values := url.Values{
		"ampersand":   {"Tom&Jerry"},        // & は %26 に変換される
		"at":          {"user@example.com"}, // @ は %40 に変換される
		"equals":      {"key=value"},        // = は %3D に変換される
		"exclamation": {"Hello!World"},      // ! は %21 に変換される
		"space":       {"hello world"},      // スペースは + に変換される
	}

	// エンコード前とエンコード後の値を表示（アルファベット順）
	fmt.Println("【変身前】")
	fmt.Println("ampersand: Tom&Jerry")
	fmt.Println("at: user@example.com")
	fmt.Println("equals: key=value")
	fmt.Println("exclamation: Hello!World")
	fmt.Println("space: hello world")

	fmt.Println("\n【変身後】")
	fmt.Println(values.Encode())

	// URLにクエリパラメータを追加してHTTP GETリクエストを送信
	resp, err := http.Get("http://localhost:18888" + "?" + values.Encode())
	if err != nil {
		// リクエスト送信中にエラーが発生した場合はログに記録して終了
		log.Fatalf("リクエスト送信エラー: %v", err)
	}
	// respがnilでないことを確認してからBodyをクローズ
	if resp != nil {
		defer resp.Body.Close()
	} else {
		log.Fatalf("レスポンスがnilです")
		return
	}

	// レスポンスボディを読み込む
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// ボディの読み込み中にエラーが発生した場合はログに記録して終了
		log.Fatalf("レスポンスボディの読み込みエラー: %v", err)
	}

	// レスポンスボディを文字列として出力
	log.Println(string(body))
}
