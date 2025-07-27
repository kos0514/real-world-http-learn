// idn_convert.go - 国際化ドメイン名（IDN）の変換例
// IDNAを使用して非ASCII文字（日本語など）をASCII文字に変換する方法を示します
// 国際化ドメイン名は、Punycode形式でエンコードされ、"xn--"プレフィックスで始まります
package main

import (
	"fmt"
	"golang.org/x/net/idna"
)

func main() {
	// 変換元の日本語文字列（例: アニメ「爆竜戦隊アバレンジャー」のタイトル）
	src := "爆竜戦隊アバレンジャー"

	// idna.ToASCII関数を使用して日本語文字列をPunycodeに変換
	// この関数は国際化ドメイン名をASCII互換エンコーディング（ACE）に変換します
	ascii, err := idna.ToASCII(src)
	if err != nil {
		// 変換中にエラーが発生した場合はパニック
		// 一般的なエラー: 無効な文字、長さ制限超過など
		panic(err)
	}

	// 元の文字列と変換後のASCII文字列を表示
	// 出力例: "爆竜戦隊アバレンジャー -> xn--n8jvkxa8fv5evkbn1rh87atj9b"
	fmt.Printf("%s -> %s\n", src, ascii)
}
