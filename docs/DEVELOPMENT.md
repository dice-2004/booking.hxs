# 💻 開発者ガイド

このドキュメントでは、開発環境のセットアップ、カスタマイズ、拡張方法について説明します。

## 📑 目次



## 🆕 UI/UX仕様・開発ルール（2025年11月更新）

### UI/UX仕様

- すべてのDiscord埋め込みメッセージは「部室予約システム | コマンド名」形式のフッター付きで統一されています。
- `/list`・`/my-reservations`コマンドは「部室予約システム | list | 予約 X/Y」など進捗付きフッターを表示します。
- 予約一覧が10件以上の場合、Ephemeralメッセージ（実行者のみに表示）で複数メッセージに分割して表示されます。
- Ephemeralメッセージは実行者のみに表示され、他のユーザーには見えません。
- コマンド登録時に一部失敗しても他のコマンドは登録され、エラーはログに記録されます。
- 予約情報の表示レイアウトは`/reserve`コマンドのフォーマット（Fields, Inline: true/false）に統一されています。

### コーディング規約

#### main.goの構造

`cmd/bot/main.go`は以下の構造で統一されています：

1. **定数の定義** - ファイル冒頭にすべての設定定数を集約
   ```go
   const (
       saveInterval       = 5 * time.Minute
       autoCompleteHour   = 3
       retentionDays      = 30
   )
   ```

2. **グローバル変数** - 必要最小限に抑える
   ```go
   var (
       store  *storage.Storage
       logger *logging.Logger
   )
   ```

3. **関数の分割** - 各関数は単一責任を持つ
   - `initializeServices()`: サービス初期化
   - `setupHandlers()`: イベントハンドラー設定
   - `startBackgroundTasks()`: バックグラウンドタスク起動

#### コマンドハンドラーの構造

- 各コマンドは独立したファイル（`internal/commands/cmd_*.go`）で管理
- `handlers.go`はルーティングのみを担当
- 共通処理は`response_helpers.go`に集約

---

## Go Modulesによる依存関係管理

このプロジェクトは **Go Modules** を使用しています。Pythonの仮想環境のように、プロジェクト固有の依存関係を管理します。

#### 主要ファイル

| ファイル | 説明 | Pythonの相当物 |
|---------|------|---------------|
| `go.mod` | 依存関係の定義 | `requirements.txt` |
| `go.sum` | チェックサム | `requirements.txt` のハッシュ |

### プロジェクトの独立性

✅ **プロジェクト固有の依存関係** - `go.mod` で管理
✅ **環境分離** - 開発/本番環境を分離
✅ **簡単なセットアップ** - 1コマンドで完了
✅ **自動化** - Makefileで一貫したワークフロー

---

## 環境の切り替え

### 環境設定ファイル

| ファイル | 説明 | Git管理 |
|---------|------|---------|
| `.env.example` | テンプレート | ✅ Yes |
| `.env.development` | 開発環境用 | ✅ Yes |
| `.env.production` | 本番環境用 | ✅ Yes |
| `.env` | 現在使用中 | ❌ No (.gitignore) |

### 環境切り替えスクリプト

```bash
# 開発環境に切り替え
./switch_env.sh development

# 本番環境に切り替え
./switch_env.sh production

# 現在の環境を確認
./switch_env.sh status
```

### スクリプトの動作

1. 現在の `.env` を `.env.backup` にバックアップ
2. 指定された環境ファイルを `.env` にコピー
3. 現在の環境変数を表示

### 環境ごとの設定例

**開発環境（.env.development）**
```env
DISCORD_TOKEN=dev_token_here
GUILD_ID=dev_server_id
FEEDBACK_CHANNEL_ID=dev_feedback_channel_id
ENV=development
```

**本番環境（.env.production）**
```env
DISCORD_TOKEN=prod_token_here
GUILD_ID=
FEEDBACK_CHANNEL_ID=prod_feedback_channel_id
ENV=production
```

**注**: `DATA_FILE` 環境変数は使用されません。データは常に `data/reservations.json` に保存されます。

---

## 依存関係の管理

### 依存関係管理スクリプト

`manage_deps.sh` スクリプトで依存関係を管理できます。

```bash
# インストール
./manage_deps.sh install

# 更新
./manage_deps.sh update

# 一覧表示
./manage_deps.sh list

# 依存関係のグラフ
./manage_deps.sh graph

# 特定の依存関係を調査
./manage_deps.sh why github.com/bwmarrin/discordgo

# クリーンアップ
./manage_deps.sh clean

# ヘルプ
./manage_deps.sh help
```

