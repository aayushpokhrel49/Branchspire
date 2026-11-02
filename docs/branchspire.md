# BranchSpire

BranchSpire is a self-hosted Git service written in Go: repository hosting,
code review, issues, wiki, and Actions-compatible CI/CD in one binary.

It is a fork of [Gitea](https://github.com/go-gitea/gitea), forked from
upstream `v29.0.0-dev` (commit `217300fff7`). The architecture, data model,
HTTP/API surface, CLI, and configuration system are inherited from Gitea —
do not duplicate that documentation here. This page documents only what is
particular to BranchSpire.

## Naming

| Thing | Value |
| --- | --- |
| Binary and module path | `branchspire` |
| Configuration directory | `custom/` (`custom/conf/app.ini`) |
| Web UI and instance name | `BranchSpire` (`APP_NAME`) |
| Built-in self-migration service | `BranchSpire` |

The migrator that imports repositories from other Gitea instances is renamed
`BranchSpire` (it is `service_type=3`, upstream type `3` = Gitea service). Its
pieces are: `GitServiceType.Title()` in `modules/structs/repo.go` (which
`Name()` lowercases for lookups), the chooser icon
`web_src/svg/gitea-branchspire.svg`, the description locale key
`repo.migrate.branchspire.description`, and the form template
`templates/repo/migrate/branchspire.tmpl`. Keep these consistent if the
service is rebranded again.

## What differs from upstream Gitea

Branding and application shell:

- BranchSpire logo (`public/assets/img/logo.svg` — the paper tile with the
  tree mark) is used in the navbar, footer, and migration chooser.
- The app name is `BranchSpire`, set by `APP_NAME` in the config (the CLI
  banner and page chrome derive from it).
- All pages share a custom footer with Product / Resource / Legal columns and
  an instance tagline (`templates/base/footer_content.tmpl`,
  `templates/base/footer_links.tmpl`, locale strings under `home.`).

Composed pages (`web_src/css/home.css`, `web_src/css/about.css`):

- `/` is the sign-up oriented landing ("Your Git. Your rules."), a fixed
  print-gray ("paper") page whose tints do not change with the user's UI
  theme.
- `/about` is the dark midnight presentation page ("Where your branches
  converge."), also theme-independent.
- Both reuse the custom footer.

Sign-in / sign-up (`web_src/css/auth.css`, commit `be5a977264`):

- A dark "ink" stage beside the form on a light "paper" panel.
- Everything upstream still applies conditionally: local password sign-in,
  passkeys (`EnablePasskeyAuth`), OAuth2/OpenID/SSPI providers
  (`templates/user/auth/external_auth_methods.tmpl`), captcha
  (`templates/user/auth/captcha.tmpl`), e-mail confirmation, and
  must-change-password handling.
- `templates/user/auth/signin.tmpl` and `signup.tmpl` currently also link
  `public/assets/auth-preview.css` — that file is a transitional copy of
  `web_src/css/auth.css` for the current binary whose bundled CSS predates
  the design. Remove those two links after the next frontend build.

Navigation bar (`web_src/css/modules/navbar.css`):

- Flat single-tone bar, no borders or shadows; items and buttons are
  fully-rounded chips; anonymous users get one ghost "Log in" link and one
  filled "Register" action tinted with the same ink tokens as the auth pages.
- The app name stays visible next to the logo on phones.
- Behaviour (dropdowns, icons, unread-count badges, mobile menu) is
  unchanged.

Dashboard onboarding (commit `203d7d60ec`):

- The signed-out-style guide card on the dashboard uses
  `User.CanCreateRepo` (added in `models/user/user.go`) so the
  "create repository" button only shows when the user may create one.

## Configuration

Everything lives in `custom/conf/app.ini`, created on first run. The full
option-by-option reference is
[`custom/conf/app.example.ini`](../custom/conf/app.example.ini); the upstream
Gitea docs linked at the bottom of this page cover the most common topics.

Essentials verified on this instance:

- `[server] HTTP_PORT = 3000` is the default web port; `ROOT_URL` must match
  whatever host serves the instance.
- `[database] DB_TYPE` accepts `mysql`, `postgres`, `mssql`, or `sqlite3`;
  `sqlite3` needs no extra server and is what this instance uses.
- `[security]` keys (internal token, secret key) are generated for you; keep
  them out of screenshots and issue reports.

First run: `./branchspire web` with no config file serves the web installer
at `/` (`routers/install/routes.go`), writes the config file, then restarts
into the normal server. The `migrate` and `doctor` subcommands handle schema
upgrades and consistency checks afterwards.

## Build and run

Requirements: a Go toolchain matching `go.mod`, Node.js and pnpm for the
frontend (details in [build-setup.md](build-setup.md)).

- `make build` — builds frontend and backend into `./branchspire`, stamping
  the version from `git describe` (or `STORED_VERSION_FILE` when it exists;
  a build without it reports `version development`).
- `./branchspire web` — start the server; `./branchspire help` lists the
  other subcommands (`admin`, `dump`, `doctor`, `migrate`,
  `migrate-storage`, `dump-repo`, `restore-repo`, `actions`, …).

Development workflow, linting, and tests are in
[development.md](development.md) and [testing.md](testing.md). `make help`
lists every target (`fmt`, `lint-css`, `lint-go`, `test-backend`,
`watch-frontend`, and so on).

## Translations

Only edit `options/locale/locale_en-US.json`. Every other language file is
synced from the upstream Gitea translations automatically; manual edits
there are thrown away on the next sync. New UI strings must define the
`en-US` key, or the raw key name leaks into rendered pages (this is exactly
how `repo.migrate.branchspire.description` surfaced).

## Upstream reference

BranchSpire's behavior outside the customization list above follows the
Gitea release it tracks (administration, config cheat sheet, usage):
<https://docs.gitea.com>. API reference: served at `/api/swagger`.
