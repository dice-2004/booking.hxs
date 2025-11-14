.PHONY: help install build run clean test dev deps fmt vet

# デフォルトターゲット
help: ## このヘルプメッセージを表示
	@echo "利用可能なコマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## 依存関係をダウンロード
	@echo "依存関係をダウンロード中..."
	go mod download
	go mod verify
	@echo "✓ 依存関係のダウンロード完了"

install: deps ## 依存関係をインストール（ローカル環境用）
	@echo "依存関係をインストール中..."
	go mod tidy
	@echo "✓ インストール完了"

# ビルド
build:
	@echo "📦 ビルド中..."
	go build -o bin/booking.hxs cmd/bot/main.go
	@echo "✓ ビルド完了: bin/booking.hxs"

run: ## アプリケーションを実行
	@echo "アプリケーションを起動中..."
	go run cmd/bot/main.go

dev: ## 開発モードで実行（ホットリロード用）
	@echo "開発モードで起動中..."
	@if command -v air > /dev/null; then \
		air -c config/.air.toml; \
	else \
		echo "air がインストールされていません。通常モードで起動します..."; \
		go run cmd/bot/main.go; \
	fi

clean: ## ビルド成果物とキャッシュを削除
	@echo "クリーニング中..."
	rm -rf bin/
	rm -f reservations.json
	go clean -cache -testcache -modcache
	@echo "✓ クリーニング完了"

fmt: ## コードをフォーマット
	@echo "コードをフォーマット中..."
	go fmt ./...
	@echo "✓ フォーマット完了"

vet: ## コードを静的解析
	@echo "静的解析中..."
	go vet ./...
	@echo "✓ 静的解析完了"

test: ## テストを実行
	@echo "テスト実行中..."
	go test -v ./...
	@echo "✓ テスト完了"

lint: ## リンターを実行（golangci-lintが必要）
	@echo "リント中..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint がインストールされていません"; \
		echo "インストール: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

setup: ## 初回セットアップ（.envファイルの作成）
	@echo "初回セットアップ中..."
	@if [ ! -f .env ]; then \
		cp config/.env.example .env; \
		echo "✓ .env ファイルを作成しました"; \
		echo "⚠️  .env ファイルを編集して、DISCORD_TOKENとGUILD_IDを設定してください"; \
	else \
		echo "✓ .env ファイルは既に存在します"; \
	fi
	@make install
	@echo "✓ セットアップ完了"

start: build ## ビルドしてから実行
	@echo "アプリケーションを起動中..."
	# 実行（ビルド済みのバイナリを使用）
run-bin: build
	@echo "🚀 実行中..."
	./bin/booking.hxs

check: fmt vet ## フォーマットと静的解析を実行
	@echo "✓ チェック完了"

all: check build ## すべて実行（チェック→ビルド）
	@echo "✓ すべて完了"
