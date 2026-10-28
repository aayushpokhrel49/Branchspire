// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"context"
	"fmt"

	"branchspire/modules/git"
	"branchspire/modules/git/gitrepo"
	"branchspire/modules/setting"
)

// CreateTemporaryGitRepo creates a temporary Git repository empty directory (not initialized)
func CreateTemporaryGitRepo(prefix string) (tmpPath string, tmpRepo git.RepositoryFacade, cancel context.CancelFunc, err error) {
	tmpNamePrefix := prefix + ".git"
	tmpPath, cancel, err = setting.AppDataTempDir("local-repo").MkdirTempRandom(tmpNamePrefix)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create temp dir with prefix %s: %w", tmpNamePrefix, err)
	}
	tmpRepo = gitrepo.RepositoryUnmanaged(tmpPath)
	return tmpPath, tmpRepo, cancel, nil
}
