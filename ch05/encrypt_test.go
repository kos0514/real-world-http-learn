// Package main は異なる暗号化アルゴリズムのパフォーマンスを比較するためのベンチマークテストを含みます。
// このファイルは特に RSA と AES の暗号化/復号化操作のベンチマークに焦点を当てています。
package main

import (
	"crypto/aes"    // 高度暗号化標準（AES）の実装
	"crypto/cipher" // ブロック暗号のインターフェース
	"crypto/md5"    // MD5ハッシュアルゴリズム（RSAパディングに使用）
	"crypto/rand"   // 暗号学的に安全な乱数生成器
	"crypto/rsa"    // RSA暗号化の実装
	"io"            // 入出力操作用
	"testing"       // Goテストフレームワーク
)

// prepareRSA はRSA暗号化/復号化ベンチマーク用のテストデータを準備します。
// 戻り値:
// - sourceData: 暗号化される乱数データ（128バイト）
// - label: OAEPパディング用のオプションコンテキストラベル（この場合は空）
// - privateKey: 新しく生成された2048ビットのRSA秘密鍵
func prepareRSA() (sourceData, label []byte, privateKey *rsa.PrivateKey) {
	// テストデータ用に128バイトのバッファを作成
	sourceData = make([]byte, 128)
	// OAEPパディング用の空のラベル
	label = []byte("")
	// バッファを暗号学的に安全な乱数データで満たす
	io.ReadFull(rand.Reader, sourceData)
	// 新しい2048ビットのRSA鍵ペアを生成
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	return
}

// BenchmarkRSAEncryption はRSA暗号化のパフォーマンスを測定します。
// RSAは公開鍵を使用して暗号化を行う非対称暗号化アルゴリズムです。
func BenchmarkRSAEncryption(b *testing.B) {
	// テストデータと鍵を準備
	sourceData, label, privateKey := prepareRSA()
	// 秘密鍵から公開鍵を抽出
	publicKey := &privateKey.PublicKey
	// OAEPパディング用のMD5ハッシュインスタンスを作成
	md5hash := md5.New()
	// セットアップ時間をベンチマークから除外するためにタイマーをリセット
	b.ResetTimer()
	// 暗号化操作をb.N回実行（ベンチマークフレームワークによって制御）
	for i := 0; i < b.N; i++ {
		// RSA-OAEP（Optimal Asymmetric Encryption Padding）を使用してデータを暗号化
		rsa.EncryptOAEP(md5hash, rand.Reader, publicKey, sourceData, label)
	}
}

// BenchmarkRSADecryption はRSA復号化のパフォーマンスを測定します。
// RSA復号化は、公開鍵で暗号化されたデータを復号化するために秘密鍵を使用します。
func BenchmarkRSADecryption(b *testing.B) {
	// テストデータと鍵を準備
	sourceData, label, privateKey := prepareRSA()
	// 秘密鍵から公開鍵を抽出
	publicKey := &privateKey.PublicKey
	// OAEPパディング用のMD5ハッシュインスタンスを作成
	md5hash := md5.New()
	// 復号化テスト用の暗号文を作成するためにテストデータを一度暗号化
	encrypted, _ := rsa.EncryptOAEP(md5hash, rand.Reader, publicKey, sourceData, label)
	// セットアップ時間をベンチマークから除外するためにタイマーをリセット
	b.ResetTimer()
	// 復号化操作をb.N回実行
	for i := 0; i < b.N; i++ {
		// RSA-OAEPを使用して暗号化されたデータを復号化
		rsa.DecryptOAEP(md5hash, rand.Reader, privateKey, encrypted, label)
	}
}

// prepareAES はAES暗号化/復号化ベンチマーク用のテストデータを準備します。
// 戻り値:
// - sourceData: 暗号化される乱数データ（128バイト）
// - nonce: 各暗号化操作に対する一意の値（12バイト）
// - gcm: AES用のGalois/Counter Modeの暗号インスタンス
func prepareAES() (sourceData, nonce []byte, gcm cipher.AEAD) {
	// テストデータ用に128バイトのバッファを作成
	sourceData = make([]byte, 128)
	// バッファを暗号学的に安全な乱数データで満たす
	io.ReadFull(rand.Reader, sourceData)
	// AES-256用の256ビット（32バイト）の鍵を作成
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)
	// 12バイトのノンス（一度だけ使用する数）を作成
	nonce = make([]byte, 12)
	io.ReadFull(rand.Reader, nonce)
	// 鍵を使用して新しいAES暗号ブロックを作成
	block, _ := aes.NewCipher(key)
	// 新しいGCM（Galois/Counter Mode）インスタンスを作成
	// GCMは暗号化と認証の両方を提供
	gcm, _ = cipher.NewGCM(block)
	return
}

// BenchmarkAESEncryption はGCMモードでのAES暗号化のパフォーマンスを測定します。
// AESは暗号化と復号化の両方に同じ鍵を使用する対称暗号化アルゴリズムです。
func BenchmarkAESEncryption(b *testing.B) {
	// テストデータ、ノンス、GCM暗号を準備
	sourceData, nonce, gcm := prepareAES()
	// セットアップ時間をベンチマークから除外するためにタイマーをリセット
	b.ResetTimer()
	// 暗号化操作をb.N回実行
	for i := 0; i < b.N; i++ {
		// GCMモードを使用してデータを暗号化および認証
		// パラメータ:
		// - nil: 宛先バッファ（nilは新しいバッファを作成することを意味する）
		// - nonce: この暗号化のための一意の値
		// - sourceData: 暗号化するデータ
		// - nil: 追加の認証データ（この場合はなし）
		gcm.Seal(nil, nonce, sourceData, nil)
	}
}

// BenchmarkAESDecryption はGCMモードでのAES復号化のパフォーマンスを測定します。
// GCMモードは復号化中に暗号化されたデータの真正性も検証します。
func BenchmarkAESDecryption(b *testing.B) {
	// テストデータ、ノンス、GCM暗号を準備
	sourceData, nonce, gcm := prepareAES()
	// 復号化テスト用の暗号文を作成するためにテストデータを一度暗号化
	encrypted := gcm.Seal(nil, nonce, sourceData, nil)

	// セットアップ時間をベンチマークから除外するためにタイマーをリセット
	b.ResetTimer()
	// 復号化操作をb.N回実行
	for i := 0; i < b.N; i++ {
		// GCMモードを使用して暗号化されたデータを復号化および検証
		// パラメータ:
		// - nil: 宛先バッファ（nilは新しいバッファを作成することを意味する）
		// - nonce: 暗号化に使用したのと同じノンス
		// - encrypted: 復号化する暗号化データ
		// - nil: 追加の認証データ（この場合はなし）
		gcm.Open(nil, nonce, encrypted, nil)
	}
}
