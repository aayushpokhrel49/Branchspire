// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2020 The Gitea Authors.
// SPDX-License-Identifier: MIT

package admin

import (
	"branchspire/models/db"
	user_model "branchspire/models/user"
	"branchspire/modules/setting"
	"branchspire/modules/structs"
	"branchspire/modules/templates"
	"branchspire/routers/web/explore"
	"branchspire/services/context"
)

const (
	tplOrgs templates.TplName = "admin/org/list"
)

// Organizations show all the organizations
func Organizations(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.organizations")
	ctx.Data["PageIsAdminOrganizations"] = true

	sortOrder := ctx.FormString("sort", UserSearchDefaultAdminSort)
	explore.RenderUserSearch(ctx, user_model.SearchUserOptions{
		Actor: ctx.Doer,
		Types: []user_model.UserType{user_model.UserTypeOrganization, user_model.UserTypeOrganizationReserved},
		ListOptions: db.ListOptions{
			PageSize: setting.UI.Admin.OrgPagingNum,
		},
		Visible: []structs.VisibleType{structs.VisibleTypePublic, structs.VisibleTypeLimited, structs.VisibleTypePrivate},
		OrderBy: db.SearchOrderBy(sortOrder),
	}, tplOrgs)
}
