package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SessionID warning for all tools that use sessionId
const sessionIdWarning = `CRITICAL: Response includes sessionId that you MUST preserve and show to user. Never truncate, hide, or modify the sessionId. User data depends on this ID. 

`

// CreateMCPTools creates and returns all MCP tools using new helper functions
func CreateMCPTools() map[string]mcp.Tool {
	return map[string]mcp.Tool{
		"start_planning":     createStartPlanningTool(),
		"choose_session":     createChooseSessionTool(),
		"get_options":        createGetOptionsTool(),
		"get_schedule":       createGetScheduleTool(),
		"get_next_session":   createGetNextSessionTool(),
		"get_session_detail": createGetSessionDetailTool(),
		"finish_planning":    createFinishPlanningTool(),
		"help":               createHelpTool(),
	}
}

// 1. Start Planning Tool - using new API
func createStartPlanningTool() mcp.Tool {
	return mcp.NewTool(
		"start_planning",
		mcp.WithDescription("Start planning COSCUP schedule for a specific day. As an LLM, use this tool when user wants to arrange their daily schedule. After using this tool, you will receive the earliest session options for that day. Please introduce these options to the user in a friendly manner in the user's preferred language and ask for their opinion."),
		mcp.WithString("day", 
			mcp.Description("The day to plan schedule for. Must be 'Aug9' or 'Aug10'"),
			mcp.Enum("Aug9", "Aug10"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain why you are calling this tool and what result you expect to get"),
		),
	)
}

func handleStartPlanning(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	day, err := request.RequireString("day")
	if err != nil || !IsValidDay(day) {
		return mcp.NewToolResultError("Error: day must be 'Aug9' or 'Aug10'"), nil
	}

	callReason := request.GetString("callReason", "")

	// Generate a secure session ID
	dayCode := map[string]string{"Aug9": "09", "Aug10": "10"}[day]
	sessionID := GenerateSessionIDWithCollisionCheck(dayCode)

	// Convert day format and create new user state
	internalDay := convertDayFormat(day)
	CreateUserState(sessionID, internalDay)

	// Get first sessions of the day
	firstSessions := GetFirstSession(internalDay)
	if len(firstSessions) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Error: no session data found for %s", internalDay)), nil
	}

	data := map[string]any{
		"day":     internalDay,
		"options": removeAbstractFromSessions(firstSessions),
	}
	
	message := fmt.Sprintf("Started planning schedule for %s, session ID: %s. Please introduce these %d morning session options to the user in their preferred language and ask them to choose the ones they're interested in.",
		internalDay, sessionID, len(firstSessions))
	
	response := buildStandardResponse(sessionID, data, message, callReason)

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 2. Choose Session Tool - using new API
func createChooseSessionTool() mcp.Tool {
	return mcp.NewTool(
		"choose_session",
		mcp.WithDescription(sessionIdWarning + "Record user's selected session to their schedule. Use this tool when user explicitly indicates they want to select a certain session. The tool will: 1. Add session to user's schedule 2. Update user's interest profile 3. Automatically recommend next timeslot session options. IMPORTANT: After receiving response, you MUST display ALL available options returned in next_options array. For each session, show the tags field first for categorization, then basic info, create a brief 1-2 sentence summary of the abstract, and include the official COSCUP URL. Group sessions by their tags and present in a clear, organized format in the user's preferred language."),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("sessionCode", 
			mcp.Description("The session code that user selected"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain what the user selected and why you want to record this choice"),
		),
	)
}

func handleChooseSession(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	sessionCode, err := request.RequireString("sessionCode")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionCode is required"), nil
	}

	callReason := request.GetString("callReason", "")

	// Add session to user's schedule
	err = AddSessionToSchedule(sessionID, sessionCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	// Get selected session details
	selectedSession := FindSessionByCode(sessionCode)
	if selectedSession == nil {
		return mcp.NewToolResultError("Error: cannot find details of selected session"), nil
	}

	// Get next recommendations
	recommendations, _ := GetRecommendations(sessionID, 5)

	var nextMessage string
	if len(recommendations) == 0 {
		if IsScheduleComplete(sessionID) {
			nextMessage = "Great! Your schedule planning is complete. Please use mcp_ask to view the full schedule."
		} else {
			nextMessage = "No more sessions available to choose from at this time."
		}
	} else {
		nextMessage = fmt.Sprintf("Selection recorded! Found %d available sessions from different rooms. CRITICAL: You MUST display ALL %d sessions below. For each session, show: 1) Tags (from the tags field) to categorize 2) Basic info (title, speakers, time, room, track) 3) The official COSCUP URL. Note: Abstract field is empty to reduce response size - use get_session_detail tool if user needs complete session information including abstract. Group sessions by their tags and organize them clearly in the user's preferred language.", len(recommendations), len(recommendations))
	}

	data := map[string]any{
		"selected_session": selectedSession,
		"next_options":     removeAbstractFromSessions(recommendations),
		"is_complete":      IsScheduleComplete(sessionID),
	}
	
	response := buildStandardResponse(sessionID, data, nextMessage, callReason)

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 3. Get Options Tool - using new API
func createGetOptionsTool() mcp.Tool {
	return mcp.NewTool(
		"get_options",
		mcp.WithDescription(sessionIdWarning + "Get user's current available session options. Use this tool when you need to recommend next timeslot sessions for the user. The tool will base recommendations on: - User's last selected session end time - Interest profile (learned from previous selections). IMPORTANT: This tool returns ALL available sessions from different rooms. You MUST display every single option returned. For each session, show the tags field first for categorization, then basic info, create a brief 1-2 sentence summary of the abstract, and include the official COSCUP URL. Group sessions by their tags and present in a clear, organized format in the user's preferred language."),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain why you need to get options, e.g., need to recommend next timeslot sessions"),
		),
	)
}

func handleGetOptions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	callReason := request.GetString("callReason", "")

	state := GetUserState(sessionID)
	if state == nil {
		return mcp.NewToolResultError("Error: cannot find specified session"), nil
	}

	recommendations, err := GetRecommendations(sessionID, 5)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	var message string
	if len(recommendations) == 0 {
		message = "No sessions currently available to choose from. May have completed today's planning or no more suitable timeslots available."
	} else {
		message = fmt.Sprintf("Found %d available sessions from different rooms. CRITICAL: You MUST display ALL %d sessions below. For each session, show: 1) Tags (from the tags field) to categorize 2) Basic info (title, speakers, time, room, track) 3) The official COSCUP URL. Note: Abstract field is empty to reduce response size - use get_session_detail tool if user needs complete session information including abstract. Group sessions by their tags, consider user's interests (%v) when organizing them, and present in their preferred language.", len(recommendations), len(recommendations), state.Profile)
	}

	data := map[string]any{
		"options":                removeAbstractFromSessions(recommendations),
		"user_profile":           state.Profile,
		"last_end_time":          state.LastEndTime,
		"current_schedule_count": len(state.Schedule),
	}
	
	response := buildStandardResponse(sessionID, data, message, callReason)

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 4. Get Schedule Tool - using new API
func createGetScheduleTool() mcp.Tool {
	return mcp.NewTool(
		"get_schedule",
		mcp.WithDescription(sessionIdWarning + "Get user's complete planned schedule timeline for a specific day. Use this tool when user wants to view their current planned agenda, check their complete schedule, or review their selected sessions in chronological order. Returns a well-formatted timeline view with session details, time gaps, and schedule statistics."),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Briefly explain why you need to get the schedule, e.g., user wants to view complete timeline"),
		),
	)
}

// 5. Get Next Session Tool - using new API
func createGetNextSessionTool() mcp.Tool {
	return mcp.NewTool(
		"get_next_session",
		mcp.WithDescription(sessionIdWarning + `獲取用戶下一個排定的議程資訊和移動建議。當用戶問：
- "what's next" / "下一場議程" / "我該去哪"  
- "接下來要去哪裡" / "下個session在哪"

工具會自動判斷當前狀態：
- 🎯 議程進行中：顯示剩餘時間，預告下一場
- ⏰ 空檔時間：提供移動建議和時間安排  
- ✅ 剛結束：立即提供下一場地點和最佳路線

回應時要像貼心助手，主動提供移動時間、路線指引和時間規劃建議。
如果用戶還沒規劃行程，請引導使用 start_planning 開始規劃。`),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain why you are calling this tool, e.g., user asks what's next"),
		),
	)
}

