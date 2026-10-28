// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo_test

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models" // register table model
	_ "branchspire/models/actions"
	_ "branchspire/models/activities"
	_ "branchspire/models/perm/access" // register table model
	_ "branchspire/models/repo"        // register table model
	_ "branchspire/models/user"        // register table model
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
