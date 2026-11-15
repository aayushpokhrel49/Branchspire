// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strconv"

	"gitea.dev/models/db"
	"gitea.dev/models/organization"
	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/log"
	"gitea.dev/modules/optional"
	"gitea.dev/modules/setting"
	"gitea.dev/modules/sitemap"
	"gitea.dev/modules/structs"
	"gitea.dev/modules/templates"
	"gitea.dev/routers/web/feed"
	"gitea.dev/routers/web/user"
	"gitea.dev/services/context"
)

const tplHome templates.TplName = "home"

func Home(ctx *context.Context) {
	if _, feedType := feed.GetFeedType(ctx.Req.URL.Path, ctx.Req); feedType != "" {
		ctx.Resp.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	if ctx.IsSigned {
		user.Dashboard(ctx)
		return
	} else if setting.LandingPageURL != setting.LandingPageHome {
		// Check non-logged users landing page
		ctx.Redirect(setting.AppSubURL + string(setting.LandingPageURL))
		return
	}

	// Check auto-login.
	if ctx.GetSiteCookie(setting.CookieRememberName) != "" {
		ctx.Redirect(setting.AppSubURL + "/user/login")
		return
	}

	ctx.Data["PageIsHome"] = true
	ctx.Data["IsRepoIndexerEnabled"] = setting.Indexer.RepoIndexerEnabled
	ctx.Data["AppVer"] = setting.AppVer
	ctx.Data["NumUsers"] = user_model.CountUsers(ctx, &user_model.CountUserFilter{IsActive: optional.Some(true)})
	numRepos, _ := repo_model.CountRepositories(ctx, repo_model.CountRepositoryOptions{})
	ctx.Data["NumRepos"] = numRepos
	numOrgs, _ := db.Count[organization.Organization](ctx, organization.FindOrgOptions{IncludeVisibility: structs.VisibleTypePrivate})
	ctx.Data["NumOrgs"] = numOrgs
	ctx.HTML(http.StatusOK, tplHome)
}

func About(ctx *context.Context) {
	ctx.Data["PageIsAbout"] = true
	ctx.Data["AppVer"] = setting.AppVer
	ctx.Data["NumUsers"] = user_model.CountUsers(ctx, &user_model.CountUserFilter{IsActive: optional.Some(true)})
	numRepos, _ := repo_model.CountRepositories(ctx, repo_model.CountRepositoryOptions{})
	ctx.Data["NumRepos"] = numRepos
	numOrgs, _ := db.Count[organization.Organization](ctx, organization.FindOrgOptions{IncludeVisibility: structs.VisibleTypePrivate})
	ctx.Data["NumOrgs"] = numOrgs
	ctx.HTML(http.StatusOK, "about")
}

// Legal renders the terms and privacy pages for this instance.
func Legal(ctx *context.Context) {
	page := ctx.PathParam("page")
	switch page {
	case "terms", "privacy":
		ctx.Data["PageIsLegal"] = page
	default:
		ctx.NotFound()
		return
	}
	ctx.HTML(http.StatusOK, "legal")
}

// Docs renders the BranchSpire documentation page for this instance.
func Docs(ctx *context.Context) {
	ctx.HTML(http.StatusOK, "docs")
}

// HomeSitemap renders the main sitemap
func HomeSitemap(ctx *context.Context) {
	m := sitemap.NewSitemapIndex()
	if !setting.Service.Explore.DisableUsersPage {
		_, cnt, err := user_model.SearchUsers(ctx, user_model.SearchUserOptions{
			Types:       []user_model.UserType{user_model.UserTypeIndividual},
			ListOptions: db.ListOptions{PageSize: 1},
			IsActive:    optional.Some(true),
			Visible:     []structs.VisibleType{structs.VisibleTypePublic},
		})
		if err != nil {
			ctx.ServerErrorAuto(err)
			return
		}
		count := int(cnt)
		idx := 1
		for i := 0; i < count; i += setting.UI.SitemapPagingNum {
			m.Add(sitemap.URL{URL: setting.AppURL + "explore/users/sitemap-" + strconv.Itoa(idx) + ".xml"})
			idx++
		}
	}

	_, cnt, err := repo_model.SearchRepository(ctx, repo_model.SearchRepoOptions{
		ListOptions: db.ListOptions{
			PageSize: 1,
		},
		Actor:     ctx.Doer,
		AllPublic: true,
	})
	if err != nil {
		ctx.ServerErrorAuto(err)
		return
	}
	count := int(cnt)
	idx := 1
	for i := 0; i < count; i += setting.UI.SitemapPagingNum {
		m.Add(sitemap.URL{URL: setting.AppURL + "explore/repos/sitemap-" + strconv.Itoa(idx) + ".xml"})
		idx++
	}

	ctx.Resp.Header().Set("Content-Type", "text/xml")
	if _, err := m.WriteTo(ctx.Resp); err != nil {
		log.Error("Failed writing sitemap: %v", err)
	}
}
