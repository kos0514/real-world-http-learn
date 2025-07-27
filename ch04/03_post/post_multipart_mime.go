// post_multipart_mime.go - MIMEヘッダーを明示的に設定したマルチパートフォームデータの例
// multipart/form-dataフォーマットでテキストフィールドとファイルを一緒に送信します
// このファイルでは、textproto.MIMEHeaderを使用してContent-TypeやContent-Dispositionを
// 明示的に設定する方法を示しています
package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
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

	// ファイルフィールドをフォームに追加（MIMEヘッダーを明示的に設定）
	// HTMLフォームの <input type="file" name="thumbnail"> に相当
	// CreateFormFileの代わりにCreatePartを使用し、MIMEヘッダーを手動で設定
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "image/jpeg")
	part.Set("Content-Disposition", `form-data; name="thumbnail"; filename="hello_world_small.jpg"`)
	fileWriter, err := writer.CreatePart(part)
	if err != nil {
		// フォームパート作成に失敗した場合はパニック
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

	// respがnilでないことを確認してからBodyをクローズ
	if resp != nil {
		defer resp.Body.Close()
	} else {
		log.Fatalf("レスポンスがnilです")
		return
	}

	// レスポンスのステータス（例: "200 OK"）を出力
	log.Println("Status:", resp.Status)
}
