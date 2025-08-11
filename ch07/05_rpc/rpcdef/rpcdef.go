// パッケージ rpcdef は、RPC で利用する共通のデータ型（インターフェイス定義）を
// まとめて提供します。クライアント/サーバの双方から同じ構造体を参照することで、
// 重複定義や記述ゆれを防ぎます。
package rpcdef

import "encoding/xml"

// ------------------------------
// JSON-RPC 用の共通型
// ------------------------------
// Args は JSON-RPC の Calculator.Multiply に渡す引数を定義する構造体です。
// クライアントとサーバの双方で同じ定義が必要になるため、このパッケージに配置しています。
// フィールドはエクスポートされている必要があります（net/rpc によるシリアライズ対象）。
// A と B を掛け算して結果を返す用途で使用します。
type Args struct {
	A, B int
}

// ------------------------------
// SOAP 用のインターフェイス定義
// ------------------------------
// ステータス区分: 保管庫(PutInArchive) / 鑑定(Appraise) / 全移動(MoveAll)
// 注意:
// - XMLタグにはサービスのターゲット名前空間 urn:StatusService を付与しています。
// - これらの構造体は SOAP の Body 直下に現れる要素を表現します。
// - クライアント/サーバ双方でこの型を使うことで、重複定義を排除します。

// PutInArchiveRequest は「保管庫」操作のリクエスト要素です。
// ItemID に対象アイテムの識別子を指定します。
type PutInArchiveRequest struct {
	XMLName xml.Name `xml:"urn:StatusService PutInArchive"`
	ItemID  string   `xml:"ItemID"`
}

// PutInArchiveResponse は「保管庫」操作のレスポンス要素です。
// Result に処理結果のメッセージを格納します。
type PutInArchiveResponse struct {
	XMLName xml.Name `xml:"urn:StatusService PutInArchiveResponse"`
	Result  string   `xml:"Result"`
}

// AppraiseRequest は「鑑定」操作のリクエスト要素です。
// ItemID に鑑定対象の識別子を指定します。
type AppraiseRequest struct {
	XMLName xml.Name `xml:"urn:StatusService Appraise"`
	ItemID  string   `xml:"ItemID"`
}

// AppraiseResponse は「鑑定」操作のレスポンス要素です。
// Value に鑑定額（本サンプルではダミー値）を格納します。
type AppraiseResponse struct {
	XMLName xml.Name `xml:"urn:StatusService AppraiseResponse"`
	Value   int      `xml:"Value"`
}

// MoveAllRequest は「全移動」操作のリクエスト要素です。
// From/To に移動元・移動先、IncludeHold に保留分を含めるかを指定します。
type MoveAllRequest struct {
	XMLName     xml.Name `xml:"urn:StatusService MoveAll"`
	From        string   `xml:"From"`
	To          string   `xml:"To"`
	IncludeHold bool     `xml:"IncludeHold"`
}

// MoveAllResponse は「全移動」操作のレスポンス要素です。
// Moved に移動件数（本サンプルではダミー値）を格納します。
type MoveAllResponse struct {
	XMLName xml.Name `xml:"urn:StatusService MoveAllResponse"`
	Moved   int      `xml:"Moved"`
}
