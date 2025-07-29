package mcp

import "errors"

// Standard error definitions
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidDay          = errors.New("invalid day format")
	ErrInvalidSessionCode  = errors.New("invalid session code")
	ErrSessionIDRequired   = errors.New("sessionId is required")
	ErrSessionCodeRequired = errors.New("sessionCode is required")
	ErrRoomRequired        = errors.New("room is required")
	ErrCannotFindSession   = errors.New("cannot find specified session")
)
