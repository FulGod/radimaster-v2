package ws

// MessageType defines WebSocket message types for medical viewer synchronization.
type MessageType string

const (
	// Viewer sync — doctor controls, students follow
	SliceChanged    MessageType = "slice:changed"    // payload: {phase_id, slice_index}
	PhaseChanged    MessageType = "phase:changed"    // payload: {phase_id}
	ZoomChanged     MessageType = "zoom:changed"     // payload: {level, offset_x, offset_y}
	ContrastChanged MessageType = "contrast:changed" // payload: {brightness, contrast}

	// Cursor sync — see where doctor is pointing
	CursorMoved   MessageType = "cursor:moved"   // payload: {x, y, phase_id}
	CursorClicked MessageType = "cursor:clicked" // payload: {x, y, phase_id}

	// Presence
	UserJoined     MessageType = "user:joined"
	UserLeft       MessageType = "user:left"
	PresenceUpdate MessageType = "presence:update"
)

// Message is the WebSocket message envelope.
type Message struct {
	Type     MessageType `json:"type"`
	UserID   uint        `json:"user_id,omitempty"`
	UserName string      `json:"user_name,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}
