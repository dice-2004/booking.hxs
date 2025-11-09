#!/bin/bash

# 環境を切り替えるスクリプト

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

ENV=${1:-"development"}

if [ "$ENV" != "development" ] && [ "$ENV" != "production" ]; then
    echo -e "${RED}エラー: 無効な環境名です${NC}"
    echo "使用方法: ./switch_env.sh [development|production]"
    exit 1
fi

echo -e "${YELLOW}環境を ${ENV} に切り替え中...${NC}"

if [ ! -f "config/.env.${ENV}" ]; then
    echo -e "${RED}エラー: config/.env.${ENV} ファイルが見つかりません${NC}"
    exit 1
fi

# バックアップを作成
if [ -f ".env" ]; then
    cp .env .env.backup
    echo -e "${GREEN}✓ 既存の .env をバックアップしました (.env.backup)${NC}"
fi

# 環境ファイルをコピー
cp "config/.env.${ENV}" .env
echo -e "${GREEN}✓ config/.env.${ENV} を .env にコピーしました${NC}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ 環境を ${ENV} に切り替えました${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${YELLOW}現在の環境変数:${NC}"
grep -v '^#' .env | grep -v '^$'
