// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models"
	_ "branchspire/models/actions"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
