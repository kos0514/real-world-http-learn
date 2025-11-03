## 第11章「RESTful API」の詳細要点

### RESTの歴史と背景
- **2005年Google Maps以降**: Ajax技術とAPIの普及が本格化
- **Roy Fieldingの貢献**: HTTP/1.1の主要設計者でもあり、2000年に博士論文でREST概念を提唱
- **2010年代の状況**: API全体の74%がRESTfulアプローチを採用（2010年調査）
- **近年の変化**: DropboxがREST APIからRPC APIへの移行を発表するなど、必ずしもRESTが絶対ではない

### Richardson成熟度モデル（RESTのレベル）
- **Level 0**: HTTPを転送手段として使用、POSTのみでRPC的
- **Level 1**: リソースの概念導入、複数のURLを使用
- **Level 2**: HTTPメソッド（GET、POST、PUT、DELETE）を適切に使用
- **Level 3**: HATEOAS（Hypermedia as the Engine of Application State）の実装

### APIの対象による分類
- **LSUDs（Large Set of Unknown Developers）**:
    - 大規模で不特定多数の開発者向けAPI
    - 一般公開API、パートナーAPI
    - RESTfulアプローチが有効
    - 例：GitHub API、Google Maps API、Twitter API

- **SSKDs（Small Set of Known Developers）**:
    - 小規模で既知の開発者向けAPI
    - 社内システム、マイクロサービス間通信
    - RPC的アプローチも選択肢
    - 例：社内の決済サービス、ユーザー管理システム

### REST vs RPC の比較
- **REST**: リソース指向、HTTPメソッドでCRUD操作、URLでリソース表現
- **RPC**: 機能指向、POST中心、URL1つで複数機能
- **実用性**: LSUDs向けAPIではRESTが有効、SSKDs向けならRPCも実用的
- **パフォーマンス**: RPC（Apache Thrift、gRPC）の方が高速な場合がある

### HTTPメソッドの詳細設計
- **安全性**: GET、HEAD、OPTIONSはサーバー状態を変更しない
- **冪等性**: GET、HEAD、PUT、DELETEは何度実行しても同じ結果
- **POST vs PUT**: POSTは作成（非冪等）、PUTは作成・更新（冪等）
- **PATCH**: RFC 5789で定義、部分更新専用
- **実装上の注意**: HTML formはGETとPOSTのみサポート

### URL設計のベストプラクティス
- **リソース階層**: `/users/5/orders/10/products`のような論理的構造
- **単数形 vs 複数形**: コレクションは複数形（/users）、個別リソースはID指定
- **特殊パス**:
    - `/me`: 現在認証されているユーザー
    - `/self`: 自分自身のリソース
    - `/user`: 単一ユーザーコンテキスト
- **バッチ処理**: multipart/mixedを使った複数リクエストの一括処理

### 非同期処理とサーガパターン
- **サーガパターンの実装**:
    - データベースにステータスカラムを追加
    - `pending` → `processing` → `completed` または `failed`
    - 各ステータス変更をイベント駆動で処理
- **実用例**:
  ```sql
  CREATE TABLE orders (
    id INT PRIMARY KEY,
    status ENUM('pending', 'processing', 'completed', 'failed'),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
  );
  ```
- **メリット**: 長時間処理の進捗管理、障害時の復旧容易性

### HTTPステータスコードの使い分け
- **成功系**: 200 OK（取得）、201 Created（作成）、204 No Content（削除）
- **クライアントエラー**: 400（リクエスト不正）、401（認証必要）、403（認可エラー）、404（リソース未存在）、409（競合状態）
- **レート制限**: 429 Too Many Requests
- **設計指針**: 404 vs 403の使い分け（情報漏洩防止）

### HATEOAS（Level 3 REST）
- **概念**: APIレスポンス内に次のアクション用のリンクを含める
- **実装例**: XMLの`<link>`要素やJSONでのURL埋め込み
- **銀行API例**:
```xml
<account>
  <account_number>12345</account_number>
  <balance currency="usd">100.00</balance>
  <link rel="deposit" href="/account/12345/deposit" />
  <link rel="withdraw" href="/account/12345/withdraw" />
  <link rel="close" href="/account/12345/close" />
</account>
```
- **実用性**: 理論的には美しいが、実装・運用コストが高いため採用例は少ない

### 認証・認可の実装
- **Basic認証**: 簡単だがHTTPS必須
- **OAuth 2.0**: Authorization CodeフロウでのPKCE対応
- **JWT**: ステートレスだが適切な検証が必要
- **APIキー**: シンプルだが管理が課題

### Go言語での実装例
- **Context活用**: Go 1.7以降のcontext.WithTimeoutでタイムアウト制御
- **OAuth2実装**: golang.org/x/oauth2パッケージの使用
- **レート制限**: golang.org/x/time/rateによるトークンバケット実装
- **HTTPクライアント**: http.Clientのカスタマイズとタイムアウト設定

### 実際のAPI事例分析
- **PAY.JP**: 決済APIのRESTful設計
    - HTTPS必須、Basic認証使用
    - ページネーション（limit/offset）
    - 顧客情報取得：`GET /v1/customers/:id`

- **GitHub API**:
    - OAuth2認証フロー
    - レート制限（認証ありで5000リクエスト/時間）
    - WebhookとEvent APIの組み合わせ

### 新しい規格・トレンド
- **JSON Patch**: RFC 6902による部分更新の標準化
- **Sunset HTTP Header**: APIの廃止予告
- **GraphQL**: Facebookが開発したクエリ言語、RESTの代替案の一つ
- **gRPC**: GoogleのHTTP/2ベースRPCフレームワーク

### 運用上の考慮点
- **バージョニング**: URLパス、ヘッダー、Accept/Content-Typeでの管理方法
- **エラーハンドリング**: 一貫したエラーレスポンス形式
- **ドキュメンテーション**: OpenAPI（旧Swagger）による自動生成
- **モニタリング**: 分散トレーシング、メトリクス収集
- **セキュリティ**: CORS設定、CSP、レート制限の実装