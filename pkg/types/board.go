package types

import (
	"encoding/json"
)

type BoardD struct {
	Elements    []json.RawMessage `json:"elements"`
	AppState    json.RawMessage   `json:"appState"`
	Files       json.RawMessage   `json:"files"`
	WorkspaceID string            `json:"workspaceId"`
}
