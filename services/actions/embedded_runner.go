// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	runner_client "gitea.dev/internal/runner/pkg/client"
	runner_config "gitea.dev/internal/runner/pkg/config"
	runner_labels "gitea.dev/internal/runner/pkg/labels"
	runner_poll "gitea.dev/internal/runner/poll"
	runner_run "gitea.dev/internal/runner/run"
	actions_model "gitea.dev/models/actions"
	"gitea.dev/modules/graceful"
	"gitea.dev/modules/log"
	"gitea.dev/modules/setting"

	"github.com/google/uuid"
)

// The embedded runner reuses the workflow engine vendored in
// gitea.dev/internal/runner (from gitea.com/gitea/runner, MIT licensed), so a
// fresh BranchSpire instance can run workflows without installing or
// registering a separate runner process.

// runnerRegistrationFile keeps the identity (UUID and token) of the embedded
// runner across restarts, like the .runner file of a standalone runner.
func runnerRegistrationFile() string {
	return filepath.Join(setting.AppDataPath, "runner", "runner.json")
}

const embeddedRunnerName = "BranchSpire Runner (embedded)"

// embeddedRunnerLabelSpecs declares host execution for every configured label:
// jobs run as processes of the BranchSpire service account, no Docker required.
func embeddedRunnerLabelSpecs() []string {
	names := setting.Actions.EmbeddedRunnerLabels
	specs := make([]string, 0, len(names))
	for _, name := range names {
		specs = append(specs, name+":host://-")
	}
	return specs
}

func parseEmbeddedRunnerLabels() (runner_labels.Labels, error) {
	specs := embeddedRunnerLabelSpecs()
	ls := make(runner_labels.Labels, 0, len(specs))
	for _, spec := range specs {
		label, err := runner_labels.Parse(spec)
		if err != nil {
			return nil, fmt.Errorf("parse embedded runner label %q: %w", spec, err)
		}
		ls = append(ls, label)
	}
	return ls, nil
}

var errEmbeddedRunnerSkipped = errors.New("runners are already registered")

// StartEmbeddedRunner brings up the embedded BranchSpire Runner in the web
// process. It is a no-op when Actions or the embedded runner are disabled, or
// when this instance already has runners registered.
func StartEmbeddedRunner() {
	if !setting.Actions.Enabled || !setting.Actions.EmbeddedRunner {
		return
	}

	go func() {
		reg, err := loadOrCreateEmbeddedRunnerRegistration()
		if err != nil {
			if errors.Is(err, errEmbeddedRunnerSkipped) {
				log.Info("%s not started: %v", embeddedRunnerName, err)
			} else {
				log.Error("%s failed to start: %v", embeddedRunnerName, err)
			}
			return
		}

		cfg, ls, err := embeddedRunnerConfig(reg)
		if err != nil {
			log.Error("%s failed to start: %v", embeddedRunnerName, err)
			return
		}

		cli := runner_client.New(reg.Address, false, reg.UUID, reg.Token)
		runner := runner_run.NewRunner(cfg, reg, cli)

		// The web server may not be fully listening yet; retry declare with backoff.
		var declareErr error
		for i := 0; i < 60; i++ {
			_, declareErr = runner.Declare(context.Background(), ls.Names())
			if declareErr == nil {
				break
			}
			if errors.Is(declareErr, context.Canceled) {
				log.Error("%s declare cancelled: %v", embeddedRunnerName, declareErr)
				return
			}
			log.Debug("%s declare failed (attempt %d/60): %v; retrying...", embeddedRunnerName, i+1, declareErr)
			select {
			case <-graceful.GetManager().IsShutdown():
				log.Error("%s declare aborted: server shutting down", embeddedRunnerName)
				return
			case <-time.After(time.Second):
			}
		}
		if declareErr != nil {
			log.Error("%s failed to declare after retries: %v", embeddedRunnerName, declareErr)
			return
		}
		log.Info("%s started, jobs run in host exec mode with capacity %d", embeddedRunnerName, cfg.Runner.Capacity)

		poller := runner_poll.New(cfg, cli, runner)
		graceful.GetManager().RunWithCancel(&embeddedRunnerLifecycle{poller: poller})
	}()
}

