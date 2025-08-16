// ch06/02_resume_range/server_resume_range.go
// 1つのサーバーに Range/If-Range/複数範囲/Accept-Ranges: none/gzip を集約したデモ実装。
// ブラウザだけで確認できる UI も同梱します。
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 擬似コンテンツ（20MB）
var (
	totalSize   int64 = 20 * 1024 * 1024
	content     []byte
	contentETag string
	lastMod     time.Time
	flipVersion bool

	gzipBytes   []byte
	gzipETag    string
	gzipLastMod time.Time
)

// 現行バージョン（/flip_etag で変化を演出）
func currentVersion() (etag string, lm time.Time) {
	if flipVersion {
		// 弱い ETag っぽく見える値にして差分を発生させやすくする
		return "W/\"" + strings.Trim(contentETag, "\"") + "\"", time.Now()
	}
	return contentETag, lastMod
}

func main() {
	// 20MB の規則的なデータを生成（i%256）
	content = make([]byte, totalSize)
	for i := int64(0); i < totalSize; i++ {
		content[i] = byte(i % 256)
	}
	contentETag = strongETag(content)
	lastMod = time.Now().Add(-1 * time.Hour)

	gzipBytes = mustGzip(content)
	gzipETag = strongETag(gzipBytes)
	gzipLastMod = lastMod

	mux := http.NewServeMux()
	mux.HandleFunc("/", uiIndex)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("ch06/02_resume_range/public"))))

	mux.HandleFunc("/file", handleFile)
	mux.HandleFunc("/file_gzip", handleFileGzip)
	mux.HandleFunc("/file_none", handleFileNone)
	mux.HandleFunc("/flip_etag", func(w http.ResponseWriter, r *http.Request) {
		flipVersion = !flipVersion
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "flipVersion=%v\n", flipVersion)
	})

	addr := ":18062"
	srv := &http.Server{Addr: addr, Handler: logging(mux), ReadHeaderTimeout: 5 * time.Second}

	// 起動時にブラウザで開ける URL を出力
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
	log.Printf("resume-range server listening on %s", addr)
	log.Printf("ブラウザで開く: %s", browseURL)

	log.Fatal(srv.ListenAndServe())
}

// ------------- UI -------------
func uiIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>ch06/02_resume_range</title>
<h1>ch06/02 再開ダウンロードと Range の学習ページ</h1>
<p>このトップページから各「詳細解説ページ」へ移動できます。ブラウザだけで操作・確認できます。</p>

<section>
  <h2>詳細解説ページ</h2>
  <ul>
    <li>
      <a href="/public/single.html">単一範囲</a>
      — Range: bytes=A-B を1回ずつ取得し、クリックを積み上げて1つのファイルに結合します。
    </li>
    <li>
      <a href="/public/multipart.html">複数範囲</a>
      — 1リクエストで複数の範囲を取得（multipart/byteranges）。繰り返して全体を組み上げます。
    </li>
    <li>
      <a href="/public/if-range.html">If-Range</a>
      — ETag/日時が一致なら206の部分返却。不一致なら200で全体にフォールバックする流れを体験します。
    </li>
    <li>
      <a href="/public/416.html">416</a>
      — 不正な範囲指定により 416 Range Not Satisfiable と Content-Range: bytes */total を確認します。
    </li>
    <li>
      <a href="/public/none.html">Accept-Ranges: none</a>
      — サーバーが Range を受け付けない場合の挙動（Range を無視して 200 全体）を確認します。
    </li>
    <li>
      <a href="/public/gzip.html">gzip Range</a>
      — Content-Encoding: gzip（圧縮後のバイト列）に対する Range の挙動を確認します。
    </li>
  </ul>
</section>

<section>
  <h2>並列取得体験</h2>
  <p>
    <a href="/public/parallel.html">/public/parallel.html</a>
    — HEAD でサイズ/ETag を取得してチャンク分割、複数の Range を並列で取得し、最後に結合して保存します。
  </p>
</section>

<hr>
<p>API（参考）</p>
<ul>
  <li><code>GET /file</code> … Range/If-Range/複数範囲</li>
  <li><code>GET /file_gzip</code> … Content-Encoding: gzip（圧縮後バイトに対する Range）</li>
  <li><code>GET /file_none</code> … Accept-Ranges: none（Range 無視）</li>
  <li><code>GET /flip_etag</code> … ETag/Last-Modified を変更して If-Range 不一致を発生させる</li>
