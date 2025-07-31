# HTTP通信サンプルコード集

このディレクトリには、HTTP通信の様々な側面を学ぶためのサンプルコードが含まれています。各サブディレクトリは特定のHTTPリクエスト技術やHTTP機能に焦点を当てています。

## ディレクトリ構成

1. **01_simple_get** - 基本的なGETリクエストとクエリパラメータの使用例
   - `simple_get.go` - シンプルなGETリクエスト
   - `get_with_query.go` - クエリパラメータを含むGETリクエスト

2. **02_head** - HEADリクエストの例
   - `head_request.go` - HEADメソッドを使用したリクエスト

3. **03_post** - 様々なPOSTリクエストの例
   - `post_form.go` - フォームデータを使用したPOSTリクエスト
   - `post_multipart.go` - マルチパートフォームデータを使用したPOSTリクエスト
   - `post_multipart_mime.go` - MIMEタイプを指定したマルチパートPOSTリクエスト
   - `post_raw.go` - 生のデータを送信するPOSTリクエスト
   - `post_any_content_file.go` - ファイルコンテンツを送信するPOSTリクエスト
   - `post_any_content_text.go` - テキストコンテンツを送信するPOSTリクエスト

4. **04_cookie** - Cookieの使用例
   - `cookie_client.go` - Cookieを使用したHTTPリクエスト

5. **05_proxy** - プロキシを使用したリクエストの例
   - `proxy_request.go` - プロキシ経由でのHTTPリクエスト

6. **06_file** - ファイルスキームの使用例
   - `file_scheme.go` - fileスキームを使用したリクエスト

7. **07_custom_method** - カスタムHTTPメソッドの例
   - `delete_request.go` - DELETEメソッドを使用したリクエスト

8. **08_header** - HTTPヘッダーの操作例
   - `header_send.go` - カスタムヘッダーを含むリクエスト

9. **09_idn** - 国際化ドメイン名（IDN）の変換例
   - `idn_convert.go` - 国際化ドメイン名の変換

## リクエストファイルの実行方法

これらのファイルは、プロジェクトのルートディレクトリにある`server.go`に対してHTTPリクエストを送信するクライアントプログラムです。サーバーを実行してからクライアントを実行する必要があります。

### サーバーの起動

まず、サーバーを起動します：

```
go run server.go
```

このコマンドは、ポート18888でHTTPサーバーを起動し、受信したリクエストの詳細をコンソールに表示します。

### クライアントの実行

次に、別のターミナルで任意のリクエストファイルを実行します：

#### 基本的なGETリクエスト
```
go run ch04/01_simple_get/simple_get.go
```
シンプルなGETリクエストを送信し、レスポンスを表示します。

#### クエリパラメータを含むGETリクエスト
```
go run ch04/01_simple_get/get_with_query.go
```
URLエンコードされたクエリパラメータを含むGETリクエストを送信します。特殊文字のエンコード方法も表示します。

#### HEADリクエスト
```
go run ch04/02_head/head_request.go
```
HEADメソッドを使用してリクエストを送信し、ヘッダー情報のみを取得します。

#### フォームデータを使用したPOSTリクエスト
```
go run ch04/03_post/post_form.go
```
application/x-www-form-urlencodedフォーマットでデータを送信するPOSTリクエストの例です。

#### マルチパートフォームデータを使用したPOSTリクエスト
```
go run ch04/03_post/post_multipart.go
```
multipart/form-dataフォーマットでデータを送信するPOSTリクエストの例です。

#### MIMEタイプを指定したマルチパートPOSTリクエスト
```
go run ch04/03_post/post_multipart_mime.go
```
特定のMIMEタイプを指定したマルチパートフォームデータを送信します。

#### 生のデータを送信するPOSTリクエスト
```
go run ch04/03_post/post_raw.go
```
生のJSONデータを送信するPOSTリクエストの例です。

#### ファイルコンテンツを送信するPOSTリクエスト
```
go run ch04/03_post/post_any_content_file.go
```
ファイルの内容をPOSTリクエストで送信する例です。

#### テキストコンテンツを送信するPOSTリクエスト
```
go run ch04/03_post/post_any_content_text.go
```
テキストコンテンツをPOSTリクエストで送信する例です。

#### Cookieを使用したHTTPリクエスト
```
go run ch04/04_cookie/cookie_client.go
```
Cookieを設定・送信するHTTPリクエストの例です。

#### プロキシ経由でのHTTPリクエスト
```
go run ch04/05_proxy/proxy_request.go
```
プロキシサーバーを経由してHTTPリクエストを送信する例です。

#### fileスキームを使用したリクエスト
```
go run ch04/06_file/file_scheme.go
```
ローカルファイルシステム上のファイルにアクセスするためのfileスキームの使用例です。

#### DELETEメソッドを使用したリクエスト
```
go run ch04/07_custom_method/delete_request.go
```
DELETEメソッドを使用してリソースを削除するリクエストの例です。

#### カスタムヘッダーを含むリクエスト
```
go run ch04/08_header/header_send.go
```
カスタムHTTPヘッダーを設定して送信するリクエストの例です。

#### 国際化ドメイン名の変換
```
go run ch04/09_idn/idn_convert.go
```
国際化ドメイン名（IDN）をPunycode形式に変換する例です。

各ファイルは、HTTPリクエストの異なる側面を示しており、サーバーはリクエストの詳細をコンソールに表示します。これにより、HTTPリクエストの構造と動作を理解することができます。