// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forms

import "branchspire/modules/web/middleware"

// EditRunnerForm form for admin to create runner
type EditRunnerForm struct {
	middleware.FormDefaultValidator
	Description string
}
