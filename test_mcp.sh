#!/bin/bash

echo "🚀 Testing COSCUP MCP Server..."
echo ""

# 檢查二進制檔案是否存在
if [ ! -f "./bin/coscup-mcp-server" ]; then
    echo "❌ Binary not found. Please run 'make build' first."
    exit 1
fi

echo "✅ Binary found"

# 測試服務器啟動
echo "🔄 Testing server startup..."
echo "" | timeout 2s ./bin/coscup-mcp-server 2>&1 | grep -q "MCP Server is ready"
if [ $? -eq 0 ]; then
    echo "✅ Server starts successfully"
else
    echo "❌ Server failed to start"
    exit 1
fi

echo ""
echo "📋 MCP Configuration for Claude Code:"
echo "File: ~/.claude/claude_desktop_config.json"
echo ""
cat ~/.claude/claude_desktop_config.json
echo ""
echo "🎯 Setup completed! Restart Claude Code and the tools should be available."
echo ""
echo "Available tools:"
echo "  - start_planning: 開始規劃某天的行程"
echo "  - choose_session: 選擇並記錄議程"
echo "  - get_options: 取得下個時段的推薦選項"
echo "  - mcp_ask: 查看目前的規劃狀態"