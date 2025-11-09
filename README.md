# booking.hxs（Discord Bot）

Go言語で作成されたDiscord Bot用の面接予約システムです。スラッシュコマンドを使用して、面接の予約、取り消し、完了を管理できます。
AIコーディングです。

## 📚 目次

- [機能](#機能)
- [主な特徴](#主な特徴)
- [セットアップ](#セットアップ)
- [ドキュメント](#ドキュメント)
- [トラブルシューティング](#トラブルシューティング)
- [ライセンス](#ライセンス)

## 機能

- **予約作成** (`/reserve`) - 面接の予約を作成
- **予約取り消し** (`/cancel`) - 予約をキャンセル
- **予約完了** (`/complete`) - 予約を完了状態に変更
- **予約一覧表示** (`/list`) - すべての予約を表示（実行者のみに表示）
- **自分の予約確認** (`/my-reservations`) - 自分の予約のみを表示（実行者のみに表示）

## 主な特徴

- ✅ 時間の重複をチェックし、重複する予約を防止
- ✅ 推測しにくい予約IDを自動生成
- ✅ 予約データをJSONファイルに永続化
- ✅ 予約者にはプライベートにIDを通知、他のユーザーには予約情報を公開通知
- ✅ 終了時間を指定しない場合は自動的に開始時刻+1時間を設定
- ✅ **古い予約データの自動クリーンアップ**（完了・キャンセル済み予約を30日後に自動削除）
- ✅ **期限切れ予約の自動完了**（終了時刻が過ぎたpending予約を自動的にcompletedに変更）

## セットアップ

### 1. 前提条件

- Go 1.21以上がインストールされていること
- Discord Botトークンを取得済みであること

### 2. Discord Botの作成

1. [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
2. 「New Application」をクリックして新しいアプリケーションを作成
3. 左側のメニューから「Bot」を選択
4. 「Add Bot」をクリック
5. 「TOKEN」セクションの「Copy」ボタンをクリックしてトークンをコピー
6. 「Privileged Gateway Intents」セクションで以下を有効化：
   - Server Members Intent
   - Message Content Intent

### 3. Botをサーバーに招待

1. Developer Portalの「OAuth2」→「URL Generator」を選択
2. 「SCOPES」で `bot` と `applications.commands` を選択
3. 「BOT PERMISSIONS」で以下を選択：
   - Send Messages
   - Use Slash Commands
   - Read Message History
4. 生成されたURLをブラウザで開き、Botをサーバーに招待

### 4. プロジェクトのセットアップ（自動セットアップ）

**推奨方法: セットアップスクリプトを使用**

```bash
# プロジェクトディレクトリに移動
cd hxs_reservation_system

# 自動セットアップスクリプトを実行
./setup.sh
```

このスクリプトは以下を自動で実行します：
- Goのバージョン確認
- 依存関係のダウンロードと検証
- `.env`ファイルの作成
- ビルドテスト

**または手動セットアップ:**

```bash
# Makefileを使用（推奨）
make setup

# または従来の方法
go mod download
cp config/.env.example .env
```

### 5. 環境変数の設定

`.env`ファイルを編集して、必要な情報を設定します：

```env
DISCORD_TOKEN=your_discord_bot_token_here
GUILD_ID=your_guild_id_here
```

- `DISCORD_TOKEN`: Discord Developer Portalで取得したBotトークン
- `GUILD_ID`: BotをテストするDiscordサーバーのID（開発者モードで右クリック→「IDをコピー」）

**注意**: `GUILD_ID`を設定すると、そのサーバー専用のコマンドとして即座に登録されます。空欄にするとグローバルコマンドとして登録されますが、反映に最大1時間かかります。

### 6. 開発環境と本番環境の切り替え

このプロジェクトは開発環境と本番環境を分離できます：

```bash
# 開発環境に切り替え
./switch_env.sh development

# 本番環境に切り替え
./switch_env.sh production
```

詳細は [クイックスタートガイド](docs/QUICKSTART.md) を参照してください。
./switch_env.sh development

# 本番環境に切り替え
./switch_env.sh production
```

各環境用の設定ファイル：
- `.env.development` - 開発環境用
- `.env.production` - 本番環境用

## 使用方法

### Makefileコマンド一覧

プロジェクトには便利なMakefileコマンドが用意されています：

```bash
make help          # 利用可能なコマンドを表示
make setup         # 初回セットアップ
make deps          # 依存関係をダウンロード
make install       # 依存関係をインストール
make build         # ビルド
make run           # 実行
make start         # ビルドしてから実行
make dev           # 開発モード（ホットリロード、airが必要）
make clean         # ビルド成果物を削除
make fmt           # コードフォーマット
make vet           # 静的解析
make check         # フォーマット+静的解析
make test          # テスト実行
```

### Botの起動方法

**推奨: Makefileを使用**

```bash
# 開発モードで実行（変更を監視）
make run

# ビルドしてから実行
make build
make start

# または一度にビルド＋実行
./bin/hxs_reservation_system
```

**従来の方法:**

```bash
# 直接実行
go run main.go

# ビルドして実行
go build -o hxs_reservation_system
./hxs_reservation_system
```

### 依存関係の管理

依存関係管理スクリプトを使用できます：

```bash
# 依存関係をインストール
./manage_deps.sh install

# 依存関係を更新
./manage_deps.sh update

# 依存関係を一覧表示
./manage_deps.sh list

# 依存関係を検証
./manage_deps.sh verify

# ヘルプを表示
./manage_deps.sh help
```

### コマンドの使用方法

#### 1. 予約作成 (`/reserve`)

面接の予約を作成します。

**必須オプション:**
- `date`: 予約日（YYYY/MM/DD形式、例: 2025/10/15）
- `start_time`: 開始時間（HH:MM形式、例: 14:00）

**任意オプション:**
- `end_time`: 終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間
- `comment`: コメント

**使用例:**
```
/reserve date:2025/10/15 start_time:14:00 end_time:15:00 comment:技術面接
```

**動作:**
- 予約者には予約IDがプライベート（Ephemeral）メッセージで通知されます
- チャンネルの全員には予約情報（予約IDを除く）が公開されます
- 時間が重複している場合は予約できず、エラーメッセージが表示されます

#### 2. 予約取り消し (`/cancel`)

予約をキャンセルします。

**必須オプション:**
- `reservation_id`: 予約ID

**任意オプション:**
- `comment`: コメント

**使用例:**
```
/cancel reservation_id:a1b2c3d4e5f6g7h8 comment:都合が悪くなったため
```

**動作:**
- チャンネルの全員に予約取り消しが通知されます

#### 3. 予約完了 (`/complete`)

予約を完了状態にします。

**必須オプション:**
- `reservation_id`: 予約ID

**任意オプション:**
- `comment`: コメント

**使用例:**
```
/complete reservation_id:a1b2c3d4e5f6g7h8 comment:面接完了
```

**動作:**
- チャンネルの全員に予約完了が通知されます

#### 4. 予約一覧表示 (`/list`)

すべての予約を表示します。

**使用例:**
```
/list
```

**動作:**
- コマンドを実行した人だけに予約一覧が表示されます（Ephemeral）
- 予約者のユーザーID、予約ID、状態などが表示されます

#### 5. 自分の予約確認 (`/my-reservations`)

自分の予約のみを表示します。

**使用例:**
```
/my-reservations
```

**動作:**
- コマンドを実行した人の予約のみが表示されます（Ephemeral）

## データの保存

予約データは `reservations.json` ファイルに自動的に保存されます。

- Bot起動時に既存のデータを読み込みます
- 予約の作成・更新・削除時に即座に保存されます
- 5分ごとに定期的に保存されます
- Bot終了時にも保存されます

## プロジェクト構造

```
hxs_reservation_system/
├── main.go                   # エントリーポイント、Bot初期化
├── go.mod                    # Go モジュール定義
├── go.sum                    # 依存関係のチェックサム
├── Makefile                  # ビルドとタスク管理
├── setup.sh                  # 自動セットアップスクリプト
├── manage_deps.sh            # 依存関係管理スクリプト
├── switch_env.sh             # 環境切り替えスクリプト
├── .air.toml                 # ホットリロード設定（開発用）
├── .env                      # 環境変数（gitignoreに含まれる）
├── .env.example              # 環境変数のテンプレート
├── .env.development          # 開発環境用設定
├── .env.production           # 本番環境用設定
├── .gitignore                # Git除外設定
├── reservations.json         # 予約データ（自動生成）
├── bin/                      # ビルド成果物
├── models/
│   └── reservation.go        # 予約データモデルとビジネスロジック
├── storage/
│   └── storage.go            # データ永続化処理
└── commands/
    └── handlers.go           # コマンドハンドラー
```

## 開発ワークフロー

### 初めて開発を始める場合

```bash
# 1. セットアップ
./setup.sh

# 2. 環境変数を設定
vi .env

# 3. 開発環境で実行
make run
```

### 日常の開発フロー

```bash
# コードを編集後、フォーマットと検証
make check

# 実行して動作確認
make run

# ビルドして配布用バイナリを作成
make build
```

### 環境を切り替える

```bash
# 開発環境に切り替え
./switch_env.sh development
make run

# 本番環境に切り替え
./switch_env.sh production
make start
```

### 依存関係を更新する

```bash
# 依存関係を最新版に更新
./manage_deps.sh update

# または
make install
```

### クリーンビルド

```bash
# すべてをクリーンにしてビルド
make clean
make build
```

## ホットリロード（開発効率化）

開発時にファイルの変更を自動検知して再起動する機能を利用できます：

```bash
# airをインストール
go install github.com/cosmtrek/air@latest

# ホットリロードで起動
make dev
```

## トラブルシューティング

### コマンドが表示されない

- Botが正しくサーバーに招待されているか確認
- `GUILD_ID`が正しく設定されているか確認
- Botに適切な権限が付与されているか確認

### 予約データが消える

- `reservations.json`ファイルが削除されていないか確認
- ファイルの書き込み権限があるか確認

### Botが起動しない

- `DISCORD_TOKEN`が正しく設定されているか確認
- `.env`ファイルが正しい場所に配置されているか確認
- Goのバージョンが1.21以上か確認

### 依存関係のエラー

```bash
# 依存関係をクリーンアップして再インストール
./manage_deps.sh clean
./manage_deps.sh install

# または
make clean
make install
```

### ビルドエラー

```bash
# Go モジュールを整理
go mod tidy

# コードをフォーマット
make fmt

# 静的解析で問題をチェック
make vet
```

## データ管理

### 自動クリーンアップ機能

このBotには、予約データが永久的に蓄積しないように、自動クリーンアップ機能が実装されています。

**機能:**
1. **期限切れ予約の自動完了**（1日1回）
   - 終了時刻が過ぎた `pending` 予約を自動的に `completed` に変更

2. **古い予約の自動削除**（1日1回）
   - `completed` または `cancelled` ステータスの予約で、最終更新から **30日以上** 経過したものを自動削除

詳細は [クリーンアップガイド](docs/CLEANUP.md) を参照してください。

## 📚 ドキュメント

より詳しい情報は、以下のドキュメントを参照してください：

### 📖 基本ガイド
- **[クイックスタートガイド](docs/QUICKSTART.md)** - 最速でBotを起動する方法
- **[開発ガイド](docs/DEVELOPMENT.md)** - 開発環境のセットアップと開発フロー
- **[プロジェクト概要](docs/PROJECT_SUMMARY.md)** - プロジェクトの全体像と構造
- **[プロジェクト構造](docs/PROJECT_STRUCTURE.md)** - ディレクトリ構造とファイル配置の詳細

### 🔧 運用ガイド
- **[クリーンアップガイド](docs/CLEANUP.md)** - 予約データの自動クリーンアップ機能
- **[ロギングガイド](docs/LOGGING.md)** - ログ機能とコマンド統計
- **[systemd セットアップ](docs/SYSTEMD_SETUP.md)** - Linux サーバーでの自動起動設定（詳細版）
- **[systemd クイックリファレンス](docs/SYSTEMD_QUICK_REFERENCE.md)** - サービスファイルの設定早見表

### 📁 設定ファイル
- **[config/](config/)** - 設定ファイルのサンプルとテンプレート
  - `.env.example` - 環境変数の設定例
  - `.env.development` - 開発環境用設定
  - `.env.production` - 本番環境用設定
  - `hxs-reservation-bot.service` - **systemd サービスファイル**（本番運用時に使用）
  - `.air.toml` - ホットリロード設定

**サービスファイルの使用方法**:
- Linux サーバーで自動起動させる場合は、`config/hxs-reservation-bot.service` を `/etc/systemd/system/` にコピーして使用します
- 詳細は [systemd セットアップガイド](docs/SYSTEMD_SETUP.md) を参照してください

### 🛠️ スクリプト
- **`setup.sh`** - 自動セットアップスクリプト
- **`switch_env.sh`** - 環境切り替えスクリプト
- **`manage_deps.sh`** - 依存関係管理スクリプト
- **`setup-systemd.sh`** - systemd セットアップスクリプト

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 貢献

バグ報告や機能追加の提案は、GitHubのIssuesでお願いします。

## 作者

dice
