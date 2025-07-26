package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateMCPTools creates and returns all MCP tools using new helper functions
func CreateMCPTools() map[string]mcp.Tool {
	return map[string]mcp.Tool{
		"start_planning": createStartPlanningTool(),
		"choose_session": createChooseSessionTool(),
		"get_options":    createGetOptionsTool(),
		"mcp_ask":        createMCPAskTool(),
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

	response := Response{
		Success: true,
		Data: map[string]any{
			"sessionId": sessionID,
			"day":        internalDay,
			"options":    firstSessions,
		},
		CallReason: callReason,
		Message: fmt.Sprintf("Started planning schedule for %s, session ID: %s. Please introduce these %d morning session options to the user in their preferred language and ask them to choose the ones they're interested in.",
			internalDay, sessionID, len(firstSessions)),
	}

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 2. Choose Session Tool - using new API
func createChooseSessionTool() mcp.Tool {
	return mcp.NewTool(
		"choose_session",
		mcp.WithDescription("Record user's selected session to their schedule. Use this tool when user explicitly indicates they want to select a certain session. The tool will: 1. Add session to user's schedule 2. Update user's interest profile 3. Automatically recommend next timeslot session options. IMPORTANT: After receiving response, you MUST display ALL available options returned in next_options array. For each session, show the tags field first for categorization, then basic info, create a brief 1-2 sentence summary of the abstract, and include the official COSCUP URL. Group sessions by their tags and present in a clear, organized format in the user's preferred language."),
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
		nextMessage = fmt.Sprintf("Selection recorded! Found %d available sessions from different rooms. CRITICAL: You MUST display ALL %d sessions below. For each session, show: 1) Tags (from the tags field) to categorize 2) Basic info (title, speakers, time, room, track) 3) A brief 1-2 sentence summary instead of the full abstract 4) The official COSCUP URL for detailed information. Group sessions by their tags and organize them clearly in the user's preferred language.", len(recommendations), len(recommendations))
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

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 3. Get Options Tool - using new API
func createGetOptionsTool() mcp.Tool {
	return mcp.NewTool(
		"get_options",
		mcp.WithDescription("Get user's current available session options. Use this tool when you need to recommend next timeslot sessions for the user. The tool will base recommendations on: - User's last selected session end time - Interest profile (learned from previous selections). IMPORTANT: This tool returns ALL available sessions from different rooms. You MUST display every single option returned. For each session, show the tags field first for categorization, then basic info, create a brief 1-2 sentence summary of the abstract, and include the official COSCUP URL. Group sessions by their tags and present in a clear, organized format in the user's preferred language."),
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
		message = fmt.Sprintf("Found %d available sessions from different rooms. CRITICAL: You MUST display ALL %d sessions below. For each session, show: 1) Tags (from the tags field) to categorize 2) Basic info (title, speakers, time, room, track) 3) A brief 1-2 sentence summary instead of the full abstract 4) The official COSCUP URL for detailed information. Group sessions by their tags, consider user's interests (%v) when organizing them, and present in their preferred language.", len(recommendations), len(recommendations), state.Profile)
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

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// 4. MCP Ask Tool - using new API
func createMCPAskTool() mcp.Tool {
	return mcp.NewTool(
		"mcp_ask",
		mcp.WithDescription("Self-check and status confirmation tool. Use when: - You're uncertain about the result of the previous operation - Need to view user's complete schedule - Want to confirm if current status is correct - User asks about their current planning status. This tool returns complete user state to help you reassess and decide next actions. Respond to users in their preferred language."),
		mcp.WithString("sessionId", 
			mcp.Description("User's session ID"),
		),
		mcp.WithString("question", 
			mcp.Description("What do you want to confirm? Describe your confusion or points that need verification"),
		),
		mcp.WithString("callReason", 
			mcp.Description("Briefly explain why confirmation is needed, e.g., user wants to view complete schedule"),
		),
	)
}

func handleMCPAsk(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("sessionId")
	if err != nil {
		return mcp.NewToolResultError("Error: sessionId is required"), nil
	}

	question := request.GetString("question", "")
	callReason := request.GetString("callReason", "")

	state := GetUserState(sessionID)
	if state == nil {
		return mcp.NewToolResultError("Error: cannot find specified session"), nil
	}

	// Get available options for context
	nextOptions, _ := GetRecommendations(sessionID, 5)

	response := Response{
		Success: true,
		Data: map[string]any{
			"current_state": map[string]any{
				"sessionId":     state.SessionID,
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
		Message: fmt.Sprintf("✅ Current status confirmation completed. User has selected %d sessions, last end time %s. Answer user's questions based on this information in their preferred language.",
			len(state.Schedule), state.LastEndTime),
	}

	return mcp.NewToolResultText(fmt.Sprintf("%+v", response)), nil
}

// GetToolHandlers returns a map of tool names to their handlers using new API
func GetToolHandlers() map[string]server.ToolHandlerFunc {
	return map[string]server.ToolHandlerFunc{
		"start_planning": handleStartPlanning,
		"choose_session": handleChooseSession,
		"get_options":    handleGetOptions,
		"mcp_ask":        handleMCPAsk,
	}
}