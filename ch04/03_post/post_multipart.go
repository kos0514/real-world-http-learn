// post_multipart.go - マルチパートフォームデータを使用したHTTP POSTリクエストの例
// multipart/form-dataフォーマットでテキストフィールドとファイルを一緒に送信します
// このフォーマットはWebフォームでファイルアップロードを行う際に使用されます
package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	// バッファを作成してマルチパートフォームデータを構築
	var buffer bytes.Buffer
	// マルチパートライターを作成（バウンダリ文字列は自動的に生成される）
	writer := multipart.NewWriter(&buffer)

	// テキストフィールドをフォームに追加
	// HTMLフォームの <input type="text" name="name" value="Stevie Wonder"> に相当
	writer.WriteField("name", "Stevie Wonder")

	// ファイルフィールドをフォームに追加
	// HTMLフォームの <input type="file" name="thumbnail"> に相当
	fileWriter, err := writer.CreateFormFile("thumbnail", "hello_world_small.jpg")
	if err != nil {
		// フォームファイルフィールド作成に失敗した場合はパニック
		panic(err)
	}

	// アップロードするファイルを開く
	readFile, err := os.Open("ch04/03_post/hello_world_small.jpg")
	if err != nil {
		// ファイル読み込み失敗した場合はパニック
		panic(err)
	}
	// 関数終了時にファイルを必ず閉じる
	defer readFile.Close()

	// ファイルの内容をフォームのファイルフィールドにコピー
	io.Copy(fileWriter, readFile)

	// マルチパートフォームを完成させる（必須）
	// これによりフォームの終端が書き込まれる
	writer.Close()

	// マルチパートフォームデータをPOSTリクエストのボディとして送信
	// Content-Typeヘッダーには writer.FormDataContentType() で
	// 「multipart/form-data; boundary=...」形式の値が設定される
	resp, err := http.Post("http://localhost:18888", writer.FormDataContentType(), &buffer)
	if err != nil {
		// リクエスト送信に失敗した場合はパニック
		panic(err)
	}

	// 注意: 本来はrespがnilでないことを確認し、defer resp.Body.Close()を呼び出すべき
	// 接続リソースをリークさせないために、レスポンスボディは必ず閉じる必要がある

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("Status:", resp.Status)
}
