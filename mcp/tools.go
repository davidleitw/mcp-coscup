package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateMCPTools creates and returns all MCP tools
func CreateMCPTools() map[string]mcp.Tool {
	return map[string]mcp.Tool{
		"start_planning": createStartPlanningTool(),
		"choose_session": createChooseSessionTool(),
		"get_options":    createGetOptionsTool(),
		"mcp_ask":        createMCPAskTool(),
	}
}

// 1. Start Planning Tool
func createStartPlanningTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"day": map[string]any{
				"type":        "string",
				"description": "要規劃的日期，必須是 'Aug.9' 或 'Aug.10'",
				"enum":        []string{"Aug.9", "Aug.10"},
			},
			"call_reason": map[string]any{
				"type":        "string",
				"description": "說明你為什麼要呼叫這個工具，以及預期得到什麼結果",
			},
		},
		"required": []string{"day", "call_reason"},
	}

	return mcp.NewTool(
		"start_planning",
		"開始規劃 COSCUP 某天的行程。作為 LLM，當用戶說想安排某天行程時，請使用此工具開始規劃流程。使用後你會收到當天最早的議程選項，請向用戶友善地介紹這些選項並徵求意見。",
		schema,
	)
}

func handleStartPlanning(args map[string]any) (*mcp.CallToolResult, error) {
	day, ok := args["day"].(string)
	if !ok || !IsValidDay(day) {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：日期必須是 'Aug.9' 或 'Aug.10'"),
			},
		}, nil
	}

	callReason, _ := args["call_reason"].(string)

	// Generate a simple session ID
	sessionID := fmt.Sprintf("user_%s_%d",
		map[string]string{"Aug.9": "09", "Aug.10": "10"}[day],
		len(userStates)+1)

	// Create new user state
	CreateUserState(sessionID, day)

	// Get first sessions of the day
	firstSessions := GetFirstSession(day)
	if len(firstSessions) == 0 {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent(fmt.Sprintf("錯誤：找不到 %s 的議程資料", day)),
			},
		}, nil
	}

	response := Response{
		Success: true,
		Data: map[string]any{
			"session_id": sessionID,
			"day":        day,
			"options":    firstSessions,
		},
		CallReason: callReason,
		Message: fmt.Sprintf("已開始規劃 %s 的行程，session ID: %s。請向用戶介紹這 %d 個早上的議程選項，並請他們選擇感興趣的。",
			day, sessionID, len(firstSessions)),
	}

	return &mcp.CallToolResult{
		Content: []any{
			mcp.NewTextContent(fmt.Sprintf("%+v", response)),
		},
	}, nil
}

// 2. Choose Session Tool
func createChooseSessionTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "用戶的 session ID",
			},
			"session_code": map[string]any{
				"type":        "string",
				"description": "用戶選擇的議程代碼",
			},
			"call_reason": map[string]any{
				"type":        "string",
				"description": "說明用戶選擇了什麼，以及你為什麼要記錄這個選擇",
			},
		},
		"required": []string{"session_id", "session_code", "call_reason"},
	}

	return mcp.NewTool(
		"choose_session",
		"記錄用戶選擇的議程到他們的行程中。當用戶明確表示要選擇某個議程時使用此工具。工具會：1. 將議程加入用戶行程 2. 更新用戶興趣 profile 3. 自動推薦下個時段的議程選項。收到回應後，請向用戶確認選擇，並介紹下個時段的推薦選項。",
		schema,
	)
}

func handleChooseSession(args map[string]any) (*mcp.CallToolResult, error) {
	sessionID, _ := args["session_id"].(string)
	sessionCode, _ := args["session_code"].(string)
	callReason, _ := args["call_reason"].(string)

	if sessionID == "" || sessionCode == "" {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：session_id 和 session_code 都是必要參數"),
			},
		}, nil
	}

	// Add session to user's schedule
	err := AddSessionToSchedule(sessionID, sessionCode)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent(fmt.Sprintf("錯誤：%s", err.Error())),
			},
		}, nil
	}

	// Get selected session details
	selectedSession := FindSessionByCode(sessionCode)
	if selectedSession == nil {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：找不到所選議程的詳細資訊"),
			},
		}, nil
	}

	// Get next recommendations
	recommendations, _ := GetRecommendations(sessionID, 5)

	var nextMessage string
	if len(recommendations) == 0 {
		if IsScheduleComplete(sessionID) {
			nextMessage = "太棒了！您的行程規劃已完成。請使用 mcp_ask 查看完整行程。"
		} else {
			nextMessage = "目前沒有更多可選的議程了。"
		}
	} else {
		nextMessage = fmt.Sprintf("已記錄選擇！現在請為用戶推薦下個時段的 %d 個議程選項。", len(recommendations))
	}

	response := Response{
		Success: true,
		Data: map[string]any{
			"selected_session": selectedSession,
			"next_options":     recommendations,
			"is_complete":      IsScheduleComplete(sessionID),
		},
		CallReason: callReason,
		Message:    nextMessage,
	}

	return &mcp.CallToolResult{
		Content: []any{
			mcp.NewTextContent(fmt.Sprintf("%+v", response)),
		},
	}, nil
}

