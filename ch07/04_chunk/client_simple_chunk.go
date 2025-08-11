package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
)

// シンプルな chunked レスポンス受信クライアント。
// サーバーの /chunked から届くデータを行単位で逐次読み取り、ログに出力する。
// ポイント:
// - クライアント側は「チャンク境界」を意識する必要はなく、通常どおり Body を読み進めればよい。
// - サーバー側が逐次 Flush していれば、こちらも到着した分だけすぐに ReadBytes で受け取れる。
func main() {
	// サーバーへ GET。server_chunk.go が :18888/chunked で待ち受けている想定。
	resp, err := http.Get("http://localhost:18888/chunked")
	if err != nil {
		log.Fatal(err)
	}
	// 応答ボディは最後に必ず Close する
	defer resp.Body.Close()

	// 改行（\n）までを 1 単位として読み取る
	reader := bufio.NewReader(resp.Body)
	for {
		// 1 行分を読み込む（末尾の \n を含む）。チャンクの切れ目は透過的に処理される。
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			// サーバーが送信を終えた（接続/本文が終端に達した）
			break
		}
		// 受け取った行をログ出力（末尾の改行などを除去）
		log.Println(string(bytes.TrimSpace(line)))
	}
}
