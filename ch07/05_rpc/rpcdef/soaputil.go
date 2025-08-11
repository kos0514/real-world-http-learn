package rpcdef

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// SoapEnvNS は SOAP 1.1 の Envelope/Body で使用する名前空間です。
const SoapEnvNS = "http://schemas.xmlsoap.org/soap/envelope/"

// EnvelopeIn は受信用の SOAP Envelope です。Body の中身は innerxml で保持します。
type EnvelopeIn struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		Content []byte   `xml:",innerxml"`
	} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

// EnvelopeOut は送信用の SOAP Envelope です。Payload に任意の要素を格納します。
type EnvelopeOut struct {
	XMLName xml.Name     `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    EnvelopeBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type EnvelopeBody struct {
	Payload any `xml:",any"`
}

// MakeEnvelope はペイロードを包んだ送信用 Envelope を作成します。
func MakeEnvelope(payload any) EnvelopeOut {
	return EnvelopeOut{Body: EnvelopeBody{Payload: payload}}
}

// EncodeEnvelope は Envelope を XML にエンコードして writer に書き込みます。
// pretty が true の場合はインデントを付与し、withHeader が true の場合は XML 宣言を先頭に出力します。
func EncodeEnvelope(w io.Writer, payload any, pretty bool, withHeader bool) error {
	if withHeader {
		if _, err := w.Write([]byte(xml.Header)); err != nil {
			return err
		}
	}
	enc := xml.NewEncoder(w)
	if pretty {
		enc.Indent("", "  ")
	}
	if err := enc.Encode(MakeEnvelope(payload)); err != nil {
		return fmt.Errorf("encode envelope: %w", err)
	}
	return enc.Flush()
}

// MarshalEnvelope は Envelope をメモリ上に構築し、バイト列を返します。
func MarshalEnvelope(payload any, pretty bool, withHeader bool) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := EncodeEnvelope(buf, payload, pretty, withHeader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalEnvelope は受信した XML を EnvelopeIn に復号します。
func UnmarshalEnvelope(data []byte, out *EnvelopeIn) error {
	return xml.Unmarshal(data, out)
}

// BodyFirstStart は Body の inner XML から最初の開始タグを取り出し、
// そのタグを返すとともに、残りのデコードを続けるためのデコーダを返します。
func BodyFirstStart(content []byte) (*xml.Decoder, xml.StartElement, error) {
	dec := xml.NewDecoder(bytes.NewReader(content))
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, xml.StartElement{}, err
		}
		if st, ok := tok.(xml.StartElement); ok {
			return dec, st, nil
		}
	}
}
