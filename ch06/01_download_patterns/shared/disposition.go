// パッケージ shared には、Content-Disposition ヘッダ生成の共通ヘルパをまとめます。
// RFC 6266（および RFC 5987）に沿った filename / filename* の扱いを簡単にするための関数群です。
package shared

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// rfc5987Encode は、Content-Disposition の filename* 用に UTF-8 文字列を
// パーセントエンコード（RFC 5987 形式）に変換します。
// 例: "レポート.xlsx" → "%E3%83%AC%E3%83%9D%E3%83%BC%E3%83%88.xlsx"
func rfc5987Encode(s string) string {
	return url.PathEscape(s)
}

// sanitizeASCIIFallback は、古いブラウザ向けに ASCII のみからなる安全な代替ファイル名を作ります。
// 非 ASCII 文字は下線に置換し、ダブルクォートやバックスラッシュも避けます。
// 完全に空になってしまう場合は "download" を採用します。
func sanitizeASCIIFallback(s string) string {
	re := regexp.MustCompile(`[^\x20-\x7E]`)
	safe := re.ReplaceAllString(s, "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, `\\`, "_")
	if strings.TrimSpace(safe) == "" {
		safe = "download"
	}
	return safe
}

// BuildAttachmentDisposition は、ダウンロード保存（attachment）を指示する Content-Disposition 値を生成します。
// 近代的なブラウザ向けに filename*（UTF-8 + パーセントエンコード）を付与し、
// 後方互換のために ASCII の filename も併記します。
// 例: attachment; filename*=utf-8”%E3%83%86%E3%82%B9%E3%83%88.txt; filename=test.txt
func BuildAttachmentDisposition(filenameUTF8 string, asciiFallback string) string {
	if asciiFallback == "" {
		asciiFallback = sanitizeASCIIFallback(filenameUTF8)
	}
	utf8Encoded := rfc5987Encode(filenameUTF8)
	return fmt.Sprintf(`attachment; filename*=utf-8''%s; filename=%s`, utf8Encoded, asciiFallback)
}

// BuildInlineDisposition は、ブラウザ内表示（inline）を指示しつつ、ファイル名のヒントを与える値を生成します。
// inline; filename*; filename の 3点セットを返します（未対応環境ではダウンロードになる場合もあります）。
func BuildInlineDisposition(filenameUTF8 string, asciiFallback string) string {
	if asciiFallback == "" {
		asciiFallback = sanitizeASCIIFallback(filenameUTF8)
	}
	utf8Encoded := rfc5987Encode(filenameUTF8)
	return fmt.Sprintf(`inline; filename*=utf-8''%s; filename=%s`, utf8Encoded, asciiFallback)
}
