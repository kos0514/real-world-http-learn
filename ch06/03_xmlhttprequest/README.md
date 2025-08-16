# ch06/03_xmlhttprequest — XMLHttpRequest の基本/Comet/CORS（ブラウザで体験）

この教材では、XMLHttpRequest の基本（GET/POST/ヘッダ/フォーム/リダイレクト）から、ポーリング/ロングポーリング（Comet）、
CORS、Cookie と withCredentials の関係までを、1 本のサーバーと HTML で体験します。

- サーバーは 1 ファイル（server_xhr.go）に集約
- UI は /public/ 配下にページ分割

---

## 実行方法
1. サーバー起動
   - `go run ch06/03_xmlhttprequest/server_xhr.go`
   - 既定ポート: `:18063`
2. ブラウザでトップへ
   - http://localhost:18063/

---

## エンドポイント一覧
- `GET /json` JSON 応答 + HttpOnly Cookie（demo_session）
- `ANY /echo` 受けたメソッド/ヘッダ/ボディ/フォームを JSON で反射
- `POST /upload` multipart/form-data を受け取り、ファイル名/サイズなど要約
- `GET /redirect` → 302 Location: /json
- `GET /headers` リクエストヘッダの一覧
- `GET /poll` 短い JSON（現在時刻）。ポーリング用。
- `POST /comet/send` メッセージ送信
- `GET /comet/recv` ロングポーリング（最大 25 秒待機 / 1件受信 or 204）
- `GET /set_cookie` 非 HttpOnly Cookie をセット（document.cookie から見える）
- `GET /clear_cookie` 上記 Cookie を削除
- `GET /cors/json` CORS: `Access-Control-Allow-Origin: *`
- `GET /cors/with_credentials` CORS 資格情報許可（`Access-Control-Allow-Credentials: true`）+ Cookie セット
- `OPTIONS /cors/preflight` プリフライト応答

---

## ブラウザでの学び方
- `/public/01_basic_get.html` … XHR の最小例（onload, status, responseText, getAllResponseHeaders）
- `/public/02_post_json.html` … Content-Type: application/json で POST
- `/public/03_formdata_upload.html` … FormData による multipart/form-data 送信
- `/public/04_custom_headers.html` … setRequestHeader と Forbidden Header の違い
- `/public/05_redirect.html` … XHR によるリダイレクト追従（最終レスポンスを取得）
- `/public/06_polling.html` … 定期ポーリングのコスト/仕組み
- `/public/07_long_polling.html` … Comet ロングポーリング（サーバーからの擬似 push）
- `/public/08_cors.html` … `withCredentials` の挙動（同一オリジンでは差を感じにくいのでヘッダを観察）
- `/public/09_security_notes.html` … Same Origin / HttpOnly / withCredentials の注意

---

## curl 例
- 基本: `curl -v http://localhost:18063/json`
- POST JSON: `curl -v -H 'Content-Type: application/json' -d '{"a":1}' http://localhost:18063/echo`
- アップロード: `curl -v -F file=@/etc/hosts -F note=hello http://localhost:18063/upload`
- ヘッダ反射: `curl -v -H 'MyHeader: X' http://localhost:18063/headers`
- Comet 送信: `curl -v -X POST -d 'msg=hi' http://localhost:18063/comet/send`

---

## 注意・補足
- XHR は Forbidden Header（Accept-Encoding, Cookie, Host, Origin, Referer など）を自分で設定できません。
- CONNECT/TRACE/TRACK は open() 時にブラウザが拒否します。
- Same Origin Policy により、既定では同一オリジン以外へ送れません。CORS で許可を表明する必要があります。
- `withCredentials=true` のとき、サーバーは `Access-Control-Allow-Credentials: true` と具体的 `Access-Control-Allow-Origin` を返す必要があります（`*` は不可）。
- HttpOnly Cookie は document.cookie から参照できません（/json で demo_session を付与）。
