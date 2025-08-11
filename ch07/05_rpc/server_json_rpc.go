package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"real-world-http-learn/ch07/05_rpc/rpcdef"
)

// Calculator は RPC で公開するメソッド（Multiply）を持つ型です。
// net/rpc の制約として、
// - 受信側の型名（ここでは Calculator）はエクスポートされている必要があります（先頭大文字）。
// - 公開するメソッド名（Multiply）もエクスポートされている必要があります。
// - 引数は1つ（リクエスト）、戻り値は error、かつ結果を格納するためのポインタ引数をとるシグネチャにします。
type Calculator int

// Multiply は 2 つの整数を掛け算し、その結果を result に書き込みます。
// 第1引数の型には、クライアントと共通の rpcdef.Args を使用します。
func (c *Calculator) Multiply(args rpcdef.Args, result *int) error {
	log.Printf("Multiply called: %d, %d\n", args.A, args.B)
	*result = args.A * args.B
	return nil
}

func main() {
	// サービス実体を作成し、RPC サーバに登録します。
	calculator := new(Calculator)
	server := rpc.NewServer()
	server.Register(calculator)

	// 今回は HTTP サーバは使わず、下の TCP リスナー + JSON-RPC コーデックで処理します。
	// （http.Handle の行は残していますが、本例では使用しません）
	http.Handle(rpc.DefaultRPCPath, server)

	log.Println("start tcp listening :18888")
	listener, err := net.Listen("tcp", ":18888")
	if err != nil {
		panic(err)
	}

	// 接続を受け付け、各接続ごとに JSON-RPC でリクエストを処理します。
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
