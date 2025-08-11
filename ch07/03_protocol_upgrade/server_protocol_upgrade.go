package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// /upgrade へのリクエストを HTTP/1.1 の Upgrade で独自プロトコルに切り替える
// ---- 独自プロトコル用のヘルパー関数群 ----
// 指定の読みルールで 1..40 の日本語読みを生成する（非アホ時の表示）
func readingJP(n int) string {
	units := map[int]string{
		1: "いち", 2: "にー", 3: "さん", 4: "よん", 5: "ごー",
		6: "ろく", 7: "なな", 8: "はち", 9: "くー",
	}
	tens := map[int]string{
		1: "じゅう", 2: "にじゅう", 3: "さんじゅう", 4: "よんじゅう",
	}
	if n < 10 {
		return units[n]
	}
	t, u := n/10, n%10
	base := tens[t]
	if u == 0 {
		return base
	}
	return base + units[u]
}

func contains3(n int) bool {
	for x := n; x > 0; x /= 10 {
		if x%10 == 3 {
			return true
		}
	}
	return false
}

// yomi の語末母音を返す。末尾の「ー」「ん」は無視して判定する。
func lastVowel(yomi string) rune {
	rs := []rune(yomi)
	i := len(rs) - 1
	for i >= 0 {
		if rs[i] == 'ー' || rs[i] == 'ん' {
			i--
			continue
		}
		break
	}
	if i < 0 {
		return 'あ' // フォールバック
	}
	return vowelOfKana(rs[i])
}

// 仮名1文字の母音を求める（五十音の行ベース）
func vowelOfKana(k rune) rune {
	switch k {
	case 'あ', 'か', 'さ', 'た', 'な', 'は', 'ま', 'や', 'ら', 'わ', 'が', 'ざ', 'だ', 'ば', 'ぱ', 'ぁ', 'ゃ':
		return 'あ'
	case 'い', 'き', 'し', 'ち', 'に', 'ひ', 'み', 'り', 'ぎ', 'じ', 'ぢ', 'び', 'ぴ', 'ぃ':
		return 'い'
	case 'う', 'く', 'す', 'つ', 'ぬ', 'ふ', 'む', 'ゆ', 'る', 'ぐ', 'ず', 'づ', 'ぶ', 'ぷ', 'ぅ', 'ゅ':
		return 'う'
	case 'え', 'け', 'せ', 'て', 'ね', 'へ', 'め', 'れ', 'げ', 'ぜ', 'で', 'べ', 'ぺ', 'ぇ':
		return 'え'
	case 'お', 'こ', 'そ', 'と', 'の', 'ほ', 'も', 'よ', 'ろ', 'ご', 'ぞ', 'ど', 'ぼ', 'ぽ', 'ぉ', 'ょ':
		return 'お'
	default:
		return 'あ'
	}
}

// アホ尾部（小さい母音 + 同母音×15）
func ahoTail(v rune) string {
	switch v {
	case 'あ':
		return "ぁ" + strings.Repeat("あ", 15)
	case 'い':
		return "ぃ" + strings.Repeat("い", 15)
	case 'う':
		return "ぅ" + strings.Repeat("う", 15)
	case 'え':
		return "ぇ" + strings.Repeat("え", 15)
	default: // 'お'
		return "ぉ" + strings.Repeat("お", 15)
	}
}

// 1 行分の出力を作る。アホ時は「ー」を全削除した基礎表示に尾部を追加。
// 読みが「ん」で終わる場合は尾部を「ん」の直前に挿入する。
func lineFor(n int) string {
	y := readingJP(n)
	isAho := contains3(n) || n%3 == 0
	if !isAho {
		return y
	}
	v := lastVowel(y)
	base := strings.ReplaceAll(y, "ー", "")
	tail := ahoTail(v)
	// 読みが「ん」で終わるなら尾部を直前に挿入
	if strings.HasSuffix(y, "ん") {
		brs := []rune(base)
		if len(brs) > 0 && brs[len(brs)-1] == 'ん' {
			return string(brs[:len(brs)-1]) + tail + "ん"
		}
	}
	return base + tail
}

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
	// サーバーは 1..40 を仕様に従って送信する（クライアントからの受信はしない）
	writer := rw.Writer // *bufio.Writer

	for i := 1; i <= 40; i++ {
		out := lineFor(i)
		if _, err := fmt.Fprintf(writer, "%s\n", out); err != nil {
			log.Println("write error:", err)
			return
		}
		if err := writer.Flush(); err != nil { // バッファをフラッシュ
			log.Println("flush error:", err)
			return
		}
		log.Printf("-> %d: %s", i, out)
		time.Sleep(200 * time.Millisecond)
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
