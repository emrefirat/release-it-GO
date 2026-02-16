package ui

import (
	"os"
	"testing"
)

func TestIsCI_NoEnvVars(t *testing.T) {
	// Clear all CI env vars
	for _, env := range ciEnvVars {
		os.Unsetenv(env)
	}

	if IsCI() {
		t.Error("expected IsCI() to return false when no CI env vars are set")
	}
}

func TestIsCI_WithCIVar(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
	}{
		{"CI", "CI"},
		{"GITHUB_ACTIONS", "GITHUB_ACTIONS"},
		{"GITLAB_CI", "GITLAB_CI"},
		{"CIRCLECI", "CIRCLECI"},
		{"TRAVIS", "TRAVIS"},
		{"JENKINS_URL", "JENKINS_URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all CI env vars first
			for _, env := range ciEnvVars {
				os.Unsetenv(env)
			}

			os.Setenv(tt.envVar, "true")
			defer os.Unsetenv(tt.envVar)

			if !IsCI() {
				t.Errorf("expected IsCI() to return true when %s is set", tt.envVar)
			}
		})
	}
}

func TestDetectCIProvider_NoProvider(t *testing.T) {
	for env := range ciProviders {
		os.Unsetenv(env)
	}

	provider := DetectCIProvider()
	if provider != "" {
		t.Errorf("expected empty provider, got %q", provider)
	}
}

func TestDetectCIProvider_WithProvider(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{"GitHub Actions", "GITHUB_ACTIONS", "github-actions"},
		{"GitLab CI", "GITLAB_CI", "gitlab-ci"},
		{"CircleCI", "CIRCLECI", "circle-ci"},
		{"Travis CI", "TRAVIS", "travis-ci"},
		{"Jenkins", "JENKINS_URL", "jenkins"},
		{"Bitbucket", "BITBUCKET_PIPELINE", "bitbucket"},
		{"AWS CodeBuild", "CODEBUILD_BUILD_ID", "aws-codebuild"},
		{"Azure DevOps", "TF_BUILD", "azure-devops"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for env := range ciProviders {
				os.Unsetenv(env)
			}

			os.Setenv(tt.envVar, "true")
			defer os.Unsetenv(tt.envVar)

			provider := DetectCIProvider()
			if provider != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, provider)
			}
		})
	}
}
