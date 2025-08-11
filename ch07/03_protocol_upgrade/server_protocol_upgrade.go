package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// /upgrade へのリクエストを HTTP/1.1 の Upgrade で独自プロトコルに切り替える
func handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	// このエンドポイントでは Upgrade 要求以外は受け付けない
	if r.Header.Get("Connection") != "Upgrade" || r.Header.Get("Upgrade") != "MyProtocol" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Upgrade to MyProtocol required\n"))
		return
	}
	log.Println("Upgrade to MyProtocol requested")

	// レスポンスを書き出すために低層ソケットへ切替（Hijack）
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server does not support hijacking", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		log.Println("hijack error:", err)
		return
	}
	defer conn.Close()

	// 101 Switching Protocols を返して HTTP を終了
	response := http.Response{
		StatusCode: http.StatusSwitchingProtocols, // 101
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	response.Header.Set("Upgrade", "MyProtocol")
	response.Header.Set("Connection", "Upgrade")
	if err := response.Write(conn); err != nil {
		log.Println("write 101 response error:", err)
		return
	}

	// ここからは HTTP ではなく、独自プロトコル（行単位のテキスト）
	// サーバーは 1..10 を送信し、クライアントからの応答行を受け取る
	reader := rw.Reader            // *bufio.Reader
	writer := rw.Writer            // *bufio.Writer
	_ = bufio.ErrInvalidUnreadByte // 参照保持で bufio が未使用にならないように（ツールの最適化対策）

	for i := 1; i <= 10; i++ {
		// 行で送信
		if _, err := fmt.Fprintf(writer, "%d\n", i); err != nil {
			log.Println("write error:", err)
			return
		}
		if err := writer.Flush(); err != nil { // バッファをフラッシュ
			log.Println("flush error:", err)
			return
		}
		log.Println("->", i)

		// クライアントから 1 行受信
		recv, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("client closed")
			}
			return
		}
		log.Printf("<- %s", string(recv))

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	// /upgrade のみを扱う簡易 HTTP サーバー
	http.HandleFunc("/upgrade", handlerUpgrade)

	addr := ":18888"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