// 6. Get Session Detail Tool - using new API
func createGetSessionDetailTool() mcp.Tool {
	return mcp.NewTool(
		"get_session_detail",
		mcp.WithDescription("獲取特定議程的完整詳細資訊，包含完整摘要內容。當你需要瞭解某個議程的詳細描述、難度、語言等完整資訊時使用此工具。這是取得議程 abstract 等完整欄位的唯一方式。"),
		mcp.WithString("sessionCode", 
			mcp.Description("The session code to get details for"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain why you need the session details"),
		),
	)
}

// 7. Finish Planning Tool - using new API
func createFinishPlanningTool() mcp.Tool {
	return mcp.NewTool(
		"finish_planning",
		mcp.WithDescription(sessionIdWarning + "User wants to finish planning and complete their schedule. Use this tool when user explicitly says they want to end planning or when you ask and they confirm they're satisfied with current schedule. This marks their planning as completed and prevents further 'planning_available' status from appearing."),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Explain why user wants to finish planning"),
		),
	)
}

// 8. Help Tool - using new API  
func createHelpTool() mcp.Tool {
	return mcp.NewTool(
		"help",
		mcp.WithDescription("Get user-friendly help about COSCUP planning operations and usage examples. Use this tool when user asks for help, wants to know what they can do, or needs usage guidance. Provides practical examples and operation categories rather than technical tool lists."),
		mcp.WithString("callReason", 
			mcp.Description("Explain why you are calling this tool and what help the user needs"),
		),
	)
}

