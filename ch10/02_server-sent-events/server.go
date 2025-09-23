package main

import (
	_ "embed"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

//go:embed index.html
var html []byte

// HTMLをブラウザに送信
func handlerHtml(w http.ResponseWriter, r *http.Request) {
	// Pusherにキャスト可能であればプッシュする
	w.Header().Add("Content-Type", "text/html")
	w.Write(html)
}

// 素数をブラウザに送信
func handlerPrimeSSE(w http.ResponseWriter, r *http.Request) {
	c := http.NewResponseController(w)
	// 接続断の検知用にコンテキストを取得
	ctx := r.Context()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var num int64 = 1
	for id := 1; id <= 100; id++ {
		// 通信が切れても終了
		select {
		case <-ctx.Done():
			fmt.Println("Connection closed from client")
			return
		default:
			// do nothing
		}
		for {
			num++
			// 確率論的に素数を求める
			if big.NewInt(num).ProbablyPrime(20) {
				fmt.Println(num)
				fmt.Fprintf(w, "data: {\"id\": %d, \"number\": %d}\n\n", id, num)
				c.Flush()
				time.Sleep(time.Second)
				break
			}
		}
		time.Sleep(time.Second)
	}
	// 100個超えたら送信終了
	fmt.Println("Connection closed from server")
}

func main() {
	http.HandleFunc("/", handlerHtml)
	http.HandleFunc("/prime", handlerPrimeSSE)
	fmt.Println("start http listening :18888（http://localhost:18888/）")
	if err := http.ListenAndServe(":18888", nil); err != nil {
		fmt.Println(err)
	}
}
