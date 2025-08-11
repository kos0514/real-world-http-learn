package main

// client_soap.go
//
// 目的:
// - Go 標準ライブラリのみで SOAP サーバ(:8081/soap)へ 3 操作（保管庫/鑑定/全移動）を呼び出すクライアント。
// - リクエスト/レスポンス要素の定義は ch07/05_rpc/rpcdef で共通化し、重複を排除します。
//
// 実行:
//   1) 別ターミナルで server_soap.go を起動
//   2) 本ファイルを実行すると、3 操作を順に呼び出し、ログに結果を出力します。

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"real-world-http-learn/ch07/05_rpc/rpcdef"
)

const (
	endpoint = "http://localhost:8081/soap"
)

func main() {
	// 保管庫
	res1, err := callPutInArchive("ITEM-001")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[保管庫]", res1.Result)

	// 鑑定
	res2, err := callAppraise("ITEM-XYZ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[鑑定] Value=", res2.Value)

	// 全移動
	res3, err := callMoveAll("Shelf-A", "Shelf-B", true)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[全移動] Moved=", res3.Moved)
}

// callPutInArchive は「保管庫」操作を呼び出します。
func callPutInArchive(itemID string) (*rpcdef.PutInArchiveResponse, error) {
	req := rpcdef.PutInArchiveRequest{ItemID: itemID}
	var resp rpcdef.PutInArchiveResponse
	if err := doSOAPCall(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// callAppraise は「鑑定」操作を呼び出します。
func callAppraise(itemID string) (*rpcdef.AppraiseResponse, error) {
	req := rpcdef.AppraiseRequest{ItemID: itemID}
	var resp rpcdef.AppraiseResponse
	if err := doSOAPCall(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// callMoveAll は「全移動」操作を呼び出します。
func callMoveAll(from, to string, includeHold bool) (*rpcdef.MoveAllResponse, error) {
	req := rpcdef.MoveAllRequest{From: from, To: to, IncludeHold: includeHold}
	var resp rpcdef.MoveAllResponse
	if err := doSOAPCall(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// doSOAPCall は任意のリクエスト要素を SOAP Envelope で包んで送信し、
// Body 直下のレスポンス要素を out にデコードします。Fault の場合は error を返します。
func doSOAPCall(payload any, out any) error {
	// Envelope に包む
	breq, err := rpcdef.MarshalEnvelope(payload, false, false)
	if err != nil {
		return fmt.Errorf("encode envelope: %w", err)
	}
	// 送信するXMLを標準出力にダンプ（確認用）
	fmt.Println("---- Request XML ----")
	fmt.Println(string(breq))

	// HTTP POST で送信
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(breq))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// SOAPAction ヘッダは今回は未使用（サーバ側も参照しない）

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	// 受信したXMLを標準出力にダンプ（確認用）
	fmt.Println("---- Response XML ----")
	fmt.Println(string(b))

	// Envelope として復号
	var in rpcdef.EnvelopeIn
	if err := rpcdef.UnmarshalEnvelope(b, &in); err != nil {
		return fmt.Errorf("invalid SOAP envelope: %w", err)
	}

	// Body 直下の最初の要素を確認し、Fault ならエラー返却
	dec, st, err := rpcdef.BodyFirstStart(in.Body.Content)
	if err != nil {
		return fmt.Errorf("empty SOAP body: %w", err)
	}
	// Fault 判定（soapenv:Fault）
	if st.Name.Local == "Fault" && st.Name.Space == rpcdef.SoapEnvNS {
		// 最低限 faultstring を拾って返す
		type fault struct {
			FaultCode   string `xml:"faultcode"`
			FaultString string `xml:"faultstring"`
		}
		var f fault
		if err := dec.DecodeElement(&f, &st); err != nil {
			return errors.New("soap fault")
		}
		if f.FaultString != "" {
			return errors.New(f.FaultString)
		}
		return errors.New("soap fault")
	}
	// Fault でなければ期待のレスポンスとしてデコード
	if err := dec.DecodeElement(out, &st); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
