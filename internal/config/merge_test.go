package config

import "testing"

func TestApplyFlags_DryRun(t *testing.T) {
	cfg := DefaultConfig()
	dryRun := true

	ApplyFlags(cfg, FlagOverrides{DryRun: &dryRun})

	if !cfg.DryRun {
		t.Error("DryRun should be true after applying flag")
	}
}

func TestApplyFlags_CI(t *testing.T) {
	cfg := DefaultConfig()
	ci := true

	ApplyFlags(cfg, FlagOverrides{CI: &ci})

	if !cfg.CI {
		t.Error("CI should be true after applying flag")
	}
}

func TestApplyFlags_Verbose(t *testing.T) {
	cfg := DefaultConfig()
	verbose := 2

	ApplyFlags(cfg, FlagOverrides{Verbose: &verbose})

	if cfg.Verbose != 2 {
		t.Errorf("Verbose = %d, expected 2", cfg.Verbose)
	}
}

func TestApplyFlags_Increment(t *testing.T) {
	cfg := DefaultConfig()
	inc := "major"

	ApplyFlags(cfg, FlagOverrides{Increment: &inc})

	if cfg.Increment != "major" {
		t.Errorf("Increment = %q, expected %q", cfg.Increment, "major")
	}
}

func TestApplyFlags_EmptyIncrement_NotApplied(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Increment = "minor"
	empty := ""

	ApplyFlags(cfg, FlagOverrides{Increment: &empty})

	if cfg.Increment != "minor" {
		t.Errorf("Increment should remain %q when empty flag provided, got %q", "minor", cfg.Increment)
	}
}

func TestApplyFlags_NoCommit(t *testing.T) {
	cfg := DefaultConfig()
	noCommit := true

	ApplyFlags(cfg, FlagOverrides{NoCommit: &noCommit})

	if cfg.Git.Commit {
		t.Error("git.commit should be false after --no-git.commit")
	}
}

func TestApplyFlags_NoTag(t *testing.T) {
	cfg := DefaultConfig()
	noTag := true

	ApplyFlags(cfg, FlagOverrides{NoTag: &noTag})

	if cfg.Git.Tag {
		t.Error("git.tag should be false after --no-git.tag")
	}
}

func TestApplyFlags_NoPush(t *testing.T) {
	cfg := DefaultConfig()
	noPush := true

	ApplyFlags(cfg, FlagOverrides{NoPush: &noPush})

	if cfg.Git.Push {
		t.Error("git.push should be false after --no-git.push")
	}
}

func TestApplyFlags_NilFields_NoChange(t *testing.T) {
	cfg := DefaultConfig()
	original := *cfg

	ApplyFlags(cfg, FlagOverrides{})

	if cfg.DryRun != original.DryRun {
		t.Error("DryRun should not change with nil flag")
	}
	if cfg.CI != original.CI {
		t.Error("CI should not change with nil flag")
	}
	if cfg.Git.Commit != original.Git.Commit {
		t.Error("git.commit should not change with nil flag")
	}
}

func TestApplyFlags_PreReleaseID(t *testing.T) {
	cfg := DefaultConfig()
	preID := "beta"

	ApplyFlags(cfg, FlagOverrides{PreReleaseID: &preID})

	if cfg.PreReleaseID != "beta" {
		t.Errorf("PreReleaseID = %q, expected %q", cfg.PreReleaseID, "beta")
	}
}
