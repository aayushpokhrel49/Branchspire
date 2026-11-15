// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"

	audit_model "code.gitea.io/gitea/models/audit"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/services/audit"
)

func cliAuditContext(ctx context.Context) context.Context {
	ctx = audit.WithOrigin(ctx, audit_model.OriginCLI)
	return audit.WithDoer(ctx, user_model.NewCliUser())
}
