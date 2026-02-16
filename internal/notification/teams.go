package notification

import "encoding/json"

const teamsDefaultTemplate = "🚀 ${repo.repository} v${version} released!\n${releaseUrl}"

// teamsPayload is the JSON structure for a Microsoft Teams MessageCard webhook.
type teamsPayload struct {
	Type    string `json:"@type"`
	Summary string `json:"summary"`
	Text    string `json:"text"`
}

// buildTeamsPayload creates a Teams MessageCard JSON payload from a message string.
func buildTeamsPayload(message string) ([]byte, error) {
	payload := teamsPayload{
		Type:    "MessageCard",
		Summary: message,
		Text:    message,
	}
	return json.Marshal(payload)
}
