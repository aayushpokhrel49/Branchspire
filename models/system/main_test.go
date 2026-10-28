// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package system_test

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models" // register models
	_ "branchspire/models/actions"
	_ "branchspire/models/activities"
	_ "branchspire/models/system" // register models of system
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
