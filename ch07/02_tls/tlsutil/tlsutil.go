package tlsutil

import (
	"crypto/tls"
	"fmt"
	"log"
)

// VersionName は TLS バージョンの定数値を人が読める文字列に変換します。
func VersionName(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS1.3"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS10:
		return "TLS1.0"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

// LogServerState は、サーバー側ハンドラで利用できる r.TLS を用いて、交渉結果の TLS 情報をログ出力します。
// 出力内容: バージョン、暗号スイート、ALPN、SNI、セッション再開フラグ、クライアント証明書の Subject（提示があれば）。
func LogServerState(state *tls.ConnectionState) {
	if state == nil {
		return
	}
	log.Printf("[TLS][server] version=%s cipher=%s alpn=%q sni=%q resumed=%v peerCerts=%d",
		VersionName(state.Version),
		tls.CipherSuiteName(state.CipherSuite),
		state.NegotiatedProtocol,
		state.ServerName,
		state.DidResume,
		len(state.PeerCertificates),
	)
	if len(state.PeerCertificates) > 0 {
		log.Printf("[TLS][server] client cert subject=%s", state.PeerCertificates[0].Subject.String())
	} else {
		log.Printf("[TLS][server] no client certificate presented")
	}
}

// LogClientState は、クライアント側で resp.TLS を用いて交渉結果の TLS 情報をログ出力します。
// 出力内容: バージョン、暗号スイート、ALPN、SNI、セッション再開フラグ、検証済みチェーン数、サーバー証明書の Subject。
func LogClientState(state *tls.ConnectionState) {
	if state == nil {
		return
	}
	log.Printf("[TLS][client] version=%s cipher=%s alpn=%q sni=%q resumed=%v verifiedChains=%d",
		VersionName(state.Version),
		tls.CipherSuiteName(state.CipherSuite),
		state.NegotiatedProtocol,
		state.ServerName,
		state.DidResume,
		len(state.VerifiedChains),
	)
	if len(state.PeerCertificates) > 0 {
		log.Printf("[TLS][client] server cert subject=%s", state.PeerCertificates[0].Subject.String())
	}
}
