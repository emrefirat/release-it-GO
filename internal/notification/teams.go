package notification

import (
	"encoding/json"
	"strconv"
	"strings"
)

const teamsDefaultTemplate = "🚀 ${repo.repository} v${version} released!\n${releaseUrl}"

// Teams MessageCard structures
// https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using

type teamsMessageCard struct {
	Type       string         `json:"@type"`
	Context    string         `json:"@context"`
	ThemeColor string         `json:"themeColor,omitempty"`
	Summary    string         `json:"summary"`
	Sections   []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle    string      `json:"activityTitle,omitempty"`
	ActivitySubtitle string      `json:"activitySubtitle,omitempty"`
	ActivityImage    string      `json:"activityImage,omitempty"`
	Facts            []teamsFact `json:"facts,omitempty"`
	Text             string      `json:"text,omitempty"`
	Markdown         bool        `json:"markdown,omitempty"`
}

type teamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// buildTeamsPayload creates a simple Teams MessageCard JSON payload.
// Used when no rich context is available (backward compatible).
func buildTeamsPayload(message string) ([]byte, error) {
	card := teamsMessageCard{
		Type:    "MessageCard",
		Context: "http://schema.org/extensions",
		Summary: message,
		Sections: []teamsSection{
			{Text: message, Markdown: true},
		},
	}
	return json.Marshal(card)
}

// buildTeamsRichPayload creates a rich Teams MessageCard with facts, changelog sections,
// and contributor information — similar to release-it-teams-notifier npm plugin.
func buildTeamsRichPayload(ctx *RichNotificationContext) ([]byte, error) {
	repoURL := ""
	if ctx.RepoHost != "" && ctx.RepoOwner != "" && ctx.RepoName != "" {
		repoURL = "https://" + ctx.RepoHost + "/" + ctx.RepoOwner + "/" + ctx.RepoName
	}

	// Build facts
	facts := []teamsFact{
		{Name: "Version", Value: ctx.Version},
	}
	if ctx.LatestVersion != "" {
		facts = append(facts, teamsFact{Name: "Last Release", Value: ctx.LatestVersion})
	}
	if ctx.CommitCount > 0 {
		facts = append(facts, teamsFact{Name: "Commits", Value: strconv.Itoa(ctx.CommitCount)})
	}
	if len(ctx.Contributors) > 0 {
		facts = append(facts, teamsFact{Name: "Contributors", Value: strings.Join(ctx.Contributors, ", ")})
	}

	themeColor := ctx.ThemeColor
	if themeColor == "" {
		themeColor = "0076D7"
	}

	imageURL := ctx.ImageURL
	if imageURL == "" {
		imageURL = "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e1/GitLab_logo.svg/64px-GitLab_logo.svg.png"
	}

	title := "🚀🚀 A new version for " + ctx.RepoName + " has been released 🚀🚀"

	// Main section with facts
	sections := []teamsSection{
		{
			ActivityTitle:    title,
			ActivitySubtitle: repoURL,
			ActivityImage:    imageURL,
			Facts:            facts,
			Markdown:         true,
		},
	}

	// Changelog section
	if ctx.Changelog != "" {
		sections = append(sections, teamsSection{Text: "---"})
		sections = append(sections, teamsSection{Text: ctx.Changelog, Markdown: true})
	}

	card := teamsMessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: themeColor,
		Summary:    title,
		Sections:   sections,
	}

	return json.Marshal(card)
}
