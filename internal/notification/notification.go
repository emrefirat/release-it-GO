// Package notification provides webhook notification support for Slack and Microsoft Teams.
// Notifications are sent after a successful release to inform the team.
package notification

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

const defaultTimeout = 30

// RichNotificationContext holds release context data for rich notifications.
// This enables Teams MessageCard and Slack Block Kit payloads with structured data.
type RichNotificationContext struct {
	Version       string
	LatestVersion string
	TagName       string
	Changelog     string
	ReleaseURL    string
	BranchName    string
	RepoHost      string
	RepoOwner     string
	RepoName      string
	CommitCount   int
	Contributors  []string
	ThemeColor    string // Hex color for Teams card (default: "0076D7")
	ImageURL      string // Activity image URL for Teams card
}

// Client sends webhook notifications to configured endpoints.
type Client struct {
	webhooks    []config.WebhookConfig
	vars        map[string]string
	richContext *RichNotificationContext
	logger      *applog.Logger
	dryRun      bool
	httpClient  *http.Client
}

// NewClient creates a new notification client.
func NewClient(webhooks []config.WebhookConfig, vars map[string]string, logger *applog.Logger, dryRun bool) *Client {
	return &Client{
		webhooks:   webhooks,
		vars:       vars,
		logger:     logger,
		dryRun:     dryRun,
		httpClient: &http.Client{},
	}
}

// SetRichContext sets the rich notification context for structured payloads.
func (c *Client) SetRichContext(ctx *RichNotificationContext) {
	c.richContext = ctx
}

// SendAll sends notifications to all configured webhooks.
// Errors are collected but do not stop processing; a combined error is returned.
func (c *Client) SendAll() error {
	var errs []string

	for _, wh := range c.webhooks {
		if err := c.sendOne(wh); err != nil {
			errs = append(errs, fmt.Sprintf("%s(%s): %v", wh.Type, wh.URLRef, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// sendOne sends a notification to a single webhook endpoint.
func (c *Client) sendOne(wh config.WebhookConfig) error {
	url, err := resolveURL(wh.URLRef)
	if err != nil {
		return err
	}

	message := c.renderMessage(wh)

	var payload []byte
	switch strings.ToLower(wh.Type) {
	case "slack":
		payload, err = buildSlackPayload(message)
	case "teams":
		// Use rich MessageCard if context is available and no custom template override
		if c.richContext != nil && wh.MessageTemplate == "" {
			payload, err = buildTeamsRichPayload(c.richContext)
		} else {
			payload, err = buildTeamsPayload(message)
		}
	default:
		return fmt.Errorf("unsupported webhook type: %q", wh.Type)
	}
	if err != nil {
		return fmt.Errorf("building payload: %w", err)
	}

	if c.dryRun {
		c.logger.DryRun("Would send %s notification to %s", wh.Type, wh.URLRef)
		return nil
	}

	timeout := wh.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	c.httpClient.Timeout = time.Duration(timeout) * time.Second

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("HTTP POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d from %s webhook", resp.StatusCode, wh.Type)
	}

	c.logger.Verbose("Sent %s notification successfully", wh.Type)
	return nil
}

// renderMessage applies template variables to the message template,
// falling back to a platform-specific default if none is configured.
func (c *Client) renderMessage(wh config.WebhookConfig) string {
	tmpl := wh.MessageTemplate
	if tmpl == "" {
		tmpl = defaultTemplate(wh.Type)
	}

	result := tmpl
	for k, v := range c.vars {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}

// defaultTemplate returns the default message template for a webhook type.
func defaultTemplate(webhookType string) string {
	switch strings.ToLower(webhookType) {
	case "slack":
		return slackDefaultTemplate
	case "teams":
		return teamsDefaultTemplate
	default:
		return "🚀 v${version} released!\n${releaseUrl}"
	}
}

// resolveURL reads the webhook URL from the environment variable named by urlRef.
func resolveURL(urlRef string) (string, error) {
	if urlRef == "" {
		return "", fmt.Errorf("urlRef is empty")
	}
	url := os.Getenv(urlRef)
	if url == "" {
		return "", fmt.Errorf("environment variable %s is not set", urlRef)
	}
	return url, nil
}
