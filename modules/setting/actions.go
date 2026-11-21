// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"gitea.dev/modules/log"
)

const (
	// some of the values are from GitHub defaults
	defaultMaxRerunAttempts       = 50
	defaultMaxConcurrentTaskPicks = 16
	defaultArtifactRetentionDays  = 90
	defaultLogRetentionDays       = 365
	defaultRunRetentionDays       = 400
)

// defaultEmbeddedRunnerLabels matches the common hosted-runner labels for the
// server's OS, plus generic labels that always describe a self-hosted host.
func defaultEmbeddedRunnerLabels() []string {
	labels := []string{"self-hosted"}
	switch runtime.GOOS {
	case "linux":
		labels = append([]string{"ubuntu-latest", "ubuntu-24.04", "ubuntu-22.04", "linux"}, labels...)
	case "windows":
		labels = append([]string{"windows-latest", "windows-2022", "windows"}, labels...)
	case "darwin":
		labels = append([]string{"macos-latest", "macos-14", "macos"}, labels...)
	default:
		labels = append([]string{runtime.GOOS}, labels...)
	}
	if arch := embeddedRunnerArchLabel(); arch != "" {
		labels = append(labels, arch)
	}
	return labels
}

func embeddedRunnerArchLabel() string {
	switch runtime.GOARCH {
	case "amd64":
		return "X64"
	case "arm64":
		return "ARM64"
	case "386":
		return "X86"
	default:
		return ""
	}
}

// normalizeEmbeddedRunnerLabels keeps label names, drops executor suffixes
// such as ":host://-", and removes empty or duplicate entries.
func normalizeEmbeddedRunnerLabels(labels []string) []string {
	normalized := make([]string, 0, len(labels))
	seen := make(map[string]bool, len(labels))
	for _, label := range labels {
		name, _, _ := strings.Cut(label, ":")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, name)
	}
	return normalized
}

// Actions settings
var (
	Actions = struct {
		Enabled               bool
		LogStorage            *Storage          // how the created logs should be stored
		LogRetentionDays      int64             `ini:"LOG_RETENTION_DAYS"`
		LogCompression        logCompression    `ini:"LOG_COMPRESSION"`
		ArtifactStorage       *Storage          // how the created artifacts should be stored
		ArtifactRetentionDays int64             `ini:"ARTIFACT_RETENTION_DAYS"`
		RunRetentionDays      int64             `ini:"RUN_RETENTION_DAYS"`
		DefaultActionsURL     defaultActionsURL `ini:"DEFAULT_ACTIONS_URL"`
		ZombieTaskTimeout     time.Duration     `ini:"ZOMBIE_TASK_TIMEOUT"`
		EndlessTaskTimeout    time.Duration     `ini:"ENDLESS_TASK_TIMEOUT"`
		AbandonedJobTimeout   time.Duration     `ini:"ABANDONED_JOB_TIMEOUT"`
		SkipWorkflowStrings   []string          `ini:"SKIP_WORKFLOW_STRINGS"`
		WorkflowDirs          []string          `ini:"WORKFLOW_DIRS"`
		ScopedWorkflowDirs    []string          `ini:"SCOPED_WORKFLOW_DIRS"`
		MaxRerunAttempts      int64             `ini:"MAX_RERUN_ATTEMPTS"`
		// MaxConcurrentTaskPicks bounds how many runners may run the task-assignment
		// transaction at once per Gitea instance, to avoid a thundering herd when many
		// runners poll together. It is a per-process limit, not a cluster-wide one.
		MaxConcurrentTaskPicks int `ini:"MAX_CONCURRENT_TASK_PICKS"`

		ArtifactPreviewMaxSize int64 `ini:"ARTIFACT_PREVIEW_MAX_SIZE"`
		// EmbeddedRunner runs a BranchSpire Runner inside the web process for
		// zero-setup workflow execution. It only self-registers when the
		// instance has no runners, so operators with external runners are
		// unaffected unless they explicitly enable it.
		EmbeddedRunner bool `ini:"EMBEDDED_RUNNER"`
		// EmbeddedRunnerLabels are the runs-on labels the embedded runner
		// advertises. Defaults to labels for the host OS and architecture.
		EmbeddedRunnerLabels []string `ini:"EMBEDDED_RUNNER_LABELS"`
	}{
		Enabled:                true,
		EmbeddedRunner:         true,
		EmbeddedRunnerLabels:   defaultEmbeddedRunnerLabels(),
		DefaultActionsURL:      defaultActionsURLGitHub,
		SkipWorkflowStrings:    []string{"[skip ci]", "[ci skip]", "[no ci]", "[skip actions]", "[actions skip]"},
		WorkflowDirs:           []string{".branchspire/workflows", ".github/workflows"},
		ScopedWorkflowDirs:     []string{".branchspire/scoped_workflows"},
		MaxRerunAttempts:       defaultMaxRerunAttempts,
		MaxConcurrentTaskPicks: defaultMaxConcurrentTaskPicks,
		ArtifactPreviewMaxSize: 10 * 1024 * 1024,
		LogRetentionDays:       defaultLogRetentionDays,
		ArtifactRetentionDays:  defaultArtifactRetentionDays,
		RunRetentionDays:       defaultRunRetentionDays,
	}
)

