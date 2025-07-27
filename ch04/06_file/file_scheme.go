// file_scheme.go - ローカルファイルにアクセスするHTTPクライアントの例
// file://スキームを使用してローカルファイルシステム上のファイルを読み込む方法を示します
// 通常のHTTPクライアントを使用して、ネットワーク経由ではなくファイルシステムから直接ファイルを取得できます
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	// カスタムトランスポートを作成
	transport := &http.Transport{}

	// "file"プロトコルハンドラを登録
	// http.Dir(".")は現在のディレクトリをルートとするファイルシステムを指定
	// http.NewFileTransportはファイルシステムからファイルを提供するトランスポートを作成
	transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(".")))

	// カスタムトランスポートを使用するHTTPクライアントを作成
	// このクライアントは"file://"スキームを使用したURLを処理できるようになります
	client := http.Client{
		Transport: transport,
	}

	// ローカルファイルにHTTP GETリクエストを送信
	// "file://./main.go"は現在のディレクトリ内のmain.goファイルを指定
	resp, err := client.Get("file://./main.go")
	if err != nil {
		// ファイルアクセスに失敗した場合はパニック
		panic(err)
	}

	// 注意: 本来はrespがnilでないことを確認し、defer resp.Body.Close()を呼び出すべき
	// 接続リソースをリークさせないために、レスポンスボディは必ず閉じる必要があります

	// レスポンス全体（ヘッダーとボディ）をダンプ
	// ボディにはファイルの内容が含まれます
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// レスポンスのダンプに失敗した場合はパニック
		panic(err)
	}

	// ダンプしたレスポンスを文字列として出力
	log.Println(string(dump))
}
