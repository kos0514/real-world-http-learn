# ch07/02_tls — TLS 学習用ディレクトリ

このディレクトリは、自己署名 CA と localhost サーバー証明書で TLS を学習するためのサンプルです。

## ディレクトリ構成

```
ch07/02_tls/
  README.md
  server_tls.go
  conf/                      # OpenSSL 設定
    openssl.cnf
  ca/
    private/                 # CA 秘密鍵 (コミット対象外推奨)
      ca.key
    certs/                   # CA の公開証明書を格納
      ca.crt
    csr/                     # CA の証明書署名要求（CSR）を格納
      ca.csr
  server/
    private/                 # サーバー秘密鍵 (コミット対象外推奨)
      server.key
    certs/                   # サーバーの公開証明書を格納
      server.crt
    csr/                     # サーバーの証明書署名要求（CSR）を格納
      server.csr
  client/
    private/                 # クライアント秘密鍵 (コミット対象外推奨)
      client.key
    certs/                   # クライアントの公開証明書を格納
      client.crt
    csr/                     # クライアントの証明書署名要求（CSR）を格納
      client.csr
```

注意:
- ここではリポジトリの現状レイアウトを示しています。
- 以下の「コマンド」は、ご提供のオプションはそのままに、パスのみ本レイアウトに合わせて修正しています。

## 証明書生成コマンド

CA 用:

```
# 256ビットの楕円曲線デジタル署名アルゴリズム (ECDSA) の認証局秘密鍵を作成
$ openssl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:prime256v1 -out ca/private/ca.key

# 証明書署名要求(CSR)を作成
$ openssl req -new -sha256 -key ca/private/ca.key -out ca/csr/ca.csr -config conf/openssl.cnf

# 証明書を自分の秘密鍵で署名して作成
$ openssl x509 -in ca/csr/ca.csr -days 365 -req -signkey ca/private/ca.key -sha256 -out ca/certs/ca.crt -extfile ./conf/openssl.cnf -extensions CA
```

サーバー用（Common Nameには localhostを指定）:

```
# 256ビットの楕円曲線デジタル署名アルゴリズム (ECDSA) のサーバー秘密鍵を作成
$ openssl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:prime256v1 -out server/private/server.key

# 証明書署名要求(CSR)を作成
$ openssl req -new -nodes -sha256 -key server/private/server.key -out server/csr/server.csr -config conf/openssl.cnf

# 証明書を自分の秘密鍵で署名して作成
$ openssl x509 -req -days 365 -in server/csr/server.csr -sha256 -out server/certs/server.crt -CA ca/certs/ca.crt -CAkey ca/private/ca.key -CAcreateserial -extfile ./conf/openssl.cnf -extensions Server
```

クライアント用（クライアント証明書）:

```
# 256ビットの楕円曲線デジタル署名アルゴリズム (ECDSA) のクライアント用秘密鍵を作成
$ openssl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:prime256v1 -out client/private/client.key

# 証明書署名要求(CSR)を作成
$ openssl req -new -nodes -sha256 -key client/private/client.key -out client/csr/client.csr -config conf/openssl.cnf

# 証明書を自分の秘密鍵で署名して作成
$ openssl x509 -req -days 365 -in client/csr/client.csr -sha256 -out client/certs/client.crt -CA ca/certs/ca.crt -CAkey ca/private/ca.key -CAcreateserial -extfile ./conf/openssl.cnf -extensions Client
```

## サーバーの実行

サーバーの証明書と鍵を配置したうえで、以下の手順で起動します（カレントディレクトリを ch07/02_tls にして実行）。

```
cd ch07/02_tls
go run ./server_tls.go
```

## TLS 可視化ログ

- サーバー側（server_tls.go）
  - 各リクエストで、交渉された TLS の情報をログ出力します。
  - 出力例のキー: version, cipher, alpn, sni, resumed, peerCerts, client cert subject（あれば）。
- クライアント側（client_tls_with_cert.go）
  - 応答受信後に、交渉結果の TLS 情報をログ出力します。
  - 出力例のキー: version, cipher, alpn, sni, resumed, verifiedChains, server cert subject。

実行例:

```
# サーバー
cd ch07/02_tls
go run ./server_tls.go

# クライアント（別シェルで）
cd ch07/02_tls
go run ./client_tls_with_cert.go
```

これらのログにより、TLS バージョンや暗号スイート、SNI、セッション再開の有無などが簡単に可視化できます。