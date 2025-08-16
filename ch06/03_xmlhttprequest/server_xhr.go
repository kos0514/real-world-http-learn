// ch06/03_xmlhttprequest/server_xhr.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// トップ + 静的 UI
	mux.HandleFunc("/", uiIndex)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("ch06/03_xmlhttprequest/public"))))

	// 基本 API
	mux.HandleFunc("/json", handleJSON)                  // GET: JSON を返す（HttpOnly Cookie 付与）
	mux.HandleFunc("/echo", handleEcho)                  // 任意メソッド: 送られた内容/ヘッダ/クッキーを JSON で反射
	mux.HandleFunc("/upload", handleUpload)              // POST: multipart/form-data を受信して要約を返す
	mux.HandleFunc("/redirect", handleRedirect)          // GET: 302 → /json
	mux.HandleFunc("/headers", handleHeaders)            // GET: リクエストヘッダを JSON で返す
	mux.HandleFunc("/poll", handlePoll)                  // GET: 簡易ポーリング（現在時刻）
	mux.HandleFunc("/set_cookie", handleSetCookie)       // GET: 非 HttpOnly Cookie 設定（document.cookie 実験用）
	mux.HandleFunc("/clear_cookie", handleClearCookie)   // GET: クッキー削除
	mux.HandleFunc("/clear_session", handleClearSession) // GET: HttpOnly demo_session 削除

	// Comet (ロングポーリング)
	mux.HandleFunc("/comet/send", cometSend) // POST: msg を送信
	mux.HandleFunc("/comet/recv", cometRecv) // GET: 長時間待機して msg を受け取る

	// CORS デモ
	mux.HandleFunc("/cors/json", corsJSON)                  // GET: Access-Control-Allow-Origin: *
	mux.HandleFunc("/cors/with_credentials", corsWithCreds) // GET: 資格情報付き CORS の例
	mux.HandleFunc("/cors/preflight", corsPreflight)        // OPTIONS: プリフライト応答

	addr := ":18063"
	srv := &http.Server{Addr: addr, Handler: logging(mux), ReadHeaderTimeout: 5 * time.Second}

	// 起動ログ（ローカル URL を見やすく）
	host := "localhost"
	port := strings.TrimPrefix(addr, ":")
	if h, p, err := net.SplitHostPort(addr); err == nil {
		if h == "" || h == "0.0.0.0" || h == "::" {
			host = "localhost"
		} else {
			host = h
		}
		port = p
	}
	browseURL := fmt.Sprintf("http://%s:%s/", host, port)
	log.Printf("xhr server listening on %s", addr)
	log.Printf("ブラウザで開く: %s", browseURL)

	log.Fatal(srv.ListenAndServe())
}

// ------------- UI -------------
func uiIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>ch06/03_xmlhttprequest</title>
<h1>ch06/03 XMLHttpRequest 学習ページ</h1>
<p>各デモページ（/public/）からブラウザだけで操作・確認できます。</p>
<ul>
  <li><a href="/public/01_basic_get.html">01 Basic GET</a></li>
  <li><a href="/public/02_post_json.html">02 POST JSON</a></li>
  <li><a href="/public/03_formdata_upload.html">03 FormData アップロード</a></li>
  <li><a href="/public/04_custom_headers.html">04 カスタムヘッダ</a></li>
  <li><a href="/public/05_redirect.html">05 リダイレクト</a></li>
  <li><a href="/public/06_polling.html">06 ポーリング</a></li>
  <li><a href="/public/07_long_polling.html">07 ロングポーリング（Comet）</a></li>
  <li><a href="/public/08_cors.html">08 CORS / withCredentials</a></li>
  <li><a href="/public/09_security_notes.html">09 セキュリティの注意</a></li>
</ul>
`)
}

// ------------- Handlers: 基本 -------------
func handleJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// HttpOnly Cookie を付与（document.cookie からは読めない）
	http.SetCookie(w, &http.Cookie{Name: "demo_session", Value: "abc123", Path: "/", HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode})
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "hello",
		"method":  r.Method,
		"now":     time.Now().Format(time.RFC3339Nano),
	})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := io.ReadAll(r.Body)
	_ = r.ParseMultipartForm(32 << 20) // 32MB
	cookies := map[string]string{}
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}
	resp := map[string]any{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.Query(),
		"headers": r.Header,
		"cookies": cookies,
		"length":  len(body),
	}
	if ct := r.Header.Get("Content-Type"); strings.HasPrefix(ct, "application/json") {
		var j any
		_ = json.Unmarshal(body, &j)
		resp["json"] = j
	}
	if len(r.PostForm) > 0 {
		resp["form"] = r.PostForm
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := r.ParseMultipartForm(64 << 20); err != nil { // 64MB
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	files := []map[string]any{}
	for key, fhs := range r.MultipartForm.File {
		for _, fh := range fhs {
			files = append(files, map[string]any{"field": key, "filename": fh.Filename, "size": fh.Size})
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"form":  r.MultipartForm.Value,
		"files": files,
	})
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/json", http.StatusFound) // 302
}

func handleHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(r.Header)
}

func handlePoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"now": time.Now().Format(time.RFC3339Nano)})
}

func handleSetCookie(w http.ResponseWriter, r *http.Request) {
	// 非 HttpOnly（document.cookie から見える）
	http.SetCookie(w, &http.Cookie{Name: "visible_token", Value: "v1", Path: "/", HttpOnly: false})
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "set visible_token=v1")
}

func handleClearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "visible_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "cleared visible_token")
}

// ------------- Comet: ロングポーリング -------------
var (
	cometMu   sync.Mutex
	cometSubs = map[int]chan string{} // 簡易: 接続ごとのチャネル
	cometSeq  = 0
)

func cometSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	msg := r.Form.Get("msg")
	if msg == "" {
		http.Error(w, "msg required", http.StatusBadRequest)
		return
	}
	// 全購読者に配信
	cometMu.Lock()
	for _, ch := range cometSubs {
		select {
		case ch <- msg:
		default:
		}
	}
	cometMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cometRecv(w http.ResponseWriter, r *http.Request) {
	// ロングポーリング: 25秒 or 1件受信まで待つ
	flusher, _ := w.(http.Flusher)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	ch := make(chan string, 1)
	cometMu.Lock()
	cometSeq++
	id := cometSeq
	cometSubs[id] = ch
	cometMu.Unlock()

	defer func() {
		cometMu.Lock()
		delete(cometSubs, id)
		cometMu.Unlock()
	}()

	select {
	case msg := <-ch:
		_ = json.NewEncoder(w).Encode(map[string]any{"msg": msg, "at": time.Now().Format(time.RFC3339Nano)})
		if flusher != nil {
			flusher.Flush()
		}
	case <-time.After(25 * time.Second):
		w.WriteHeader(http.StatusNoContent) // 204 (クライアントは再接続)
	}
}

// ------------- CORS -------------
func corsJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"ok": "cors-any"})
}

func corsWithCreds(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// 資格情報を許可する場合はワイルドカード不可。実オリジンを返す。
	if origin == "" {
		origin = "http://localhost:18063"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	http.SetCookie(w, &http.Cookie{Name: "cross_demo", Value: "1", Path: "/", HttpOnly: true})
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"ok": "with-credentials"})
}

func corsPreflight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodOptions {
		http.Error(w, "OPTIONS only", http.StatusMethodNotAllowed)
		return
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:18063"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, MyHeader")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

// ------------- 共通 -------------
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func handleClearSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "demo_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "cleared demo_session")
}
