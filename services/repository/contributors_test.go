// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"testing"

	repo_model "gitea.dev/models/repo"
	"gitea.dev/models/unittest"
	"gitea.dev/modules/cache"
	"gitea.dev/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_RepoContributors(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 2})
	require.NoError(t, repo.LoadOwner(t.Context()))
	mockCache, err := cache.NewStringCache(setting.Cache{})
	require.NoError(t, err)

	list, err := GetRepoContributors(t.Context(), mockCache, repo, "master")
	require.NoError(t, err)
	require.NotNil(t, list)
	require.Len(t, list.Top, 3)
	assert.EqualValues(t, 3, list.Total)
	for _, contributor := range list.Top {
		assert.NotEmpty(t, contributor.Name)
		assert.NotEmpty(t, contributor.AvatarLink)
		assert.Positive(t, contributor.Commits)
	}

	cached, err := GetRepoContributors(t.Context(), mockCache, repo, "master")
	require.NoError(t, err)
	assert.Equal(t, list, cached)
}
