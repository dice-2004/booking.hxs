#!/bin/bash

# Go プロジェクトの依存関係管理スクリプト

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== Go 依存関係管理 ===${NC}\n"

# コマンドライン引数のパース
COMMAND=${1:-"install"}

case $COMMAND in
  "install"|"i")
    echo -e "${YELLOW}依存関係をインストール中...${NC}"
    go mod download
    go mod tidy
    go mod verify
    echo -e "${GREEN}✓ 依存関係のインストール完了${NC}"
    ;;
    
  "update"|"u")
    echo -e "${YELLOW}依存関係を更新中...${NC}"
    go get -u ./...
    go mod tidy
    echo -e "${GREEN}✓ 依存関係の更新完了${NC}"
    ;;
    
  "clean"|"c")
    echo -e "${YELLOW}依存関係のキャッシュをクリア中...${NC}"
    go clean -modcache
    echo -e "${GREEN}✓ キャッシュのクリア完了${NC}"
    ;;
    
  "verify"|"v")
    echo -e "${YELLOW}依存関係を検証中...${NC}"
    go mod verify
    echo -e "${GREEN}✓ 依存関係の検証完了${NC}"
    ;;
    
  "list"|"l")
    echo -e "${YELLOW}依存関係の一覧:${NC}"
    go list -m all
    ;;
    
  "graph"|"g")
    echo -e "${YELLOW}依存関係のグラフ:${NC}"
    go mod graph
    ;;
    
  "why"|"w")
    if [ -z "$2" ]; then
      echo -e "${RED}エラー: パッケージ名を指定してください${NC}"
      echo "使用例: ./manage_deps.sh why github.com/bwmarrin/discordgo"
      exit 1
    fi
    echo -e "${YELLOW}なぜ $2 が必要か:${NC}"
    go mod why "$2"
    ;;
    
  "help"|"h"|*)
    echo "使用方法: ./manage_deps.sh [command]"
    echo ""
    echo "利用可能なコマンド:"
    echo "  install, i    - 依存関係をインストール"
    echo "  update, u     - 依存関係を最新版に更新"
    echo "  clean, c      - 依存関係のキャッシュをクリア"
    echo "  verify, v     - 依存関係を検証"
    echo "  list, l       - 依存関係の一覧を表示"
    echo "  graph, g      - 依存関係のグラフを表示"
    echo "  why, w <pkg>  - なぜそのパッケージが必要か表示"
    echo "  help, h       - このヘルプを表示"
    ;;
esac