type defaultActionsURL string

func (url defaultActionsURL) URL() string {
	switch url {
	case defaultActionsURLGitHub:
		return "https://github.com"
	case defaultActionsURLSelf:
		return strings.TrimSuffix(AppURL, "/")
	default:
		// This should never happen, but just in case, use GitHub as fallback
		return "https://github.com"
	}
}

const (
	defaultActionsURLGitHub = "github" // https://github.com
	defaultActionsURLSelf   = "self"   // the root URL of the self-hosted Gitea instance
	// DefaultActionsURL only supports GitHub and the self-hosted Gitea.
	// It's intentionally not supported more, so please be cautious before adding more like "gitea" or "gitlab".
	// If you get some trouble with `uses: username/action_name@version` in your workflow,
	// please consider to use `uses: https://the_url_you_want_to_use/username/action_name@version` instead.
)

type logCompression string

func (c logCompression) IsValid() bool {
	return c.IsNone() || c.IsZstd()
}

func (c logCompression) IsNone() bool {
	return string(c) == "none"
}

func (c logCompression) IsZstd() bool {
	return c == "" || string(c) == "zstd"
}

func loadActionsFrom(rootCfg ConfigProvider) error {
	sec := rootCfg.Section("actions")
	err := sec.MapTo(&Actions)
	if err != nil {
		return fmt.Errorf("failed to map Actions settings: %v", err)
	}

	Actions.EmbeddedRunnerLabels = normalizeEmbeddedRunnerLabels(Actions.EmbeddedRunnerLabels)
	if len(Actions.EmbeddedRunnerLabels) == 0 {
		Actions.EmbeddedRunnerLabels = defaultEmbeddedRunnerLabels()
	}

	if urls := string(Actions.DefaultActionsURL); urls != defaultActionsURLGitHub && urls != defaultActionsURLSelf {
		url, _, _ := strings.Cut(urls, ",")
		if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
			log.Error("[actions] DEFAULT_ACTIONS_URL does not support %q as custom URL any longer, fallback to %q",
				urls,
				defaultActionsURLGitHub,
			)
			Actions.DefaultActionsURL = defaultActionsURLGitHub
		} else {
			return fmt.Errorf("unsupported [actions] DEFAULT_ACTIONS_URL: %q", urls)
		}
	}

	// don't support to read configuration from [actions]
	Actions.LogStorage, err = getStorage(rootCfg, "actions_log", "", nil)
	if err != nil {
		return err
	}

	actionsSec, _ := rootCfg.GetSection("actions.artifacts")

	Actions.ArtifactStorage, err = getStorage(rootCfg, "actions_artifacts", "", actionsSec)
	if err != nil {
		return err
	}

	Actions.ZombieTaskTimeout = sec.Key("ZOMBIE_TASK_TIMEOUT").MustDuration(10 * time.Minute)
	Actions.EndlessTaskTimeout = sec.Key("ENDLESS_TASK_TIMEOUT").MustDuration(3 * time.Hour)
	Actions.AbandonedJobTimeout = sec.Key("ABANDONED_JOB_TIMEOUT").MustDuration(24 * time.Hour)

	if !Actions.LogCompression.IsValid() {
		return fmt.Errorf("invalid [actions] LOG_COMPRESSION: %q", Actions.LogCompression)
	}

	workflowDirs := normalizeWorkflowDirs(Actions.WorkflowDirs)
	if len(workflowDirs) == 0 {
		return errors.New("[actions] WORKFLOW_DIRS must contain at least one entry")
	}
	Actions.WorkflowDirs = workflowDirs

	// SCOPED_WORKFLOW_DIRS may be empty (feature disabled), but it must not overlap with WORKFLOW_DIRS:
	// a scoped dir nested in (or equal to) a workflow dir would make the same file run both repo-level and scope-level.
	Actions.ScopedWorkflowDirs = normalizeWorkflowDirs(Actions.ScopedWorkflowDirs)
	for _, scopedDir := range Actions.ScopedWorkflowDirs {
		for _, workflowDir := range Actions.WorkflowDirs {
			if scopedDir == workflowDir ||
				strings.HasPrefix(scopedDir, workflowDir+"/") ||
				strings.HasPrefix(workflowDir, scopedDir+"/") {
				return fmt.Errorf("[actions] SCOPED_WORKFLOW_DIRS entry %q overlaps with WORKFLOW_DIRS entry %q", scopedDir, workflowDir)
			}
		}
	}

	return nil
}

// normalizeWorkflowDirs trims, normalizes separators and drops empty/trailing-slash entries.
func normalizeWorkflowDirs(dirs []string) []string {
	normalized := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dir = strings.ReplaceAll(dir, `\`, `/`)
		dir = strings.TrimRight(dir, "/")
		normalized = append(normalized, dir)
	}
	return normalized
}
