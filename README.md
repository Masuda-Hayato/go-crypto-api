# 📈 仮想通貨リアルタイム監視・異常検知LINE通知システム

外部APIから暗号資産の価格を定期取得してデータベースに永続化しつつ、異常な価格変動を検知して管理者のLINEに即時通知するバックエンドシステムです。収集したデータはWebダッシュボードおよびJSON APIとしても配信されます。

## 🛠 技術スタック
* **言語:** Go (Golang)
* **データベース:** SQLite3
* **外部API:** CoinGecko API, LINE Messaging API
* **主要パッケージ:**
  * 標準ライブラリ: `net/http`, `database/sql`, `encoding/json`, `html/template`
  * 外部ライブラリ: `[github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)` (DBドライバ), `[github.com/joho/godotenv](https://github.com/joho/godotenv)` (環境変数管理)

## ✨ 主な機能とアーキテクチャ

1. **自律型データ収集バッチ (`main.go`)**
   * `time.Ticker` を用いた1分ごとの非同期定期実行フェッチ
   * JSONデータの構造体へのマッピングと、SQLiteへの安全なデータ永続化
2. **異常検知とLINEへのアラート通知**
   * ポインタ (`*float64`) を用いて前回の価格状態をメモリ上に保持し、比較処理を高速化
   * 1分間で設定した閾値以上の変動があった場合、LINE Messaging APIを経由して管理者のスマホへ警告メッセージを自動送信
3. **実務標準のセキュリティ対策**
   * APIトークンなどの機密情報は `.env` ファイルに分離し、Git管理下から除外（`.gitignore`）することで漏洩を防止
4. **Webサーバー＆REST API (`server.go`)**
   * Goの標準ライブラリのみで構築した軽量Webサーバー
   * `html/template` を用いた直近30件の価格推移ダッシュボード提供
   * 外部から利用可能なJSON APIエンドポイント（`/api/prices`）の提供

## 🚀 環境構築と実行方法

本システムは、機密情報を保護するために環境変数ファイルを使用しています。実行前に以下のセットアップを行ってください。

### 1. セットアップ
リポジトリをクローン後、プロジェクトのルートディレクトリ（`main.go`と同じ階層）に `.env` ファイルを作成し、ご自身のLINE APIの情報を記述してください。

```text
LINE_TOKEN=あなたのLINEチャネルアクセストークン
LINE_USER_ID=あなたのユーザーID
```

### 2. システムの起動
本システムは、2つのプロセスを独立して稼働させる構成になっています。ターミナルを2つ開いて実行してください。

**データ収集・監視バッチの起動（ターミナル1）**
```bash
go run main.go
```

**Webサーバーの起動（ターミナル2）**
```bash
go run server.go
```

### 3. エンドポイント
* ダッシュボード (HTML): `http://localhost:8080`
* 配信 API (JSON): `http://localhost:8080/api/prices`
