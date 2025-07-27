package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"mcp-coscup/mcp"
)

func main() {
	fmt.Println("=== COSCUP 2025 Aug.9 議程驗證 ===")
	
	// 載入資料
	err := mcp.LoadCOSCUPData()
	if err != nil {
		log.Fatalf("載入資料失敗: %v", err)
	}
	
	// 直接從 sessionsByDay 取得所有 Aug.9 的議程
	day := "Aug.9"
	sessions := getAllSessionsDirectly(day)
	
	// 按房間分組
	roomSessions := make(map[string][]mcp.Session)
	for _, session := range sessions {
		roomSessions[session.Room] = append(roomSessions[session.Room], session)
	}
	
	// 取得所有房間並排序
	rooms := make([]string, 0, len(roomSessions))
	for room := range roomSessions {
		rooms = append(rooms, room)
	}
	sort.Strings(rooms)
	
	fmt.Printf("\n找到 %d 個房間，共 %d 個議程\n\n", len(rooms), len(sessions))
	
	// 印出每個房間的議程
	for _, room := range rooms {
		sessions := roomSessions[room]
		
		// 按時間排序
		sort.Slice(sessions, func(i, j int) bool {
			return timeToMinutes(sessions[i].Start) < timeToMinutes(sessions[j].Start)
		})
		
		fmt.Printf("=== %s ===\n", room)
		for _, session := range sessions {
			// 格式：時間 | 議程代碼 | 標題 | 講者 | 語言/難度
			speakers := strings.Join(session.Speakers, ", ")
			if speakers == "" {
				speakers = "未知"
			}
			fmt.Printf("%s-%s | %-7s | %s\n              講者: %s | %s/%s\n", 
				session.Start, session.End, session.Code, session.Title,
				speakers, session.Language, session.Difficulty)
		}
		fmt.Println()
	}
}

// getAllSessionsDirectly 嘗試更全面地取得所有房間的議程
func getAllSessionsDirectly(day string) []mcp.Session {
	// 根據實際 embedded data 中的房間名稱搜尋
	allPossibleRooms := []string{
		// 視聽館
		"AU", 
		
		// 綜合研究大樓 (兩種格式)
		"RB-101", "RB-102", "RB-105", 
		"RB101", "RB102", "RB105", 
		
		// 研揚大樓 (基於實際發現的房間)
		"TR209", "TR210", "TR211", "TR212", "TR213", "TR214",
		"TR310-2", "TR311", "TR313", 
		"TR409-2", "TR410", "TR411", "TR412-1", "TR412-2",
		"TR509", "TR510", "TR511", "TR512", "TR513", "TR514", "TR515",
		
		// 走廊和特殊場地
		"Hallway outside TR309", "Hallway outside TR409",
		
		// 其他可能的格式
		"Hacking Corner", "BoF", "Lightning Talk",
	}
	
	var allSessions []mcp.Session
	foundRooms := make(map[string]bool)
	
	// 對每個可能的房間嘗試取得議程
	for _, room := range allPossibleRooms {
		sessions := mcp.FindRoomSessions(day, room)
		if len(sessions) > 0 {
			allSessions = append(allSessions, sessions...)
			foundRooms[room] = true
		}
	}
	
	// 印出找到的房間
	fmt.Printf("找到的房間 (%d個): ", len(foundRooms))
	roomList := make([]string, 0, len(foundRooms))
	for room := range foundRooms {
		roomList = append(roomList, room)
	}
	sort.Strings(roomList)
	for _, room := range roomList {
		fmt.Printf("%s ", room)
	}
	fmt.Println()
	
	return allSessions
}

// timeToMinutes 將時間字串轉換為分鐘數
func timeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}
	
	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	
	if err1 != nil || err2 != nil {
		return 0
	}
	
	return hours*60 + minutes
}