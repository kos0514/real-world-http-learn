// http1.0/server.go は単純なHTTPサーバーを実装します。
// このサーバーは受信したHTTPリクエストの内容をコンソールに表示し、
// シンプルなHTMLレスポンスを返します。
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// handler はHTTPリクエストを処理する関数です。
// 受信したリクエストの内容をダンプして表示し、
// 単純なHTML応答を返します。
//
// パラメータ:
//   - w: HTTPレスポンスを書き込むためのResponseWriter
//   - r: 処理すべきHTTPリクエスト
func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(dump))
	fmt.Fprintf(w, "<html><body>hello</body></html>\n")
}

// main はプログラムのエントリーポイントです。
// ポート18888でHTTPサーバーを起動し、
// すべてのパスに対して handler 関数を呼び出します。
func main() {
	var httpServer http.Server
	http.HandleFunc("/", handler)
	log.Println("start http listening :18888")
	httpServer.Addr = ":18888"
	log.Println(httpServer.ListenAndServe())
}
