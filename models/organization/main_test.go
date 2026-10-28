// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package organization_test

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models"
	_ "branchspire/models/actions"
	_ "branchspire/models/activities"
	_ "branchspire/models/organization"
	_ "branchspire/models/repo"
	_ "branchspire/models/user"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
