// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	audit_model "branchspire/models/audit"
	shared_audit "branchspire/routers/web/shared/audit"
	"branchspire/services/context"
)

func ViewAuditLogs(ctx *context.Context) {
	shared_audit.View(ctx, shared_audit.ViewOptions{
		Template:  "repo/settings/audit_logs",
		ScopeType: audit_model.ScopeRepository,
		ScopeID:   ctx.Repo.Repository.ID,
		PageData:  map[string]any{"PageIsSettingsAudit": true},
	})
}
