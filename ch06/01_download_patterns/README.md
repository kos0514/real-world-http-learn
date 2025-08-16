# ch06/01_download_patterns — ブラウザのダウンロード挙動 入門（ブラウザだけで完結）

このディレクトリは、HTTP の Content-Disposition を中心に「ブラウザのダウンロード挙動」を体験的に学ぶための最小実装集です。
サーバーは 1 ファイル（server_download_patterns.go）に集約し、URL パスで機能を切り替えています。

重要: すべてブラウザだけで確認できます。専用クライアント（clients/）は削除しました。

## できること（学べる観点）
- Content-Type と Content-Disposition による「表示 or ダウンロード」の分岐
- attachment と inline の違い、filename と filename*（RFC 6266/5987）の扱い
- `<a download>` 属性によるダウンロード強制
- JavaScript で `<a>` 要素を動的生成してプログラム的にダウンロード
- 「ダウンロードありがとう」ページ → meta refresh で自動ダウンロード
- 擬似的な署名付き URL（Signed URL）発行 → 有効期限・署名検証 → ダウンロード（UIあり）
- curl の `-O` / `-J` の違い（URL 名 vs ヘッダの filename）

---

## 実行方法（サーバー）
1. サーバー起動
   - `go run ch06/01_download_patterns/server_download_patterns.go`
   - デフォルトで `:18061` で待ち受けます。
2. ブラウザでトップページへ
   - http://localhost:18061/
   - 各デモページへのリンクがあります。

## エンドポイント一覧（サーバー 1 本に集約）
- `GET /` トップページ：各デモへのリンクを表示。
- `GET /attachment?name=...`
  - Content-Disposition: attachment を返します。
  - 日本語ファイル名は `filename*`（UTF-8 + パーセントエンコード）を使用し、後方互換の `filename` も付与します。
  - 例: `attachment; filename*=utf-8''%E3%83%AC%E3%83%9D...xlsx; filename=report.xlsx`
- `GET /inline`
  - `Content-Type: image/png` + `Content-Disposition: inline` で 1x1 PNG をインライン表示。
  - ブラウザが対応していない形式であればダウンロードになります。
- `GET /public/a_download/`
  - `<a download>` 属性の静的デモ。`/assets/hello.txt` を使います。
- `GET /public/js_download/`
  - JS で `<a>` を生成して `.click()` し、`/api/download` のレスポンスをダウンロード。
- `GET /public/thanks/`
  - 「ダウンロードありがとうございます」ページ。`<meta http-equiv="refresh" content="0;URL=/download_file">` により自動的に `/download_file` を開きます。
- `GET /public/signed_url/`
  - ブラウザからボタンひとつで「擬似署名 URL 発行 → ダウンロード」を体験できます。
- `GET /download_file`
  - `Content-Disposition: attachment` を付与し、PDF として保存させます（中身はデモ用の最小バイト列）。
- `GET /api/download?name=...`
  - JS から叩くダウンロード用 API。`attachment` で HTML を返します。
- `GET /api/sign?name=...`
  - 擬似署名 URL を JSON で返します。ペイロード例: `{ "url": "/api/file?token=...&exp=...&name=..." }`
- `GET /api/file?token=...&exp=...&name=...`
  - 署名と有効期限を簡易検証し、OK なら `attachment` でプレーンテキストを返します。

静的ファイル:
- `/assets/hello.txt` … `<a download>` デモ用

---

## curl での確認（任意）
- URL 名で保存（`-O`）:
  - `curl -O http://localhost:18061/assets/hello.txt`
- ヘッダの `filename`/`filename*` で保存（`-J` と組み合わせ）:
  - `curl -OJ "http://localhost:18061/attachment?name=$(python - <<'PY'\nprint('レポート 2025-03.xlsx')\nPY)"`

注意:
- curl は URL エンコード名の自動デコードをしない場合があります（`%20` が残る等）。
- 同名ファイルがある場合、ブラウザは自動で連番を振りますが、curl はエラーになることがあります。

---

## 仕組みの要点（初心者向け）
- ブラウザが「表示するかダウンロードするか」を決める主因は MIME（`Content-Type`）です。
  - 例: `image/png` はインライン表示。`application/pdf` は多くのブラウザでインライン表示可能。
- これに対して、`Content-Disposition: attachment` を返すと「表示ではなく保存対象」と認識されます。
- 日本語ファイル名は `filename*`（RFC 5987）で UTF-8 + パーセントエンコードを使い、後方互換目的で ASCII の `filename` も併記します。
  - 多くのモダンブラウザは `filename*` を優先します。
- `<a download>` は「リンク先をダウンロードさせたい」時に便利。ただしクロスオリジン先では無効になることが多いです（同一オリジン推奨）。
- JavaScript でダウンロードする場合、`<a>` を動的生成して `click()` すれば、ページ遷移せずにダウンロードを開始できます。
- 「ダウンロードありがとう」ページは、meta refresh でダウンロード URL に遷移させることで、ページ自体は表示を保ったままダウンロードを始められます。

---

## セキュリティと実運用の注意
- ユーザー入力のファイル名をそのまま使わない（パストラバーサル等に注意）。
- 署名 URL（本物の運用では）
  - HTTPS 必須
  - 強固な署名鍵・短い有効期限・最小権限
  - 必要に応じてワンタイム化・IP 制限など
- 正しい `Content-Type` を返す（誤ると意図しない表示や関連付けになる）。

---

## 実装メモ
- サーバーは `server_download_patterns.go` に集約し、`http.ServeMux` でパス分岐しています。
- `shared/disposition.go` に `Content-Disposition` の生成ヘルパを用意し、`attachment`/`inline` の両方に対応。
- 1x1 PNG はコード内にバイト列で同梱（`/inline` で使用）。

---

## トラブルシューティング
- 18061 ポートが衝突する → 別のポートに変更して実行してください。
- ブラウザでダウンロードが始まらない
  - キャッシュをクリア／別ブラウザで再確認
  - `Content-Disposition` が正しく設定されているか確認
- 文字化けしたファイル名になる
  - ブラウザの実装差やバージョンによって挙動が異なる場合があります。`filename*` と ASCII `filename` の両方を付与して回避を試みてください。

---

以上です。トップページ（/）のリンクから順に試すと、各機能の違いを直感的に理解できます。