### Go Modulesコマンド

```bash
# 依存関係をダウンロード
go mod download

# 不要な依存関係を削除
go mod tidy

# 依存関係を最新版に更新
go get -u ./...

# キャッシュをクリア
go clean -modcache

# 依存関係の一覧
go list -m all

# 依存関係のグラフ
go mod graph
```

---

## 開発ワークフロー

### Makefileコマンド一覧

```bash
make help          # すべてのコマンドを表示
make setup         # 初回セットアップ
make deps          # 依存関係ダウンロード
make install       # 依存関係インストール
make build         # ビルド
make run           # 実行
make start         # ビルド→実行
make dev           # 開発モード（ホットリロード）
make clean         # クリーンアップ
make fmt           # コードフォーマット
make vet           # 静的解析
make check         # fmt + vet
make test          # テスト実行
make all           # check + build
```

### 日常の開発フロー

```bash
# 1. コードを編集
vi internal/commands/handlers.go

# 2. フォーマット＋静的解析
make check

# 3. 実行して動作確認
make run

# 4. ビルドして配布用バイナリ作成
make build
```

---

## ホットリロード

開発時に、ファイルの変更を自動検知して再起動する機能を利用できます。

### airのインストール

```bash
go install github.com/cosmtrek/air@latest
```

### ホットリロードで起動

```bash
make dev

# または
air
```

### 設定ファイル

`.air.toml` に設定があります：

```toml
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "vendor", "bin"]
```

---

## コードの品質管理

### コードフォーマット

```bash
# フォーマット
make fmt

# または
go fmt ./...
```

### 静的解析

```bash
# 静的解析
make vet

# または
go vet ./...
```

### フォーマット＋静的解析

```bash
make check
```

### テストの実行

```bash
make test

# または
go test ./...
```

---

## プロジェクト構造

```
booking.hxs/
├── go.mod / go.sum            # 依存関係管理
│
├── cmd/                       # アプリケーションエントリーポイント
│   └── bot/                   # Discord Botアプリケーション
│       └── main.go            # メインエントリーポイント
│
├── internal/                  # プライベートアプリケーションコード
│   ├── commands/              # コマンドハンドラー
│   │   ├── handlers.go        # インタラクション処理のルーティング
│   │   ├── autocomplete.go    # オートコンプリート
│   │   ├── cmd_reserve.go     # /reserve コマンド
│   │   ├── cmd_cancel.go      # /cancel コマンド
│   │   ├── cmd_complete.go    # /complete コマンド
│   │   ├── cmd_edit.go        # /edit コマンド
│   │   ├── cmd_list.go        # /list コマンド
│   │   ├── cmd_my_reservations.go # /my-reservations コマンド
│   │   ├── cmd_help.go        # /help コマンド
│   │   ├── cmd_feedback.go    # /feedback コマンド
│   │   └── response_helpers.go # レスポンス共通関数
│   │
│   ├── models/                # データモデル
│   │   └── reservation.go     # 予約データ構造
│   │
│   ├── storage/               # データ永続化
│   │   ├── storage.go         # JSON読み書き、クリーンアップ
│   │   └── storage_test.go    # ストレージテスト
│   │
│   └── logging/               # ログ管理
│       └── logger.go          # コマンドログ、統計
│
├── bin/                       # ビルド成果物
│   └── booking.hxs            # ビルド済みバイナリ
│
├── data/                      # データファイル
│   └── reservations.json      # 予約データ（自動生成）
│
├── logs/                      # ログファイル（自動生成）
│   ├── commands_YYYY-MM.log   # 月別コマンドログ
│   └── command_stats.json     # コマンド統計
│
├── config/                    # 設定ファイル
│   ├── .env.example           # 環境変数テンプレート
│   ├── .env.development       # 開発環境
│   ├── .env.production        # 本番環境
│   ├── booking-hxs.service    # systemdサービス
│   └── .air.toml              # ホットリロード設定
│
├── docs/                      # ドキュメント
│   ├── SETUP.md               # 起動ガイド
│   ├── COMMANDS.md            # コマンドリファレンス
│   ├── DATA_MANAGEMENT.md     # データ管理
│   ├── SYSTEMD.md             # systemdセットアップ
│   ├── DEVELOPMENT.md         # 開発者ガイド（本ファイル）
│   ├── CHANGELOG.md           # 変更履歴
│   ├── RELEASE_NOTES.md       # リリースノート一覧
│   └── releases/              # 各バージョンのリリースノート
│
├── Makefile                   # ビルドタスク
├── .env                      # 現在の環境設定（Git除外）
├── .env.example              # 設定テンプレート
├── .env.development          # 開発環境設定
├── .env.production           # 本番環境設定
├── .gitignore                # Git除外ファイル
├── go.mod                    # 依存関係定義
├── go.sum                    # 依存関係チェックサム
├── Makefile                  # タスク自動化
├── setup.sh                  # セットアップスクリプト
├── manage_deps.sh            # 依存関係管理スクリプト
└── switch_env.sh             # 環境切り替えスクリプト

```

