// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"

	audit_model "branchspire/models/audit"
	user_model "branchspire/models/user"
	"branchspire/services/audit"
)

func cliAuditContext(ctx context.Context) context.Context {
	ctx = audit.WithOrigin(ctx, audit_model.OriginCLI)
	return audit.WithDoer(ctx, user_model.NewCliUser())
}
