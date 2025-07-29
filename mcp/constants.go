package mcp

// COSCUP 2025 event constants
const (
	COSCUPYear  = 2025
	COSCUPMonth = 8
	COSCUPDay1  = 9
	COSCUPDay2  = 10
)

// System configuration constants
const (
	DefaultNumShards    = 16
	SessionCleanupHours = 24
	LongSessionMinutes  = 240 // 4 hours
)

// Venue walking time constants (minutes)
const (
	SameBuildingWalkTime = 1
	AUToRBWalkTime       = 2
	AUToTRWalkTime       = 4
	RBToAUWalkTime       = 2
	RBToRBWalkTime       = 1
	RBToTRWalkTime       = 3
	TRToAUWalkTime       = 4
	TRToRBWalkTime       = 3
	TRInternalWalkTime   = 2
	UnknownWalkTime      = 5 // Default for unknown routes
)

// String constants
const (
	DayAug9             = "Aug9"
	DayAug10            = "Aug10"
	DayFormatAug9       = "Aug.9"
	DayFormatAug10      = "Aug.10"
	DifficultyBeginner  = "入門"
	StatusOutsideCOSCUP = "OutsideCOSCUP"
)

// Building codes
const (
	BuildingAU = "AU"
	BuildingRB = "RB"
	BuildingTR = "TR"
)
