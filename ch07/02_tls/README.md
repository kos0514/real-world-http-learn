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


## サーバーの実行

サーバーの証明書と鍵を配置したうえで、以下の手順で起動します（カレントディレクトリを ch07/02_tls にして実行）。

```
cd ch07/02_tls
go run ./server_tls.go
```
