// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package avatars_test

import (
	"testing"

	"branchspire/models/unittest"

	_ "branchspire/models"
	_ "branchspire/models/activities"
	_ "branchspire/models/perm/access"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
