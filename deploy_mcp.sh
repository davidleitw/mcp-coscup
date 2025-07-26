#!/bin/bash

# COSCUP MCP Server - 編譯和部署腳本
# 用途: 自動化編譯、移除舊服務、部署新服務的完整流程

set -e  # 如果任何命令失敗就退出

# 設定變數
MCP_NAME="coscup-arranger"
BINARY_PATH="$(pwd)/bin/coscup-mcp-server"
SCOPE="user"  # user | local | project

# 顏色輸出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 COSCUP MCP Server - 編譯和部署腳本${NC}"
echo "========================================"

# 1. 檢查是否在正確目錄
if [ ! -f "go.mod" ] || [ ! -f "Makefile" ]; then
    echo -e "${RED}❌ 錯誤: 請在專案根目錄執行此腳本${NC}"
    exit 1
fi

echo -e "${YELLOW}📍 當前目錄: $(pwd)${NC}"

# 2. 清理並編譯
echo -e "${BLUE}🔨 步驟 1: 清理和編譯...${NC}"
make clean > /dev/null 2>&1 || true
make build

if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}❌ 編譯失敗: 找不到 $BINARY_PATH${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 編譯完成: $BINARY_PATH${NC}"

# 3. 檢查 MCP 服務器是否存在
echo -e "${BLUE}🔍 步驟 2: 檢查現有 MCP 服務器...${NC}"

# 檢查是否已存在
if claude mcp get "$MCP_NAME" >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  發現現有的 MCP 服務器: $MCP_NAME${NC}"
    
    # 移除現有服務器
    echo -e "${BLUE}🗑️  移除現有服務器...${NC}"
    
    # 嘗試從所有可能的 scope 移除
    for scope in "user" "local" "project"; do
        if claude mcp remove "$MCP_NAME" -s "$scope" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ 已從 $scope scope 移除 $MCP_NAME${NC}"
        fi
    done
else
    echo -e "${GREEN}✅ 沒有發現現有的 MCP 服務器${NC}"
fi

# 4. 添加新的 MCP 服務器
echo -e "${BLUE}➕ 步驟 3: 添加新的 MCP 服務器...${NC}"
claude mcp add "$MCP_NAME" "$BINARY_PATH" -s "$SCOPE"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 成功添加 MCP 服務器: $MCP_NAME (scope: $SCOPE)${NC}"
else
    echo -e "${RED}❌ 添加 MCP 服務器失敗${NC}"
    exit 1
fi

# 5. 驗證部署
echo -e "${BLUE}✨ 步驟 4: 驗證部署...${NC}"

# 檢查服務器是否成功註冊
if claude mcp list | grep -q "$MCP_NAME"; then
    echo -e "${GREEN}✅ MCP 服務器成功註冊!${NC}"
    
    # 顯示當前狀態
    echo -e "${BLUE}當前狀態:${NC}"
    claude mcp list | grep "$MCP_NAME"
    
    # 檢查是否已連接（不強制要求）
    if claude mcp list | grep -q "$MCP_NAME.*Connected"; then
        echo -e "${GREEN}🔗 服務器已成功連接${NC}"
    else
        echo -e "${YELLOW}⏳ 服務器正在連接中，這是正常的...${NC}"
        echo -e "${BLUE}💡 重新啟動 Claude Code 後將自動連接${NC}"
    fi
else
    echo -e "${RED}❌ MCP 服務器註冊失敗${NC}"
    echo -e "${YELLOW}請手動檢查: claude mcp list${NC}"
    exit 1
fi

# 6. 顯示服務器資訊
echo -e "${BLUE}📊 部署資訊:${NC}"
claude mcp get "$MCP_NAME"

echo ""
echo -e "${GREEN}🎉 部署完成!${NC}"
echo -e "${BLUE}可用工具:${NC}"
echo "  - start_planning: 開始規劃某天的行程"
echo "  - choose_session: 選擇並記錄議程"
echo "  - get_options: 取得下個時段的推薦選項"
echo "  - mcp_ask: 查看目前的規劃狀態"
echo ""
echo -e "${YELLOW}💡 測試命令: 在 Claude Code 中說 '幫我安排 Aug.9 COSCUP 的行程'${NC}"