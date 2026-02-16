// Package ui provides terminal UI components including prompts,
// spinners, colored output, and CI environment detection.
package ui

import "os"

// ciEnvVars lists environment variables that indicate a CI environment.
var ciEnvVars = []string{
	"CI",
	"CONTINUOUS_INTEGRATION",
	"BUILD_NUMBER",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"CIRCLECI",
	"TRAVIS",
	"JENKINS_URL",
	"BITBUCKET_PIPELINE",
	"CODEBUILD_BUILD_ID",
	"TF_BUILD",
}

// IsCI checks if the program is running in a CI environment
// by looking for known CI environment variables.
func IsCI() bool {
	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// ciProviders maps environment variables to CI provider names.
var ciProviders = map[string]string{
	"GITHUB_ACTIONS":     "github-actions",
	"GITLAB_CI":          "gitlab-ci",
	"CIRCLECI":           "circle-ci",
	"TRAVIS":             "travis-ci",
	"JENKINS_URL":        "jenkins",
	"BITBUCKET_PIPELINE": "bitbucket",
	"CODEBUILD_BUILD_ID": "aws-codebuild",
	"TF_BUILD":           "azure-devops",
}

// DetectCIProvider returns the name of the detected CI provider.
// Returns an empty string if no CI provider is detected.
func DetectCIProvider() string {
	for envVar, provider := range ciProviders {
		if os.Getenv(envVar) != "" {
			return provider
		}
	}
	return ""
}
