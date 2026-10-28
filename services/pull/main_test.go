// Copyright 2019 The Gitea Authors.
// All rights reserved.
// SPDX-License-Identifier: MIT

package pull

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models/actions"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
