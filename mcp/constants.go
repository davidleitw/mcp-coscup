package mcp

// COSCUP 2025 活動相關常數
const (
	COSCUPYear  = 2025
	COSCUPMonth = 8
	COSCUPDay1  = 9
	COSCUPDay2  = 10
	DemoHour    = 10
	DemoMinute  = 23
)

// 系統配置常數
const (
	DefaultNumShards     = 16
	SessionCleanupHours  = 24
	LongSessionMinutes   = 240 // 4 hours
)

// 場地移動時間常數 (分鐘)
const (
	SameBuildingWalkTime = 1
	AUToRBWalkTime      = 2
	AUToTRWalkTime      = 4
	RBToAUWalkTime      = 2
	RBToRBWalkTime      = 1
	RBToTRWalkTime      = 3
	TRToAUWalkTime      = 4
	TRToRBWalkTime      = 3
	TRInternalWalkTime  = 2
	UnknownWalkTime     = 5 // Default for unknown routes
)

// 字串常數
const (
	DayAug9             = "Aug9"
	DayAug10            = "Aug10"
	DayFormatAug9       = "Aug.9"
	DayFormatAug10      = "Aug.10"
	DifficultyBeginner  = "入門"
	StatusOutsideCOSCUP = "OutsideCOSCUP"
)

// 建築物代碼
const (
	BuildingAU = "AU"
	BuildingRB = "RB"
	BuildingTR = "TR"
)