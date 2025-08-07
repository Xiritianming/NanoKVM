package ws

type Stream struct {
	Type  string `json:"type"`
	State int    `json:"state"`
}

type ResolutionChange struct {
	Type   string `json:"type"`
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}
