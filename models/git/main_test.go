// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git_test

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models"
	_ "branchspire/models/actions"
	_ "branchspire/models/activities"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
