# 開発環境管理ガイド

このドキュメントでは、Go言語プロジェクトを「仮想環境のように」管理する方法を説明します。

## 🎯 概要

Pythonの仮想環境（venv）のように、このプロジェクトは以下の特徴を持っています：

- ✅ **プロジェクト独立性**: `go.mod`でプロジェクト固有の依存関係を管理
- ✅ **環境分離**: 開発環境と本番環境を分離
- ✅ **簡単なセットアップ**: 1コマンドでセットアップ完了
- ✅ **依存関係の明示**: `go.mod`と`go.sum`で完全に管理
- ✅ **自動化されたワークフロー**: Makefileで一貫したコマンド

## 📦 依存関係の管理

### go.modとgo.sum

Goでは`go.mod`と`go.sum`ファイルで依存関係を管理します：

- **go.mod**: プロジェクトの依存関係を定義
- **go.sum**: 依存関係のチェックサムを記録（セキュリティ）

これらはPythonの`requirements.txt`やNode.jsの`package.json`に相当します。

### 依存関係の操作

#### インストール
```bash
# 推奨: Makefileを使用
make install

# または: スクリプトを使用
./manage_deps.sh install

# または: 直接実行
go mod download
go mod tidy
```

#### 更新
```bash
# すべての依存関係を最新版に更新
./manage_deps.sh update

# または
go get -u ./...
go mod tidy
```

#### 確認
```bash
# 依存関係の一覧
./manage_deps.sh list

# 依存関係のグラフ
./manage_deps.sh graph

# なぜこの依存関係が必要か
./manage_deps.sh why github.com/bwmarrin/discordgo
```

#### クリーンアップ
```bash
# キャッシュをクリア
./manage_deps.sh clean

# または
go clean -modcache
```

## 🔄 環境の切り替え

### 環境ファイル

3つの環境設定ファイルがあります：

1. **`.env.example`** - テンプレート（Git管理対象）
2. **`.env.development`** - 開発環境用
3. **`.env.production`** - 本番環境用

### 環境の切り替え方法

```bash
# 開発環境に切り替え
./switch_env.sh development

# 本番環境に切り替え
./switch_env.sh production
```

このスクリプトは：
1. 現在の`.env`をバックアップ（`.env.backup`）
2. 指定された環境ファイルを`.env`にコピー
3. 現在の環境変数を表示

### 環境ごとの設定例

**開発環境（.env.development）:**
```env
DISCORD_TOKEN=dev_token_here
GUILD_ID=dev_server_id
ENV=development
DATA_FILE=reservations_dev.json
```

**本番環境（.env.production）:**
```env
DISCORD_TOKEN=prod_token_here
GUILD_ID=
ENV=production
DATA_FILE=reservations.json
```

## 🛠️ Makefileによる自動化

### 利用可能なコマンド

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

### Makefileの利点

- **一貫性**: チーム全員が同じコマンドを使用
- **自動化**: 複数のコマンドを1つにまとめる
- **ドキュメント**: コマンドが自己文書化される
- **効率**: タイプ量が減る

## 🔥 ホットリロード（開発効率化）

開発中にファイルの変更を自動検知して再起動：

### セットアップ
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

## 📁 プロジェクト構造

```
hxs_reservation_system/
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
├── switch_env.sh             # 環境切り替えスクリプト
├── .air.toml                 # ホットリロード設定
├── main.go                   # エントリーポイント
├── bin/                      # ビルド成果物
├── models/                   # データモデル
├── storage/                  # データ永続化
└── commands/                 # コマンドハンドラー
```

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
./bin/hxs_reservation_system
```

### Dockerを使用する場合

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o hxs_reservation_system main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/hxs_reservation_system .
COPY .env.production .env
CMD ["./hxs_reservation_system"]
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
./bin/hxs_reservation_system
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
