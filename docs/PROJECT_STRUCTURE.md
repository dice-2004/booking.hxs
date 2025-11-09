# プロジェクト構造

このドキュメントでは、booking.hxs プロジェクトのディレクトリ構造とファイル配置について説明します。

## 📁 ディレクトリ構造

```
booking.hxs/
├── README.md                    # プロジェクトのメインドキュメント
├── LICENSE                      # MITライセンス
├── Makefile                     # ビルド・実行タスク
├── go.mod                       # Go モジュール定義
├── go.sum                       # 依存関係のチェックサム
├── main.go                      # エントリーポイント
├── .env                         # 環境変数（gitignore対象）
├── .gitignore                   # Git除外設定
├── reservations.json            # 予約データ（gitignore対象）
│
├── setup.sh                     # 🛠️ 自動セットアップスクリプト
├── switch_env.sh                # 🛠️ 環境切り替えスクリプト
├── manage_deps.sh               # 🛠️ 依存関係管理スクリプト
├── setup-systemd.sh             # 🛠️ systemd セットアップスクリプト
│
├── bin/                         # ビルド成果物
│   └── hxs_reservation_system   # 実行可能ファイル
│
├── commands/                    # コマンド処理
│   └── handlers.go              # Discordコマンドハンドラー
│
├── models/                      # データモデル
│   └── reservation.go           # 予約モデル定義
│
├── storage/                     # データストレージ
│   ├── storage.go               # ストレージ実装
│   └── storage_test.go          # ストレージテスト
│
├── logging/                     # ログ機能
│   └── logger.go                # ロガー実装
│
├── logs/                        # ログファイル
│   ├── command_stats.json       # コマンド統計
│   └── YYYY-MM-DD.log          # 日次ログ
│
├── docs/                        # 📚 ドキュメント
│   ├── PROJECT_STRUCTURE.md     # このファイル
│   ├── QUICKSTART.md            # クイックスタート
│   ├── DEVELOPMENT.md           # 開発ガイド
│   ├── PROJECT_SUMMARY.md       # プロジェクト概要
│   ├── CLEANUP.md               # クリーンアップ機能
│   ├── LOGGING.md               # ログ機能
│   └── SYSTEMD_SETUP.md         # systemdセットアップ
│
└── config/                      # ⚙️ 設定ファイル
    ├── .env.example             # 環境変数サンプル
    ├── .env.development         # 開発環境設定
    ├── .env.production          # 本番環境設定
    ├── .air.toml                # ホットリロード設定
    └── hxs-reservation-bot.service  # systemdサービスファイル
```

## 📂 ディレクトリの説明

### ルートディレクトリ

| ファイル/ディレクトリ | 説明 |
|---------------------|------|
| `README.md` | プロジェクトのメインドキュメント。導入、セットアップ、使い方を説明 |
| `LICENSE` | MITライセンス |
| `Makefile` | ビルド、実行、テストなどのタスク定義 |
| `go.mod` / `go.sum` | Goモジュール定義と依存関係 |
| `main.go` | アプリケーションのエントリーポイント |
| `.env` | 環境変数（DISCORD_TOKEN、GUILD_IDなど）※Git管理外 |
| `reservations.json` | 予約データの永続化ファイル ※Git管理外 |
| `setup.sh` | 🛠️ 自動セットアップスクリプト |
| `switch_env.sh` | 🛠️ 環境切り替えスクリプト |
| `manage_deps.sh` | 🛠️ 依存関係管理スクリプト |
| `setup-systemd.sh` | 🛠️ systemd セットアップスクリプト |

### `/bin/` - ビルド成果物

コンパイルされた実行可能ファイルが格納されます。

- `make build` でビルドされる
- Git管理外（.gitignoreに含まれる）

### `/commands/` - コマンド処理

Discord Botのコマンド処理ロジック。

- `handlers.go`: スラッシュコマンドのハンドラー実装
  - `/reserve`, `/cancel`, `/complete`, `/list`, `/my-reservations`

### `/models/` - データモデル

