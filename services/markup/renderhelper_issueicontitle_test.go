// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package markup

import (
	"testing"

	issue_model "branchspire/models/issues"
	perm_model "branchspire/models/perm"
	"branchspire/models/repo"
	"branchspire/models/unit"
	"branchspire/models/unittest"
	"branchspire/modules/markup"
	"branchspire/modules/templates"
	"branchspire/modules/util"
	"branchspire/services/contexttest"

	"github.com/stretchr/testify/assert"
	"xorm.io/builder"
)

func TestRenderHelperIssueIconTitle(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	t.Run("RenderInCurrentRepo", func(t *testing.T) {
		ctx, _ := contexttest.MockContext(t, "/", contexttest.MockContextOption{Render: templates.PageRenderer()})
		contexttest.LoadRepo(t, ctx, 1)
		htm, err := renderRepoIssueIconTitle(ctx, markup.RenderIssueIconTitleOptions{
			LinkHref:   "/link",
			IssueIndex: 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, `<a href="/link"><span>octicon-issue-opened(16/tw-text-green)</span> issue1 (#1)</a>`, string(htm))

		ctx.Repo.Permission.SetUnitsWithDefaultAccessMode([]*repo.RepoUnit{{Type: unit.TypeWiki}}, perm_model.AccessModeRead)
		issueA := unittest.AssertExistsAndLoadBean(t, &issue_model.Issue{}, builder.Eq{"repo_id": 1, "`index`": 1, "is_pull": false})
		issueB := unittest.AssertExistsAndLoadBean(t, &issue_model.Issue{}, builder.Eq{"repo_id": 1, "`index`": 2, "is_pull": true})
		for _, issueIndex := range []int64{issueA.Index, issueB.Index} {
			_, err = renderRepoIssueIconTitle(ctx, markup.RenderIssueIconTitleOptions{
				LinkHref:   "/link",
				IssueIndex: issueIndex,
			})
			assert.ErrorIs(t, err, util.ErrPermissionDenied)
		}
	})

	t.Run("RenderAcrossRepo", func(t *testing.T) {
		ctx, _ := contexttest.MockContext(t, "/", contexttest.MockContextOption{Render: templates.PageRenderer()})
		htm, err := renderRepoIssueIconTitle(ctx, markup.RenderIssueIconTitleOptions{
			OwnerName:  "user2",
			RepoName:   "repo1",
			LinkHref:   "/link",
			IssueIndex: 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, `<a href="/link"><span>octicon-issue-opened(16/tw-text-green)</span> issue1 (user2/repo1#1)</a>`, string(htm))

		ctx, _ = contexttest.MockContext(t, "/", contexttest.MockContextOption{Render: templates.PageRenderer()})
		_, err = renderRepoIssueIconTitle(ctx, markup.RenderIssueIconTitleOptions{
			OwnerName:  "user2",
			RepoName:   "repo2",
			LinkHref:   "/link",
			IssueIndex: 2,
		})
		assert.ErrorIs(t, err, util.ErrPermissionDenied)
	})
}
