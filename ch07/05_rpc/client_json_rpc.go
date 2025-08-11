package main

import (
	"log"
	"net/rpc/jsonrpc"

	"real-world-http-learn/ch07/05_rpc/rpcdef"
)

// クライアント側: JSON-RPC を使ってサーバの Calculator.Multiply を呼び出すサンプルです。
// ポイント:
// - rpcdef.Args はクライアント/サーバ共通の引数型。重複定義を避けるためサブパッケージに切り出しています。
// - jsonrpc.Dial で JSON-RPC 用の TCP コネクションを張ります。
// - client.Call で「受信側タイプ名.メソッド名」を指定して RPC を呼び出します。
func main() {
	// サーバ(:18888)へ TCP で接続し、JSON-RPC 用のクライアントを作成
	client, err := jsonrpc.Dial("tcp", "localhost:18888")
	if err != nil {
		// 接続に失敗した場合は致命的エラーとして終了します
		panic(err)
	}

	// サーバに渡す引数（4 と 5 を掛け算）
	args := &rpcdef.Args{A: 4, B: 5}

	// RPC 呼び出しの戻り値を受け取る変数
	var result int

	// 「Calculator.Multiply」を呼び出します。第3引数は結果を書き込むポインタです。
	if err := client.Call("Calculator.Multiply", args, &result); err != nil {
		panic(err)
	}

	log.Printf("4 x 5 = %d\n", result)
}
