// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repostats

import (
	"testing"

	activities_model "branchspire/models/activities"
	"branchspire/models/organization"
	repo_model "branchspire/models/repo"
	"branchspire/models/unittest"
	user_model "branchspire/models/user"

	_ "branchspire/models/actions"
	_ "branchspire/models/system"

	"github.com/stretchr/testify/assert"
)

// TestFixturesAreConsistent assert that test fixtures are consistent
func TestFixturesAreConsistent(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())
	unittest.CheckConsistencyFor(t,
		&user_model.User{},
		&repo_model.Repository{},
		&organization.Team{},
		&activities_model.Action{})
}

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
