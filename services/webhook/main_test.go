// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package webhook

import (
	"testing"

	"branchspire/models/unittest"
	"branchspire/modules/setting"

	_ "branchspire/models"
	_ "branchspire/models/actions"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m, &unittest.TestOptions{
		SetUp: func() error {
			setting.Webhook.AllowedHostList = "loopback:*"
			setting.LoadQueueSettings()
			return Init()
		},
	})
}