### ディレクトリ構造の設計思想

このプロジェクトは、Goコミュニティで推奨される標準的なプロジェクトレイアウトに従っています:

- **`cmd/bot/`**: Botアプリケーションのエントリーポイント
  - コマンド登録、インタラクションハンドリング、定期タスクなど
  - 将来的にCLIツールや管理ツールを`cmd/`に追加可能

- **`internal/`**: プライベートアプリケーションコード
  - Goの特別なディレクトリ（外部パッケージからインポート不可）
  - `commands/`: Discord コマンドのハンドラー群（コマンドごとに分割）
  - `models/`: データモデル定義
  - `storage/`: データ永続化ロジック
  - `logging/`: ロギング機能

この構造により、コードの保守性と拡張性が向上します。各コマンドが独立したファイルで管理されているため、機能追加や修正が容易です。

---

## カスタマイズ

### 新しいコマンドを追加

#### 1. コマンド定義を追加（cmd/bot/main.go）

```go
commands := []*discordgo.ApplicationCommand{
    // ... 既存のコマンド
    {
        Name:        "your-new-command",
        Description: "コマンドの説明",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "param1",
                Description: "パラメータの説明",
                Required:    true,
            },
        },
    },
}
```

#### 2. ルーティングを追加（internal/commands/handlers.go）

```go
func HandleInteraction(...) {
    switch commandName {
    // ... 既存のケース
    case "your-new-command":
        handleYourNewCommand(s, i, store, logger, allowedChannelID, isDM)
    }
}
```

#### 3. コマンドハンドラーファイルを作成（internal/commands/cmd_your_new_command.go）

```go
package commands

import (
    "github.com/bwmarrin/discordgo"
    "github.com/dice/hxs_reservation_system/internal/logging"
    "github.com/dice/hxs_reservation_system/internal/storage"
)

// handleYourNewCommand は新しいコマンドを処理する
func handleYourNewCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
    options := i.ApplicationCommandData().Options
    optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
    for _, opt := range options {
        optionMap[opt.Name] = opt
    }

    param1 := optionMap["param1"].StringValue()

    // コマンドの処理ロジック

    // レスポンスを返す
    respondEphemeral(s, i, "処理が完了しました")

    // Botステータスを更新（必要な場合）
    if UpdateStatusCallback != nil {
        UpdateStatusCallback()
    }
}
```

#### 4. 再ビルド＆再起動

```bash
make build
make run
```

**ポイント**:
- 各コマンドは独立したファイル（`cmd_*.go`）で管理
- `handlers.go`はルーティングのみを担当
- 共通関数は`response_helpers.go`に配置



---

### データ構造の拡張

#### 予約モデルにフィールドを追加

`internal/models/reservation.go` を編集：

```go
type Reservation struct {
    ID          string             `json:"id"`
    UserID      string             `json:"user_id"`
    Username    string             `json:"username"`
    Date        string             `json:"date"`
    StartTime   string             `json:"start_time"`
    EndTime     string             `json:"end_time"`
    Comment     string             `json:"comment"`
    Status      ReservationStatus  `json:"status"`
    CreatedAt   time.Time          `json:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at"`

    // 新しいフィールドを追加
    Priority    string             `json:"priority"`    // 優先度
    Tags        []string           `json:"tags"`        // タグ
}
```

---

### ログフォーマットのカスタマイズ

`internal/logging/logger.go` でログフォーマットを変更できます。

---

## デバッグ

### デバッグログの有効化

環境変数で設定：

```env
DEBUG=true
LOG_LEVEL=debug
```

### エラーログの確認

```bash
# アプリケーションログ
tail -f logs/commands_2025-11.log | grep '"success":false'

# systemdログ（本番環境）
sudo journalctl -u booking-hxs -f
```

---

## テスト

### ユニットテストの追加

`internal/storage/storage_test.go` などにテストを追加：

```go
func TestCleanupOldReservations(t *testing.T) {
    store := NewStorage()
    // テストコード
}
```

### テストの実行

```bash
make test

