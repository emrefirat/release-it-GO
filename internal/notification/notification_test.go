package notification

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

func newTestLogger() *applog.Logger {
	return applog.NewLoggerWithWriter(0, false, io.Discard)
}

func newDryRunLogger() *applog.Logger {
	return applog.NewLoggerWithWriter(0, true, io.Discard)
}

func TestSendAll_SlackSuccess(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_SLACK_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "TEST_SLACK_URL"},
	}
	vars := map[string]string{
		"version":         "1.2.3",
		"releaseUrl":      "https://github.com/example/repo/releases/v1.2.3",
		"repo.repository": "my-repo",
	}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Slack payload format
	var payload slackPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload.Text == "" {
		t.Error("expected non-empty text in Slack payload")
	}
}

func TestSendAll_TeamsSuccess(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_TEAMS_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "teams", URLRef: "TEST_TEAMS_URL"},
	}
	vars := map[string]string{
		"version":         "2.0.0",
		"releaseUrl":      "https://github.com/example/repo/releases/v2.0.0",
		"repo.repository": "my-repo",
	}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Teams payload format
	var payload teamsMessageCard
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload.Type != "MessageCard" {
		t.Errorf("expected @type MessageCard, got %q", payload.Type)
	}
	if len(payload.Sections) == 0 {
		t.Error("expected at least one section in Teams payload")
	}
}

func TestSendAll_MissingURL(t *testing.T) {
	t.Setenv("MISSING_URL_REF", "")

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "NONEXISTENT_WEBHOOK_URL"},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestSendAll_DryRun(t *testing.T) {
	// Server should NOT receive any requests in dry-run mode
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not send HTTP request in dry-run mode")
	}))
	defer server.Close()

	t.Setenv("TEST_DRY_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "TEST_DRY_URL"},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newDryRunLogger(), true)
	err := client.SendAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendAll_InvalidType(t *testing.T) {
	t.Setenv("TEST_INVALID_TYPE_URL", "https://example.com")

	webhooks := []config.WebhookConfig{
		{Type: "discord", URLRef: "TEST_INVALID_TYPE_URL"},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestSendAll_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv("TEST_ERROR_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "TEST_ERROR_URL"},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestSendAll_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_TIMEOUT_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "TEST_TIMEOUT_URL", Timeout: 1},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestSendAll_CustomTemplate(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_TEMPLATE_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{
			Type:            "slack",
			URLRef:          "TEST_TEMPLATE_URL",
			MessageTemplate: "Release ${version} is out! Check ${releaseUrl}",
		},
	}
	vars := map[string]string{
		"version":    "3.0.0",
		"releaseUrl": "https://example.com/releases/3.0.0",
	}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload slackPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	expected := "Release 3.0.0 is out! Check https://example.com/releases/3.0.0"
	if payload.Text != expected {
		t.Errorf("expected %q, got %q", expected, payload.Text)
	}
}

func TestSendAll_MultipleWebhooks(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_MULTI_SLACK", server.URL)
	t.Setenv("TEST_MULTI_TEAMS", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "slack", URLRef: "TEST_MULTI_SLACK"},
		{Type: "teams", URLRef: "TEST_MULTI_TEAMS"},
	}
	vars := map[string]string{"version": "1.0.0"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	err := client.SendAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 webhook calls, got %d", callCount)
	}
}

func TestResolveURL_Empty(t *testing.T) {
	_, err := resolveURL("")
	if err == nil {
		t.Fatal("expected error for empty urlRef")
	}
}

func TestResolveURL_NotSet(t *testing.T) {
	_, err := resolveURL("THIS_ENV_VAR_SHOULD_NOT_EXIST_XYZ123")
	if err == nil {
		t.Fatal("expected error for unset env var")
	}
}

func TestBuildSlackPayload(t *testing.T) {
	data, err := buildSlackPayload("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var p slackPayload
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if p.Text != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", p.Text)
	}
}

func TestBuildTeamsPayload(t *testing.T) {
	data, err := buildTeamsPayload("hello teams")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var p teamsMessageCard
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if p.Type != "MessageCard" {
		t.Errorf("expected @type MessageCard, got %q", p.Type)
	}
	if len(p.Sections) == 0 || p.Sections[0].Text != "hello teams" {
		t.Errorf("expected section text %q, got %+v", "hello teams", p.Sections)
	}
	if p.Summary != "hello teams" {
		t.Errorf("expected summary %q, got %q", "hello teams", p.Summary)
	}
}