func handleGetSchedule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	callReason := request.GetString("callReason", "")

	state := GetUserState(sessionID)
	if state == nil {
		return mcp.NewToolResultError("Error: cannot find specified session"), nil
	}

	// Generate timeline format
	timeline := generateTimelineView(state)
	
	data := map[string]any{
		"day":            state.Day,
		"schedule":       state.Schedule,
		"schedule_count": len(state.Schedule),
		"last_end_time":  state.LastEndTime,
		"user_profile":   state.Profile,
		"is_complete":    IsScheduleComplete(sessionID),
		"timeline_view":  timeline,
	}
	
	message := fmt.Sprintf("完整議程時間軸已生成。用戶已選擇 %d 個 session，最後結束時間 %s。請以用戶偏好語言呈現時間軸格式的議程安排。",
		len(state.Schedule), state.LastEndTime)
	
	response := buildStandardResponse(sessionID, data, message, callReason)

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

func handleGetNextSession(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	callReason := request.GetString("callReason", "")

	// Get next session information
	nextInfo, err := GetNextSession(sessionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	// Ensure sessionId is in the next info data
	nextInfo["sessionId"] = sessionID
	
	response := Response{
		Success:    true,
		Data:       nextInfo,
		CallReason: callReason,
		Message:    nextInfo["message"].(string),
	}

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

func handleHelp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	callReason := request.GetString("callReason", "")

	helpContent := `🎯 COSCUP 議程規劃助手使用指南

我可以幫您安排 COSCUP 2025 的議程，支援以下操作：

📅 規劃議程
   • 告訴我您想規劃哪一天（8/9 或 8/10）
   • 從推薦清單中選擇感興趣的議程
   • 系統會自動避免時間衝突並推薦下一時段

🗓️ 管理行程  
   • 查看完整的議程時間軸安排
   • 獲取特定議程的詳細資訊和摘要
   • 隨時結束規劃並確定最終行程

🧭 當天導航
   • 詢問 "what's next" 或 "下一場議程"
   • 獲得即時的移動建議和路線指引
   • 查看剩餘時間和場館移動資訊

💡 使用範例：
   "幫我規劃 8/9 的 COSCUP 行程"
   "我想參加 AI 相關的議程"
   "現在下一場在哪裡？"
   "查看我今天的完整行程"
   "這個議程的詳細內容是什麼？"
   "我覺得規劃夠了，結束規劃"

✨ 特色功能：
   • 智慧推薦：根據您的興趣自動推薦相關議程
   • 路線規劃：計算場館間移動時間和最佳路線
   • 即時狀態：知道現在該做什麼、去哪裡
   • 中英文支援：可以用中文或英文與我互動

隨時說 "help" 或 "幫助" 都可以再次查看此說明！`

	data := map[string]any{
		"help_content": helpContent,
		"available_tools": []string{
			"start_planning", 
			"choose_session", 
			"get_options", 
			"get_schedule", 
			"get_next_session",
			"get_session_detail",
			"finish_planning",
			"help",
		},
	}
	
	message := "COSCUP 議程規劃助手使用指南已提供。請以用戶偏好語言友善地介紹如何使用這個規劃助手，重點說明可以進行的操作和實用範例。"
	
	// For help, we don't need a specific sessionID
	response := Response{
		Success:    true,
		Data:       data,
		CallReason: callReason,
		Message:    message,
	}

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

func handleGetSessionDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionCode, err := request.RequireString("sessionCode")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionCode is required"), nil
	}

	callReason := request.GetString("callReason", "")

	// Find the session by code
	session := FindSessionByCode(sessionCode)
	if session == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: session %s not found", sessionCode)), nil
	}

	data := map[string]any{
		"session": *session, // Return the complete session with all fields
	}
	
	message := fmt.Sprintf("議程 %s 的完整詳細資訊已提供。這包含完整的摘要內容、難度等級、授課語言等所有資訊。請以用戶偏好語言呈現完整的議程詳情。", sessionCode)
	
	// For session detail, we don't have a specific sessionID, so pass empty string
	response := Response{
		Success:    true,
		Data:       data,
		CallReason: callReason,
		Message:    message,
	}

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