// 3. Get Options Tool
func createGetOptionsTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "用戶的 session ID",
			},
			"call_reason": map[string]any{
				"type":        "string",
				"description": "說明你為什麼需要取得選項，例如：需要推薦下個時段的議程",
			},
		},
		"required": []string{"session_id", "call_reason"},
	}

	return mcp.NewTool(
		"get_options",
		"取得用戶當前可選的議程選項。當你需要為用戶推薦下個時段的議程時使用此工具。工具會根據用戶的：- 上次選擇的議程結束時間 - 興趣 profile (從之前的選擇學習)。回傳最多5個推薦的議程。請根據推薦分數排序並友善地介紹給用戶。",
		schema,
	)
}

func handleGetOptions(args map[string]any) (*mcp.CallToolResult, error) {
	sessionID, _ := args["session_id"].(string)
	callReason, _ := args["call_reason"].(string)

	if sessionID == "" {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：session_id 是必要參數"),
			},
		}, nil
	}

	state := GetUserState(sessionID)
	if state == nil {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：找不到指定的 session"),
			},
		}, nil
	}

	recommendations, err := GetRecommendations(sessionID, 5)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent(fmt.Sprintf("錯誤：%s", err.Error())),
			},
		}, nil
	}

	var message string
	if len(recommendations) == 0 {
		message = "當前沒有可選的議程。可能已經完成今日規劃，或沒有更多適合的時段。"
	} else {
		message = fmt.Sprintf("找到 %d 個推薦議程，已按照用戶興趣排序。請友善地介紹這些選項。", len(recommendations))
	}

	response := Response{
		Success: true,
		Data: map[string]any{
			"options":                recommendations,
			"user_profile":           state.Profile,
			"last_end_time":          state.LastEndTime,
			"current_schedule_count": len(state.Schedule),
		},
		CallReason: callReason,
		Message:    message,
	}

	return &mcp.CallToolResult{
		Content: []any{
			mcp.NewTextContent(fmt.Sprintf("%+v", response)),
		},
	}, nil
}

// 4. MCP Ask Tool
func createMCPAskTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "用戶的 session ID",
			},
			"question": map[string]any{
				"type":        "string",
				"description": "你想確認什麼？描述你的困惑或需要驗證的點",
			},
			"call_reason": map[string]any{
				"type":        "string",
				"description": "簡述為什麼需要確認，例如：用戶想查看完整行程",
			},
		},
		"required": []string{"session_id", "question", "call_reason"},
	}

	return mcp.NewTool(
		"mcp_ask",
		"自我檢查和狀態確認工具。使用時機：- 當你對上一個操作的結果不確定時 - 需要查看用戶完整行程時 - 想確認當前狀態是否正確時 - 用戶詢問他們目前的規劃狀況時。這個工具會回傳完整的用戶狀態，幫助你重新評估並決定下一步行動。",
		schema,
	)
}

func handleMCPAsk(args map[string]any) (*mcp.CallToolResult, error) {
	sessionID, _ := args["session_id"].(string)
	question, _ := args["question"].(string)
	callReason, _ := args["call_reason"].(string)

	if sessionID == "" {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：session_id 是必要參數"),
			},
		}, nil
	}

	state := GetUserState(sessionID)
	if state == nil {
		return &mcp.CallToolResult{
			Content: []any{
				mcp.NewTextContent("錯誤：找不到指定的 session"),
			},
		}, nil
	}

	// Get available options for context
	nextOptions, _ := GetRecommendations(sessionID, 5)

	response := Response{
		Success: true,
		Data: map[string]any{
			"current_state": map[string]any{
				"session_id":     state.SessionID,
				"day":            state.Day,
				"schedule":       state.Schedule,
				"schedule_count": len(state.Schedule),
				"last_end_time":  state.LastEndTime,
				"user_profile":   state.Profile,
				"is_complete":    IsScheduleComplete(sessionID),
			},
			"next_options": nextOptions,
			"question":     question,
		},
		CallReason: callReason,
		Message: fmt.Sprintf("✅ 當前狀態確認完成。用戶已選擇 %d 個議程，最後結束時間 %s。根據這些資訊回答用戶的問題。",
			len(state.Schedule), state.LastEndTime),
	}

	return &mcp.CallToolResult{
		Content: []any{
			mcp.NewTextContent(fmt.Sprintf("%+v", response)),
		},
	}, nil
}

// GetToolHandlers returns a map of tool names to their handlers
func GetToolHandlers() map[string]server.ToolHandlerFunc {
	return map[string]server.ToolHandlerFunc{
		"start_planning": handleStartPlanning,
		"choose_session": handleChooseSession,
		"get_options":    handleGetOptions,
		"mcp_ask":        handleMCPAsk,
	}
}