# または
go test ./...

# カバレッジ付き
go test -cover ./...
```

---

## Git管理

### .gitignore

以下のファイルはGit管理から除外されています：

- `.env` - 環境変数（機密情報）
- `bin/` - ビルド成果物
- `logs/` - ログファイル
- `data/` - データファイル（`data/reservations.json`等）
- `*.backup` - バックアップファイル

### コミット前のチェック

```bash
# フォーマット＋静的解析
make check

# ビルドテスト
make build

# すべてのテスト
make test
```

---

## まとめ

開発環境のポイント：

✅ **Go Modules** - プロジェクト固有の依存関係管理
✅ **環境分離** - 開発/本番環境を簡単に切り替え
✅ **自動化** - Makefileで一貫したワークフロー
✅ **ホットリロード** - 開発効率を向上
✅ **コード品質** - fmt, vet, testで品質維持
✅ **拡張性** - 新しい機能を簡単に追加

---

**関連ドキュメント**: [README](../README.md) | [起動ガイド](SETUP.md) | [コマンド](COMMANDS.md) | [データ管理](DATA_MANAGEMENT.md) | [systemd](SYSTEMD.md)

```bash
# airをインストール
go install github.com/cosmtrek/air@latest
```

### 使用方法
```bash
# ホットリロードで起動
make dev

# または直接
air
```

設定は`.air.toml`で管理されています。

## 🔒 セキュリティのベストプラクティス

### 秘密情報の管理

1. **絶対にコミットしない**
   - `.env`ファイルは`.gitignore`に含める
   - トークンやパスワードをコードに直接書かない

2. **環境変数を使用**
   ```go
   token := os.Getenv("DISCORD_TOKEN")
   ```

3. **テンプレートを用意**
   - `.env.example`で構造を共有
   - 実際の値は含めない

### Git管理

```gitignore
# 環境変数
.env
.env.backup
.env.local

# ビルド成果物
bin/
tmp/

# データファイル
*.json
reservations*.json
```

## 🚀 デプロイメント

### 本番環境へのデプロイ

```bash
# 1. 本番環境に切り替え
./switch_env.sh production

# 2. 依存関係を確認
./manage_deps.sh verify

# 3. ビルド
make build

# 4. 実行
./bin/booking.hxs
```

### Dockerを使用する場合

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o booking.hxs cmd/bot/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/booking.hxs .
COPY .env.production .env
CMD ["./booking.hxs"]
```

## 💡 開発ワークフロー例

### 新機能の開発

```bash
# 1. 開発環境に切り替え
./switch_env.sh development

# 2. ホットリロードで起動
make dev

# 3. コードを編集（自動で再起動される）

# 4. コミット前のチェック
make check

# 5. ビルドテスト
make build
```

### バグ修正

```bash
# 1. 問題の再現（開発環境）
./switch_env.sh development
make run

# 2. 修正

# 3. 検証
make check
make test

# 4. 本番環境でテスト
./switch_env.sh production
make build
./bin/booking.hxs
```

## 🆘 トラブルシューティング

### 依存関係の問題

```bash
# 完全クリーンアップ
make clean
./manage_deps.sh clean

# 再インストール
./manage_deps.sh install
```

### ビルドエラー

```bash
# モジュールの整理
go mod tidy

# 検証
go mod verify

# フォーマットと静的解析
make check
```

### 環境変数が読み込まれない

```bash
# 現在の環境を確認
cat .env

# 環境を再設定
./switch_env.sh development
```

## 📚 参考資料

- [Go Modules Reference](https://go.dev/ref/mod)
- [Go環境変数](https://pkg.go.dev/os#Getenv)
- [Makefile入門](https://www.gnu.org/software/make/manual/make.html)

## 🎓 まとめ

このプロジェクトでは、Goの標準的な依存関係管理システムと、便利なスクリプトを組み合わせて、Pythonの仮想環境のような使い心地を実現しています：

| 機能 | Python venv | このプロジェクト |
|------|-------------|-----------------|
| プロジェクト分離 | `python -m venv` | `go.mod` |
| 依存関係管理 | `pip install` | `go mod download` |
| 依存関係一覧 | `requirements.txt` | `go.mod` + `go.sum` |
| 環境活性化 | `source venv/bin/activate` | 不要（自動） |
| 環境切り替え | 手動 | `./switch_env.sh` |
| タスク実行 | `python script.py` | `make run` |

Happy Coding! 🚀