func handleFinishPlanning(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	callReason := request.GetString("callReason", "")

	// Check if session exists
	state := GetUserState(sessionID)
	if state == nil {
		return mcp.NewToolResultError("Error: cannot find specified session"), nil
	}

	// Mark planning as completed
	err = FinishPlanning(sessionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	data := map[string]any{
		"day":            state.Day,
		"schedule":       state.Schedule,
		"schedule_count": len(state.Schedule),
		"last_end_time":  state.LastEndTime,
		"is_completed":   true,
	}
	
	message := fmt.Sprintf("🎉 規劃完成！您已成功規劃了 %s 的議程，共選擇 %d 個 session，最後結束時間 %s。您的 COSCUP 2025 行程已確定完成。可以開始期待精彩的議程內容！",
		state.Day, len(state.Schedule), state.LastEndTime)
	
	response := buildStandardResponse(sessionID, data, message, callReason)

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// GetToolHandlers returns a map of tool names to their handlers using new API
func GetToolHandlers() map[string]server.ToolHandlerFunc {
	return map[string]server.ToolHandlerFunc{
		"start_planning":     handleStartPlanning,
		"choose_session":     handleChooseSession,
		"get_options":        handleGetOptions,
		"get_schedule":       handleGetSchedule,
		"get_next_session":   handleGetNextSession,
		"get_session_detail": handleGetSessionDetail,
		"finish_planning":    handleFinishPlanning,
		"help":               handleHelp,
	}
}