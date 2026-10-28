// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package markup_test

import (
	"os"
	"testing"

	"branchspire/modules/markup"
	"branchspire/modules/setting"
)

func TestMain(m *testing.M) {
	setting.SetupGiteaTestEnv()
	markup.RenderBehaviorForTesting.DisableAdditionalAttributes = true
	markup.RefreshFileNamePatterns()
	os.Exit(m.Run())
}
