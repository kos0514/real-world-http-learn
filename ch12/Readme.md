# 第12章「JavaScriptによるHTTP」の要点まとめ

## JavaScriptでのHTTP通信の概要

### 通信手法の進化
- **従来の手法**: フォームsubmit、`location.href`、`<a>`タグによるページ遷移
- **Ajax時代**: XMLHttpRequestによる非同期通信（2005年頃〜）
- **現代**: Fetch API、Server-Sent Events、WebSocketなど多様化

---

## XMLHttpRequest（レガシーAPI）

### 基本的な使い方
```javascript
var xhr = new XMLHttpRequest();
xhr.open("GET", "/json", true);
xhr.onload = function () {
    if (xhr.status === 200) {
        console.log(JSON.parse(xhr.responseText));
    }
};
xhr.setRequestHeader("MyHeader", "HeaderValue");
xhr.send();
```

### XMLHttpRequest Level 2の改善点
- `onload`イベントの追加（従来の`onreadystatechange`より簡潔）
- より直感的なAPI設計
- しかしInternet Explorer対応の負担

### 現状
- Fetch APIまたはAxiosライブラリの利用が推奨
- XMLHttpRequestは後方互換性のために残存

---

## Server-Sent Events

### 特徴
- **一方向通信**: サーバーからクライアントへのプッシュ型
- **自動再接続**: 接続が切れても自動的に再接続
- **イベントベース**: カスタムイベントをサポート

### 実装例
```javascript
const evtSource = new EventSource("ssedemo.php");

// デフォルトメッセージ
evtSource.onmessage = (e) => {
    console.log("message: " + e.data);
};

// カスタムイベント
evtSource.addEventListener("ping", (e) => {
    const obj = JSON.parse(e.data);
    console.log("ping at " + obj.time);
}, false);
```

### データ形式
```
id: 10
event: ping
data: {"time": "2016-12-26T15:52:01+0000"}
```

### 制限事項
- `withCredentials`が必要な場合の認証
- Authorization headerの使用不可

---

## WebSocket

### 特徴
- **双方向リアルタイム通信**: クライアント↔サーバー
- **低レイテンシ**: TCPライクなAPI
- **HTTP/2、HTTP/3対応**: 255接続まで可能（Chrome）

### 基本的な使い方
```javascript
var socket = new WebSocket('ws://game.example.com:12010/updates');

socket.onopen = () => {
    setInterval(() => {
        if (socket.bufferedAmount === 0) {
            socket.send(getUpdateData());
        }
    }, 50);
};

socket.onmessage = (event) => {
    console.log(event.data);
};

socket.close();
```

### 送受信可能なデータ型
- 文字列（テキスト）
- Blob
- ArrayBuffer

### HTTP/2・HTTP/3でのWebSocket
- HTTP/1.1: `Upgrade`ヘッダーを使用
- HTTP/2: CONNECT擬似ヘッダー + `:protocol: websocket`
- HTTP/3: RFC 9220で標準化

---

## Fetch API（モダンなHTTP通信）

### 基本構文
```javascript
const response = await fetch("news.json", {
    method: 'GET',
    mode: 'cors',
    credentials: 'include',
    cache: 'default',
    headers: {
        'Content-Type': 'application/json'
    }
});

if (response.ok) {
    const json = await response.json();
    console.log(json);
}
```

### レスポンス処理メソッド
- `arrayBuffer()`: ArrayBuffer型（Typed Array用）
- `blob()`: Blob型（ファイル処理用）
- `formData()`: FormData型
- `json()`: JavaScriptオブジェクト
- `text()`: 文字列

---

## CORSモードの設定

### modeオプション
- **cors**: CORS対応（デフォルト、Fetch API）
- **same-origin**: 同一オリジンのみ
- **no-cors**: CORSなし（レスポンス読み取り不可）
- **navigate**: ページ遷移用
- **websocket**: WebSocket用

### Simple Cross-Origin Request条件
1. メソッド: GET、POST、HEAD
2. ヘッダー: Accept、Accept-Language、Content-Language、Content-Type、Range
3. Content-Type: `application/x-www-form-urlencoded`、`multipart/form-data`、`text/plain`

---

## Credentialsの設定

### credentialsオプション
- **omit**: 認証情報を送信しない
- **same-origin**: 同一オリジンのみ（デフォルト、Fetch API）
- **include**: 常に送信（XMLHttpRequestの`withCredential=true`相当）

### CORS時の制約
- `credentials: 'include'`時は`Access-Control-Allow-Origin: *`使用不可
- 401 Forbiddenエラーが発生する可能性

---

## データ送信の方法

### 1. application/x-www-form-urlencoded
```javascript
const params = new URLSearchParams();
params.set("name", "恵比寿東公園");
params.append("hasTako", "true");

const res = await fetch("/post", {
    method: "POST",
    body: params
});
```