// embeddedRunnerLifecycle lets the graceful manager stop the poller on
// shutdown, giving running jobs time to report their final state.
type embeddedRunnerLifecycle struct {
	poller *runner_poll.Poller
}

func (l *embeddedRunnerLifecycle) Run() {
	l.poller.Poll()
}

func (l *embeddedRunnerLifecycle) Cancel() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := l.poller.Shutdown(ctx); err != nil {
		log.Warn("%s shutdown: %v", embeddedRunnerName, err)
	}
}

func loadOrCreateEmbeddedRunnerRegistration() (*runner_config.Registration, error) {
	ctx := context.Background()
	regFile := runnerRegistrationFile()
	names := slices.Clone(setting.Actions.EmbeddedRunnerLabels)
	specs := embeddedRunnerLabelSpecs()
	if reg, err := runner_config.LoadRegistration(regFile); err == nil {
		if runner, err := actions_model.GetRunnerByUUID(ctx, reg.UUID); err == nil {
			if !slices.Equal(reg.Labels, specs) || !slices.Equal(runner.AgentLabels, names) {
				runner.AgentLabels = names
				if err := actions_model.UpdateRunner(ctx, runner, "agent_labels"); err != nil {
					return nil, fmt.Errorf("update embedded runner labels: %w", err)
				}
				reg.Labels = specs
				if err := runner_config.SaveRegistration(regFile, reg); err != nil {
					return nil, fmt.Errorf("save runner registration file: %w", err)
				}
				log.Info("%s labels refreshed: %v", embeddedRunnerName, names)
			}
			return reg, nil
		}
		// The database was reset after the registration file was written, so
		// the stored identity points to a runner that no longer exists.
		_ = os.Remove(regFile)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read runner registration file: %w", err)
	}

	count, err := actions_model.CountRunners(ctx)
	if err != nil {
		return nil, fmt.Errorf("count runners: %w", err)
	}
	if count > 0 {
		return nil, errEmbeddedRunnerSkipped
	}

	runner := &actions_model.ActionRunner{
		UUID:        uuid.New().String(),
		Name:        embeddedRunnerName,
		Version:     "embedded",
		AgentLabels: names,
	}
	runner.GenerateAndFillToken()
	if err := actions_model.CreateRunner(ctx, runner); err != nil {
		return nil, fmt.Errorf("create embedded runner: %w", err)
	}

	reg := &runner_config.Registration{
		ID:      runner.ID,
		UUID:    runner.UUID,
		Name:    runner.Name,
		Token:   runner.Token,
		Address: setting.AppURL,
		Labels:  specs,
	}
	if err := os.MkdirAll(filepath.Dir(regFile), os.ModePerm); err != nil {
		return nil, fmt.Errorf("create runner data directory: %w", err)
	}
	if err := runner_config.SaveRegistration(regFile, reg); err != nil {
		return nil, fmt.Errorf("save runner registration file: %w", err)
	}
	return reg, nil
}

func embeddedRunnerConfig(reg *runner_config.Registration) (*runner_config.Config, runner_labels.Labels, error) {
	cfg, err := runner_config.LoadDefault("")
	if err != nil {
		return nil, nil, fmt.Errorf("runner default config: %w", err)
	}
	cfg.Runner.File = runnerRegistrationFile()
	cfg.Runner.Envs = map[string]string{}
	cfg.Host.WorkdirParent = filepath.Join(setting.AppDataPath, "runner", "workdir")
	cfg.Cache.Dir = filepath.Join(setting.AppDataPath, "runner", "cache")

	ls, err := parseEmbeddedRunnerLabels()
	if err != nil {
		return nil, nil, err
	}
	return cfg, ls, nil
}
