package types

type AppState struct {
	UserWorkspaceId string   `json:"userWorkspaceId"`
	MissedCalls     []string `json:"missedCalls"`
	UnreadChannels  []string `json:"unreadChannels"`
}

type AppStateUpdateRequest struct {
	Action string `json:"action"`
	Value  string `json:"value"`
}
