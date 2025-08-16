# ch06/02_resume_range — Range/If-Range/複数範囲/再開ダウンロード（ブラウザで体験）

このディレクトリは、HTTP の Range による「中断と再開」「複数範囲」、If-Range、Accept-Ranges: none、圧縮応答（gzip）での Range の挙動を、1 本のサーバーとブラウザ UI だけで体験できる教材です。

- サーバーは 1 ファイル（server_resume_range.go）に集約。
- URL パスで機能を切り替えます。
- ブラウザの UI（/public/）から各種 Range を実行し、ステータスやレスポンスヘッダを確認できます。
- 並列 Range ダウンロードの簡易 UI（/public/parallel.html）も用意しています。

---

## 実行方法
1. サーバー起動
   - `go run ch06/02_resume_range/server_resume_range.go`
   - デフォルトで `:18062` で待ち受けます。
   - 起動時にコンソールへ「ブラウザで開く URL」を出力します。
2. ブラウザでトップへ
   - http://localhost:18062/

---

## エンドポイント一覧
- `GET /` … トップ（説明とリンク）
- 静的 UI: `/public/` とその配下
  - `/public/index.html` Range 実験 UI（各ページへのリンク付き）
  - `/public/single.html` 単一範囲を積み上げて 1 ファイルに結合
  - `/public/multipart.html` 複数範囲（multipart/byteranges）を積み上げて 1 ファイルに結合（教材用の簡易パース）
  - `/public/if-range.html` If-Range の体験（一致→206 / 不一致→200 フォールバック）
  - `/public/416.html` 416 Range Not Satisfiable の体験
  - `/public/none.html` Accept-Ranges: none の体験
  - `/public/gzip.html` gzip 応答に対する Range の体験
  - `/public/parallel.html` 並列ダウンロード UI
- `HEAD/GET /file`
  - Range（単一/複数）、If-Range（ETag/日付）を解釈
  - 200/206/416 を返します
  - `Accept-Ranges: bytes`、`ETag`、`Last-Modified` を付与
- `HEAD/GET /file_gzip`
  - 一度 gzip 圧縮した「圧縮後のバイト列」に対して Range を適用
  - `Content-Encoding: gzip` を返します
- `HEAD/GET /file_none`
  - `Accept-Ranges: none` として Range 指定を無視し、常に 200 で全体を返す
- `GET /flip_etag`
  - デモ用：ETag / Last-Modified を切り替えるトグル
  - If-Range の不一致を体験するために使用

---

## 使い方（ブラウザ）
- /public/index.html
  - 単一範囲（bytes=0-1023）
  - 末尾バイト（bytes=-999）
  - 複数範囲（bytes=500-999,7000-7999）
  - If-Range（一致→206 / 不一致→200）
  - 416 を発生させる例
  - Accept-Ranges: none 先に Range を送る例
  - gzip 応答での Range
- /public/parallel.html
  - HEAD で総サイズ/ETag を取得 → チャンク分割 → 複数の Range を並列 fetch
  - 完了後、結合して 1 つのファイルとして保存
  - If-Range（ETag）を付与して「再開」をシミュレート

---

## curl 例
- 単一範囲：
  - `curl -v -H "Range: bytes=0-1023" http://localhost:18062/file -o part.bin`
- 末尾 1000 バイト：
  - `curl -v -H "Range: bytes=-999" http://localhost:18062/file -o tail.bin`
- 複数範囲（multipart/byteranges）：
  - `curl -v -H "Range: bytes=500-999,7000-7999" http://localhost:18062/file`
- 416（不正範囲）：
  - `curl -v -H "Range: bytes=999999999-" http://localhost:18062/file`
- If-Range 成功 → 206：
  - `etag=$(curl -sI http://localhost:18062/file | awk -F": " '/^ETag/{print $2}' | tr -d '\r')`
  - `curl -v -H "If-Range: $etag" -H "Range: bytes=0-1023" http://localhost:18062/file -o ok.bin`
- If-Range 失敗 → 200（flip 後）：
  - `curl -s http://localhost:18062/flip_etag`
  - `curl -v -H "If-Range: $etag" -H "Range: bytes=0-1023" http://localhost:18062/file -o full.bin`
- Accept-Ranges: none：
  - `curl -v -H "Range: bytes=0-1023" http://localhost:18062/file_none -o ignored.bin`
- gzip 後の Range：
  - `curl -v -H "Range: bytes=0-1023" http://localhost:18062/file_gzip -o gz.part`

---

## aria2 例（停止→再開）
- `aria2c -x4 http://localhost:18062/file -o big.bin`
  - Ctrl+C で停止 → 同じコマンドで再開（Range 利用）

---

## 注意
- Content-Range は単数形（Content-Ranges ではありません）。
- Range は 0 始まり、end も含む範囲です（例: 0-0 は最初の 1 バイト）。
- `bytes=-N` は「末尾 N+1 バイト」の意味です。
- gzip 応答では「圧縮後のバイト列」に対して Range が適用されます。
- 並列ダウンロードはサーバーに負荷がかかる可能性があるため、実運用では十分に注意してください。