### URLSearchParamsの便利メソッド
- `set()`: 値を設定（上書き）
- `append()`: 値を追加
- `has()`: キーの存在確認
- `get()`: 値の取得
- `getAll()`: 同名キーの全値取得
- `toString()`: URLエンコード文字列に変換

### 2. multipart/form-data
```javascript
const form = new FormData();
form.set("title", "The Art of Community");
form.set("author", "Jono Bacon");

// ファイル添付
const content = "Hello World";
const blob = new Blob([content], { type: "text/plain"});
form.set("attachment-file", blob, "test.txt");

const res = await fetch("/post", {
    method: "POST",
    body: form
});
```

### 3. JSON送信
```javascript
const res = await fetch("/post", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        "title": "The Art of Community",
        "author": "Jono Bacon"
    })
});
```

---

## キャッシュ制御

### cacheオプション
- **default**: ブラウザのデフォルト動作
- **no-store**: キャッシュしない
- **reload**: 常にサーバーから取得（ETagを送信しない）
- **no-cache**: 条件付きリクエスト（ETagで304チェック）
- **force-cache**: キャッシュ優先（Max-Age無視）
- **only-if-cached**: キャッシュのみ使用

---

## リダイレクト制御

### redirectオプション
- **follow**: 自動的に追従（デフォルト、最大20回）
- **manual**: 手動制御（`response.type`が`opaqueredirect`）
- **error**: リダイレクトをエラーとして扱う

---

## ストリーミング処理

### ReadableStreamの活用
```javascript
const response = await fetch(url);
const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
    const { done, value } = await reader.read();
    if (done) return;
    
    console.log(JSON.parse(decoder.decode(value)));
}
```

### ストリーミング送信
```javascript
const stream = new ReadableStream({
    start(controller) {
        const timer = setInterval(() => {
            controller.enqueue(JSON.stringify({timestamp: Date.now()}));
            if (count > 10) {
                controller.close();
                clearInterval(timer);
            }
        }, 1000);
    }
}).pipeThrough(new TextEncoderStream());

const res = await fetch(url, {
    method: "POST",
    body: stream,
    duplex: "half"
});
```

---

## ファイルダウンロード

### 1. 通常のダウンロード
```javascript
const anchor = document.createElement("a");
anchor.href = "https://tako.example.com";
anchor.download = "tako.json";
document.body.appendChild(anchor);
anchor.click();
document.body.removeChild(anchor);
```

### 2. 認証付きダウンロード
```javascript
const res = await fetch("https://tako.example.com", {
    headers: {
        Authorization: "Basic XXXXX"
    }
});

if (res.ok) {
    const anchor = document.createElement("a");
    anchor.href = URL.createObjectURL(await res.blob());
    anchor.download = "tako.json";
    document.body.appendChild(anchor);
    anchor.click();
    URL.revokeObjectURL(anchor.href);
    document.body.removeChild(anchor);
}
```

### 3. クライアントサイドでのExcel生成
```javascript
// SheetJSライブラリを使用
const wb = XLSX.utils.book_new();
const ws = XLSX.utils.aoa_to_sheet([
    ["三菱UFJ銀行", "0005"],
    ["三井住友銀行", "0009"]
]);
XLSX.utils.book_append_sheet(wb, ws, "Bank Codes");

const xlsx = XLSX.write(wb, { type: "array" });
const dataUri = URL.createObjectURL(new Blob([xlsx], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}));

const a = document.createElement("a");
a.href = dataUri;
a.download = "bankcode.xlsx";
a.click();
URL.revokeObjectURL(dataUri);
```

---

## その他のJavaScript HTTP技術

### 従来の方法（レガシー）
1. **location.href**: `location.href = "https://example.com";`
2. **form.submit()**: プログラマティックなフォーム送信
3. **window.open()**: 新規ウィンドウ・タブでのオープン

### マスク化（セキュリティ）
- `autocomplete`属性の活用
- `autocomplete="username"`: ユーザー名
- `autocomplete="current-password"`: 現在のパスワード
- `autocomplete="new-password"`: 新しいパスワード

---

## 環境の違い

### Node.js環境
- **Node.js 18以降**: `fetch()`がグローバルに利用可能
- **それ以前**: `node-fetch`などのライブラリが必要
- **URL**: Node.jsではファイルシステムパスも扱える

### クロスプラットフォーム対応
- **Winter CG**: Cloudflare Workers、Deno、Bun等での標準化
- **React Native**: 独自のFetch API実装

---

## まとめ

### 推奨される技術選択
- **一般的なHTTP通信**: Fetch API
- **リアルタイム一方向**: Server-Sent Events
- **リアルタイム双方向**: WebSocket
- **レガシーブラウザ対応**: XMLHttpRequest（または polyfill）

### Service Workerとの連携
- Fetch APIはService Workerで必須
- オフライン対応、キャッシュ戦略の実装に活用