アプリケーションで使用するデータ構造の定義。

- `reservation.go`: 予約データの構造体と関連メソッド
  - `Reservation` 構造体
  - 予約ID生成
  - 時間重複チェック

### `/storage/` - データストレージ

予約データの永続化と管理。

- `storage.go`: ストレージ実装
  - JSONファイルへの読み書き
  - 予約のCRUD操作
  - 重複チェック
  - 自動クリーンアップ機能
- `storage_test.go`: ユニットテスト

### `/logging/` - ログ機能

コマンド実行のログとシステムログ。

- `logger.go`: ロガー実装
  - コマンド実行ログ
  - 統計情報収集
  - ログローテーション

### `/logs/` - ログファイル

実行時に生成されるログファイル。

- `YYYY-MM-DD.log`: 日次ログファイル
- `command_stats.json`: コマンド統計データ
- Git管理外

### `/docs/` - ドキュメント 📚

プロジェクトの各種ドキュメント。

| ファイル | 説明 |
|---------|------|
| `PROJECT_STRUCTURE.md` | プロジェクト構造（このファイル） |
| `QUICKSTART.md` | 最速でBotを起動する方法 |
| `DEVELOPMENT.md` | 開発環境のセットアップと開発フロー |
| `PROJECT_SUMMARY.md` | プロジェクトの全体像と設計思想 |
| `CLEANUP.md` | 自動クリーンアップ機能の詳細 |
| `LOGGING.md` | ログ機能とコマンド統計 |
| `SYSTEMD_SETUP.md` | Linux systemdでの自動起動設定 |

### スクリプト 🛠️

セットアップや運用で使用するシェルスクリプト（ルートディレクトリに配置）。

| スクリプト | 説明 | 使用例 |
|-----------|------|--------|
| `setup.sh` | 初回セットアップを自動化 | `./setup.sh` |
| `switch_env.sh` | 開発/本番環境を切り替え | `./switch_env.sh production` |
| `manage_deps.sh` | 依存関係の管理 | `./manage_deps.sh install` |
| `setup-systemd.sh` | systemdサービスのセットアップ | `./setup-systemd.sh` |

### `/config/` - 設定ファイル ⚙️

各種設定ファイルのサンプルとテンプレート。

| ファイル | 説明 |
|---------|------|
| `.env.example` | 環境変数のサンプル（コピーして使用） |
| `.env.development` | 開発環境用の設定 |
| `.env.production` | 本番環境用の設定 |
| `.air.toml` | ホットリロード（Air）の設定 |
| `hxs-reservation-bot.service` | systemdサービスファイル |

## 🔄 ワークフロー

### 開発開始時

1. `./scripts/setup.sh` で環境をセットアップ
2. `.env` を編集してトークンを設定
3. `make dev` で開発モードで起動

### 本番デプロイ

1. `./switch_env.sh production` で本番環境に切り替え
2. `make build` でビルド
3. `./setup-systemd.sh` でサービス登録
4. `sudo systemctl start hxs-reservation-bot` で起動

## 📝 ファイル命名規則

- **Go ソースファイル**: `snake_case.go`
- **ドキュメント**: `UPPER_SNAKE_CASE.md`
- **スクリプト**: `kebab-case.sh`
- **設定ファイル**: `.lowercase` または `kebab-case.extension`

## 🔐 Git管理外ファイル

以下のファイルは `.gitignore` で除外されています：

- `.env` - 機密情報を含む環境変数
- `bin/` - ビルド成果物
- `logs/` - ログファイル
- `reservations.json` - 予約データ
- `*.log` - その他のログファイル

## 🎯 設計原則

1. **明確な責任分離**: コマンド、モデル、ストレージ、ログを分離
2. **ドキュメント重視**: 各機能について詳細なドキュメントを提供
3. **スクリプト自動化**: セットアップや運用タスクを自動化
4. **設定の分離**: 環境ごとに設定を分離し、切り替え可能に
5. **テスト可能性**: ユニットテスト可能な構造
