# ch07/02_tls — TLS 学習用ディレクトリ

このディレクトリは、自己署名 CA と localhost サーバー証明書で TLS を学習するためのサンプルです。

## ディレクトリ構成

```
ch07/02_tls/
  README.md
  server_tls.go
  conf/
    openssl.cnf              # OpenSSL 設定
  ca/
    private/                 # CA 秘密鍵 (コミット対象外推奨)
      ca.key
    certs/
      ca.crt
    csr/
      ca.csr
  server/
    private/                 # サーバー秘密鍵 (コミット対象外推奨)
      server.key
    certs/
      server.crt
    csr/
      server.csr
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

サーバー用（Common Nameには localhost　を指定）:

```
# 256ビットの楕円曲線デジタル署名アルゴリズム (ECDSA) のサーバー秘密鍵を作成
$ openssl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:prime256v1 -out server/private/server.key

# 証明書署名要求(CSR)を作成
$ openssl req -new -nodes -sha256 -key server/private/server.key -out server/csr/server.csr -config conf/openssl.cnf

# 証明書を自分の秘密鍵で署名して作成
$ openssl x509 -req -days 365 -in server/csr/server.csr -sha256 -out server/certs/server.crt -CA ca/certs/ca.crt -CAkey ca/private/ca.key -CAcreateserial -extfile ./conf/openssl.cnf -extensions Server
```