</ul>
`)
}

// ------------- ハンドラ -------------
// /file: Range / If-Range / 複数範囲
func handleFile(w http.ResponseWriter, r *http.Request) {
	etag, lm := currentVersion()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", lm.UTC().Format(http.TimeFormat))

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
		w.WriteHeader(http.StatusOK)
		return
	}

	rangeHdr := r.Header.Get("Range")
	if rangeHdr == "" { // Range 指定なし → 全体
		w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
		return
	}

	// If-Range 判定：一致しなければ Range を無視して 200 Full
	if ifr := r.Header.Get("If-Range"); ifr != "" {
		if ifRangeMismatch(ifr, etag, lm) {
			w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
			return
		}
	}

	unit, ranges, ok := parseRangeValue(rangeHdr, totalSize)
	if !ok || unit != "bytes" || len(ranges) == 0 {
		write416(w, totalSize)
		return
	}

	if len(ranges) == 1 { // 単一範囲
		r := ranges[0]
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, totalSize))
		w.Header().Set("Content-Length", strconv.FormatInt(r.length(), 10))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(content[r.start : r.end+1])
		return
	}

	// 複数範囲 → multipart/byteranges
	boundary := "THIS_STRING_SEPARATES"
	var buf bytes.Buffer
	for _, br := range ranges {
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		buf.WriteString("Content-Type: application/octet-stream\r\n")
		fmt.Fprintf(&buf, "Content-Range: bytes %d-%d/%d\r\n\r\n", br.start, br.end, totalSize)
		buf.Write(content[br.start : br.end+1])
		buf.WriteString("\r\n")
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	w.Header().Set("Content-Type", "multipart/byteranges; boundary="+boundary)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusPartialContent)
	_, _ = w.Write(buf.Bytes())
}

// /file_gzip: 圧縮後のバイト列に対して Range を解釈
func handleFileGzip(w http.ResponseWriter, r *http.Request) {
	etag := gzipETag
	lm := gzipLastMod
	if flipVersion { // flip で変化を演出
		etag = "W/\"" + gzipETag + "\""
		lm = time.Now()
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", lm.UTC().Format(http.TimeFormat))

	gzTotal := int64(len(gzipBytes))
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.FormatInt(gzTotal, 10))
		w.WriteHeader(http.StatusOK)
		return
	}

	rangeHdr := r.Header.Get("Range")
	if rangeHdr == "" { // 全体（圧縮済みバイト）
		w.Header().Set("Content-Length", strconv.FormatInt(gzTotal, 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(gzipBytes)
		return
	}

	if ifr := r.Header.Get("If-Range"); ifr != "" {
		if ifRangeMismatch(ifr, etag, lm) {
			w.Header().Set("Content-Length", strconv.FormatInt(gzTotal, 10))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(gzipBytes)
			return
		}
	}

	unit, ranges, ok := parseRangeValue(rangeHdr, gzTotal)
	if !ok || unit != "bytes" || len(ranges) == 0 {
		write416(w, gzTotal)
		return
	}
	if len(ranges) == 1 {
		r := ranges[0]
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, gzTotal))
		w.Header().Set("Content-Length", strconv.FormatInt(r.length(), 10))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(gzipBytes[r.start : r.end+1])
		return
	}

	boundary := "THIS_STRING_SEPARATES"
	var buf bytes.Buffer
	for _, br := range ranges {
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		buf.WriteString("Content-Type: application/octet-stream\r\n")
		fmt.Fprintf(&buf, "Content-Range: bytes %d-%d/%d\r\n\r\n", br.start, br.end, gzTotal)
		buf.Write(gzipBytes[br.start : br.end+1])
		buf.WriteString("\r\n")
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	w.Header().Set("Content-Type", "multipart/byteranges; boundary="+boundary)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusPartialContent)
	_, _ = w.Write(buf.Bytes())
}

// /file_none: Range を受け付けない
func handleFileNone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "none")
	w.Header().Set("ETag", contentETag)
	w.Header().Set("Last-Modified", lastMod.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// ------------- ユーティリティ -------------
func strongETag(b []byte) string {
	sum := sha256.Sum256(b)
	return "\"" + hex.EncodeToString(sum[:]) + "\""
}

func mustGzip(b []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(b)
	_ = zw.Close()
	return buf.Bytes()
}

type byteRange struct{ start, end int64 }

func (r byteRange) length() int64 { return r.end - r.start + 1 }

// parseRangeValue parses a Range header value (e.g., "bytes=0-1023,2000-")
func parseRangeValue(val string, total int64) (unit string, ranges []byteRange, ok bool) {
	v := strings.TrimSpace(val)
	parts := strings.SplitN(v, "=", 2)
	if len(parts) != 2 {
		return "", nil, false
	}
	unit = strings.ToLower(strings.TrimSpace(parts[0]))
	spec := parts[1]
	for _, seg := range strings.Split(spec, ",") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		ab := strings.SplitN(seg, "-", 2)
		if len(ab) != 2 {
			return "", nil, false
		}
		aStr, bStr := strings.TrimSpace(ab[0]), strings.TrimSpace(ab[1])
		var start, end int64
		var err error
		if aStr == "" { // -N → 末尾 N+1 バイト
			n, err := strconv.ParseInt(bStr, 10, 64)
			if err != nil || n < 0 {
				return "", nil, false
			}
			if n >= total-1 { // 要求が全体を超える → 全体
				start, end = 0, total-1
			} else {
				start, end = total-(n+1), total-1
			}
		} else if bStr == "" { // A- → A〜最後
			start, err = strconv.ParseInt(aStr, 10, 64)
			if err != nil || start < 0 || start >= total {
				return "", nil, false
			}
			end = total - 1
		} else { // A-B
			start, err = strconv.ParseInt(aStr, 10, 64)
			if err != nil || start < 0 {
				return "", nil, false
			}
			end, err = strconv.ParseInt(bStr, 10, 64)
			if err != nil || start > end {
				return "", nil, false
			}
			if start >= total {
				return "", nil, false
			}
			if end >= total {
				end = total - 1
			}
		}
		ranges = append(ranges, byteRange{start, end})
	}
	return unit, ranges, true
}

func write416(w http.ResponseWriter, total int64) {
	w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", total))
	http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
}

func ifRangeMismatch(ifr string, currentETag string, currentLM time.Time) bool {
	// ETag か日付かの簡易判別（引用符が含まれていれば ETag とみなす）
	if strings.Contains(ifr, "\"") {
		return ifr != currentETag
	}
	if t, err := time.Parse(http.TimeFormat, ifr); err == nil {
		// 厳密比較（同一日時かどうか）。
		return !currentLM.Equal(t)
	}
	// 解釈不能な場合は不一致扱い
	return true
}

// 簡易アクセスログ
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
