package main

// server_soap.go
//
// 目的:
// - Go 標準ライブラリのみで SOAP(1.1 相当) の最小サーバを実装します。
// - ステータス区分「保管庫・鑑定・全移動」に対応する 3 操作を提供します。
// - リクエスト/レスポンスの要素定義は ch07/05_rpc/rpcdef に集約して重複を回避します。
//
// 動作:
// - エンドポイント: http://localhost:8081/soap
// - XML(Envelope/Body) を受け取り、Body 直下の要素名で操作をディスパッチします。
// - レスポンスは SOAP Envelope で包んで返却します。エラー時は簡易 Fault を返します。

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	"real-world-http-learn/ch07/05_rpc/rpcdef"
)

// 受信用の Envelope（Body の中身は innerxml で取り出す）
type envelopeIn struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		Content []byte   `xml:",innerxml"`
	} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

// 送信用の Envelope（Payload に任意の要素を詰める）
type envelopeOut struct {
	XMLName xml.Name     `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    envelopeBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type envelopeBody struct {
	Payload any `xml:",any"`
}

func main() {
	http.HandleFunc("/soap", soapHandler)
	log.Println("SOAP server listening on :8081/soap")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

// SOAP 受信処理: Body 直下の要素名で分岐し、対応する処理を行う。
func soapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method Not Allowed"))
		return
	}

	// リクエストボディを読み取り
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid body"))
		return
	}

	// SOAP Envelope としてパース
	var env envelopeIn
	if err := xml.Unmarshal(bodyBytes, &env); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid SOAP envelope"))
		return
	}

	// Body 直下の最初の開始タグを取得
	dec := xml.NewDecoder(bytes.NewReader(env.Body.Content))
	var se xml.StartElement
	for {
		tok, err := dec.Token()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("empty SOAP body"))
			return
		}
		if st, ok := tok.(xml.StartElement); ok {
			se = st
			break
		}
	}

	switch se.Name.Local {
	case "PutInArchive":
		var req rpcdef.PutInArchiveRequest
		if err := dec.DecodeElement(&req, &se); err != nil {
			soapFault(w, fmt.Sprintf("bad PutInArchive: %v", err))
			return
		}
		log.Printf("[保管庫] item=%s", req.ItemID)
		resp := rpcdef.PutInArchiveResponse{Result: "Archived: " + req.ItemID}
		writeSOAP(w, resp)
		return

	case "Appraise":
		var req rpcdef.AppraiseRequest
		if err := dec.DecodeElement(&req, &se); err != nil {
			soapFault(w, fmt.Sprintf("bad Appraise: %v", err))
			return
		}
		log.Printf("[鑑定] item=%s", req.ItemID)
		// ダミー鑑定: 文字数×100
		resp := rpcdef.AppraiseResponse{Value: len(req.ItemID) * 100}
		writeSOAP(w, resp)
		return

	case "MoveAll":
		var req rpcdef.MoveAllRequest
		if err := dec.DecodeElement(&req, &se); err != nil {
			soapFault(w, fmt.Sprintf("bad MoveAll: %v", err))
			return
		}
		log.Printf("[全移動] from=%s to=%s includeHold=%v", req.From, req.To, req.IncludeHold)
		// ダミー移動件数: from/to の文字数合計
		resp := rpcdef.MoveAllResponse{Moved: len(req.From) + len(req.To)}
		writeSOAP(w, resp)
		return

	default:
		soapFault(w, "unknown operation: "+se.Name.Local)
		return
	}
}

// 正常レスポンスを SOAP Envelope で返却
func writeSOAP(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	out := envelopeOut{Body: envelopeBody{Payload: payload}}
	w.WriteHeader(http.StatusOK)
	_ = xml.NewEncoder(w).Encode(out)
}

// SOAP Fault（簡易版）を返却
func soapFault(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	// SOAP Fault の最小要素（名前空間付き）
	type fault struct {
		XMLName     xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`
		FaultCode   string   `xml:"faultcode"`
		FaultString string   `xml:"faultstring"`
	}
	out := envelopeOut{Body: envelopeBody{Payload: fault{FaultCode: "SOAP-ENV:Client", FaultString: msg}}}
	// SOAP では HTTP 200 で Fault を包むケースも多い（今回はそれに倣う）
	w.WriteHeader(http.StatusOK)
	_ = xml.NewEncoder(w).Encode(out)
}
