// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package project

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models/actions"
	_ "branchspire/models/activities"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