func TestBuildTeamsRichPayload_FullContext(t *testing.T) {
	ctx := &RichNotificationContext{
		Version:       "1.2.0",
		LatestVersion: "1.1.0",
		TagName:       "v1.2.0",
		Changelog:     "### Features\n- Add dashboard\n### Bug Fixes\n- Fix login",
		RepoHost:      "gitlab.com",
		RepoOwner:     "myteam",
		RepoName:      "myapp",
		CommitCount:   5,
		Contributors:  []string{"Alice", "Bob"},
		ThemeColor:    "FF5500",
		ImageURL:      "https://example.com/logo.png",
	}

	data, err := buildTeamsRichPayload(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var card teamsMessageCard
	if err := json.Unmarshal(data, &card); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if card.Type != "MessageCard" {
		t.Errorf("@type = %q, want MessageCard", card.Type)
	}
	if card.ThemeColor != "FF5500" {
		t.Errorf("themeColor = %q, want FF5500", card.ThemeColor)
	}
	if len(card.Sections) < 1 {
		t.Fatal("expected at least 1 section")
	}

	main := card.Sections[0]
	if main.ActivityImage != "https://example.com/logo.png" {
		t.Errorf("activityImage = %q", main.ActivityImage)
	}
	if !main.Markdown {
		t.Error("expected markdown=true")
	}

	// Check facts
	factMap := make(map[string]string)
	for _, f := range main.Facts {
		factMap[f.Name] = f.Value
	}
	if factMap["Version"] != "1.2.0" {
		t.Errorf("Version fact = %q", factMap["Version"])
	}
	if factMap["Last Release"] != "1.1.0" {
		t.Errorf("Last Release fact = %q", factMap["Last Release"])
	}
	if factMap["Commits"] != "5" {
		t.Errorf("Commits fact = %q", factMap["Commits"])
	}
	if factMap["Contributors"] != "Alice, Bob" {
		t.Errorf("Contributors fact = %q", factMap["Contributors"])
	}

	// Changelog section should be present
	if len(card.Sections) < 3 {
		t.Fatalf("expected 3+ sections (main + separator + changelog), got %d", len(card.Sections))
	}
	if card.Sections[2].Text == "" {
		t.Error("expected changelog in section text")
	}
}

func TestBuildTeamsRichPayload_MinimalContext(t *testing.T) {
	ctx := &RichNotificationContext{
		Version:  "0.1.0",
		RepoName: "test-app",
	}

	data, err := buildTeamsRichPayload(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var card teamsMessageCard
	if err := json.Unmarshal(data, &card); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Default theme color
	if card.ThemeColor != "0076D7" {
		t.Errorf("themeColor = %q, want default 0076D7", card.ThemeColor)
	}

	main := card.Sections[0]
	// Only Version fact (no Last Release, no Commits, no Contributors)
	if len(main.Facts) != 1 {
		t.Errorf("expected 1 fact (Version only), got %d", len(main.Facts))
	}
	if main.Facts[0].Value != "0.1.0" {
		t.Errorf("Version fact = %q", main.Facts[0].Value)
	}

	// No changelog → only 1 section (no separator + changelog)
	if len(card.Sections) != 1 {
		t.Errorf("expected 1 section (no changelog), got %d", len(card.Sections))
	}
}

func TestSendAll_TeamsRichPayload(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TEST_TEAMS_RICH_URL", server.URL)

	webhooks := []config.WebhookConfig{
		{Type: "teams", URLRef: "TEST_TEAMS_RICH_URL"},
	}
	vars := map[string]string{"version": "2.0.0", "repo.repository": "myapp"}

	client := NewClient(webhooks, vars, newTestLogger(), false)
	client.SetRichContext(&RichNotificationContext{
		Version:     "2.0.0",
		RepoName:    "myapp",
		RepoHost:    "github.com",
		RepoOwner:   "org",
		CommitCount: 3,
	})

	if err := client.SendAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var card teamsMessageCard
	if err := json.Unmarshal(receivedBody, &card); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if card.Type != "MessageCard" {
		t.Errorf("@type = %q", card.Type)
	}
	// Should have facts section
	if len(card.Sections) == 0 || len(card.Sections[0].Facts) == 0 {
		t.Error("expected facts in rich Teams payload")
	}
}
