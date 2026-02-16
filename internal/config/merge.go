package config

// MergeFlags applies CLI flag overrides to the config.
// Only non-zero values are applied (zero values are treated as "not set").
type FlagOverrides struct {
	DryRun       *bool
	CI           *bool
	Verbose      *int
	Increment    *string
	PreReleaseID *string
	NoCommit     *bool
	NoTag        *bool
	NoPush       *bool
}

// ApplyFlags merges CLI flag overrides into the config.
func ApplyFlags(cfg *Config, flags FlagOverrides) {
	if flags.DryRun != nil {
		cfg.DryRun = *flags.DryRun
	}
	if flags.CI != nil {
		cfg.CI = *flags.CI
	}
	if flags.Verbose != nil {
		cfg.Verbose = *flags.Verbose
	}
	if flags.Increment != nil && *flags.Increment != "" {
		cfg.Increment = *flags.Increment
	}
	if flags.PreReleaseID != nil && *flags.PreReleaseID != "" {
		cfg.PreReleaseID = *flags.PreReleaseID
	}
	if flags.NoCommit != nil && *flags.NoCommit {
		cfg.Git.Commit = false
	}
	if flags.NoTag != nil && *flags.NoTag {
		cfg.Git.Tag = false
	}
	if flags.NoPush != nil && *flags.NoPush {
		cfg.Git.Push = false
	}
}
