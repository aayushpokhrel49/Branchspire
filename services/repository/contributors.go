// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"bufio"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"gitea.dev/models/avatars"
	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/cache"
	"gitea.dev/modules/git"
	"gitea.dev/modules/git/gitcmd"
	"gitea.dev/modules/graceful"
	"gitea.dev/modules/log"
)

const (
	repoContributorsCacheKey         = "RepoContributors/%d/%s"
	repoContributorsCacheTTL   int64 = 60 * 60
	repoContributorsAwaitTime        = 500 * time.Millisecond
	repoContributorsAvatarSize       = 64
	repoContributorsMaxStored        = 100
)

var repoContributorsGenerateLock sync.Map

// RepoContributor is a lightweight contributor entry for repository sidebars.
type RepoContributor struct {
	Name       string `json:"name"`
	AvatarLink string `json:"avatar_link"`
	HomeLink   string `json:"home_link"`
	Commits    int64  `json:"commits"`
}

// RepoContributorList is the cached contributor summary for a repository revision.
type RepoContributorList struct {
	Total int                `json:"total"`
	Top   []*RepoContributor `json:"top"`
}

type repoContributorAuthor struct {
	Name    string
	Commits int64
}

// GetRepoContributors returns the commit authors of the given revision, resolved to local users where possible.
// The result is cached by revision so new commits produce a fresh list without invalidating old entries.
func GetRepoContributors(ctx context.Context, cacheStore cache.StringCache, repo *repo_model.Repository, revision string) (*RepoContributorList, error) {
	if revision == "" {
		revision = repo.DefaultBranch
	}
	cacheKey := fmt.Sprintf(repoContributorsCacheKey, repo.ID, revision)

	var cached RepoContributorList
	if exist, _ := cacheStore.GetJSON(cacheKey, &cached); exist {
		return &cached, nil
	}

	if _, running := repoContributorsGenerateLock.LoadOrStore(cacheKey, struct{}{}); running {
		return nil, ErrAwaitGeneration
	}

	resultCh := make(chan *RepoContributorList, 1)
	go func() {
		defer repoContributorsGenerateLock.Delete(cacheKey)

		list, err := generateRepoContributors(graceful.GetManager().HammerContext(), repo, revision)
		if err != nil {
			log.Error("generateRepoContributors[%s]: %v", repo.FullName(), err)
			resultCh <- nil
			return
		}
		if err := cacheStore.PutJSON(cacheKey, list, repoContributorsCacheTTL); err != nil {
			log.Error("cacheRepoContributors[%s]: %v", repo.FullName(), err)
		}
		resultCh <- list
	}()

	select {
	case list := <-resultCh:
		if list == nil {
			return nil, ErrAwaitGeneration
		}
		return list, nil
	case <-time.After(repoContributorsAwaitTime):
		return nil, ErrAwaitGeneration
	}
}

func generateRepoContributors(ctx context.Context, repo *repo_model.Repository, revision string) (*RepoContributorList, error) {
	gitRepo, closer, err := git.RepositoryFromContextOrOpen(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("generateRepoContributors: open repository: %w", err)
	}
	defer closer.Close()

	authors, err := readRepoContributorAuthors(ctx, gitRepo, revision)
	if err != nil {
		return nil, err
	}
	if len(authors) == 0 {
		return &RepoContributorList{}, nil
	}

	emails := make([]string, 0, len(authors))
	for email := range authors {
		emails = append(emails, email)
	}
	users, err := user_model.GetUsersByEmails(ctx, emails)
	if err != nil {
		return nil, fmt.Errorf("generateRepoContributors: get users by emails: %w", err)
	}

	ghostAvatar := user_model.NewGhostUser().AvatarLinkWithSize(ctx, repoContributorsAvatarSize)
	merged := make(map[string]*RepoContributor, len(authors))
	for email, author := range authors {
		contributor := &RepoContributor{Name: author.Name, Commits: author.Commits}
		key := "email:" + email

		if u := users.GetByEmail(email); u != nil {
			key = fmt.Sprintf("user:%d", u.ID)
			contributor.Name = u.DisplayName()
			contributor.HomeLink = u.HomeLink()
			contributor.AvatarLink = u.AvatarLinkWithSize(ctx, repoContributorsAvatarSize)
		} else {
			contributor.AvatarLink = avatars.GenerateEmailAvatarFastLink(ctx, email, repoContributorsAvatarSize)
			if contributor.AvatarLink == "" {
				contributor.AvatarLink = ghostAvatar
			}
		}

		if existing, ok := merged[key]; ok {
			existing.Commits += contributor.Commits
		} else {
			merged[key] = contributor
		}
	}

	top := make([]*RepoContributor, 0, len(merged))
	for _, contributor := range merged {
		top = append(top, contributor)
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].Commits != top[j].Commits {
			return top[i].Commits > top[j].Commits
		}
		return strings.ToLower(top[i].Name) < strings.ToLower(top[j].Name)
	})

	total := len(top)
	if len(top) > repoContributorsMaxStored {
		top = top[:repoContributorsMaxStored]
	}
	return &RepoContributorList{Total: total, Top: top}, nil
}

func readRepoContributorAuthors(ctx context.Context, gitRepo *git.Repository, revision string) (map[string]*repoContributorAuthor, error) {
	gitCmd := gitcmd.NewCommand("log", "--no-merges", "--use-mailmap", "--format=%aN%x00%aE")
	gitCmd.AddDynamicArguments(revision)

	stdoutReader, stdoutReaderClose := gitCmd.MakeStdoutPipe()
	defer stdoutReaderClose()

	authors := make(map[string]*repoContributorAuthor)
	err := gitCmd.WithRepo(gitRepo).
		WithPipelineFunc(func(gitcmd.Context) error {
			scanner := bufio.NewScanner(stdoutReader)
			for scanner.Scan() {
				name, email, ok := strings.Cut(scanner.Text(), "\x00")
				email = strings.ToLower(strings.TrimSpace(email))
				if !ok || email == "" {
					continue
				}
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				if author, ok := authors[email]; ok {
					author.Commits++
				} else {
					authors[email] = &repoContributorAuthor{Name: name, Commits: 1}
				}
			}
			return scanner.Err()
		}).
		RunWithStderr(ctx)
	if err != nil {
		return nil, fmt.Errorf("readRepoContributorAuthors: git log: %w", err)
	}
	return authors, nil
}
