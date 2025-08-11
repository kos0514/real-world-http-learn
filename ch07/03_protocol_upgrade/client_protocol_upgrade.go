package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	// TCP ソケットをオープン
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	conn, err := dialer.Dial("tcp", "localhost:18888")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Upgrade ヘッダ付きで HTTP リクエストを作成し、ソケットに直接書き込み
	req, _ := http.NewRequest("GET", "http://localhost:18888/upgrade", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "MyProtocol")
	if err := req.Write(conn); err != nil {
		log.Fatal(err)
	}

	// レスポンスを解析し、101 Switching Protocols を期待
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Status:", resp.Status)
	log.Println("Headers:", resp.Header)
	if resp.StatusCode != http.StatusSwitchingProtocols {
		log.Fatalf("upgrade failed: %s", resp.Status)
	}
	// 以後は独自プロトコル。resp.Body は触らない（HTTP ではない）

	// サーバから 1 行受信 -> 1 行送信を繰り返す
	counter := 10
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println("<-", string(bytes.TrimSpace(data)))

		if counter <= 0 {
			break
		}
		if _, err := fmt.Fprintf(conn, "%d\n", counter); err != nil {
			log.Fatal(err)
		}
		fmt.Println("->", counter)
		counter--
	}
}
