package notification

import "encoding/json"

const slackDefaultTemplate = "🚀 *${repo.repository}* v${version} released!\n${releaseUrl}"

// slackPayload is the JSON structure for a Slack incoming webhook message.
type slackPayload struct {
	Text string `json:"text"`
}

// buildSlackPayload creates a Slack webhook JSON payload from a message string.
func buildSlackPayload(message string) ([]byte, error) {
	payload := slackPayload{Text: message}
	return json.Marshal(payload)
}
