// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"testing"

	"gitea.dev/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEmbeddedRunnerLabels(t *testing.T) {
	old := setting.Actions.EmbeddedRunnerLabels
	setting.Actions.EmbeddedRunnerLabels = []string{"ubuntu-latest", "self-hosted"}
	t.Cleanup(func() { setting.Actions.EmbeddedRunnerLabels = old })

	assert.Equal(t, []string{"ubuntu-latest:host://-", "self-hosted:host://-"}, embeddedRunnerLabelSpecs())

	ls, err := parseEmbeddedRunnerLabels()
	require.NoError(t, err)
	assert.Equal(t, []string{"ubuntu-latest", "self-hosted"}, ls.Names())
	assert.False(t, ls.RequireDocker())
}
