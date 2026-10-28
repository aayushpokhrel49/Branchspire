# BranchSpire

BranchSpire is a painless self-hosted Git service written in Go, providing an
all-in-one software development platform: Git hosting, code management, code
review, issue tracking, project kanban, wiki, team collaboration, package
registry and CI/CD with GitHub Actions-compatible workflows.

BranchSpire is a fork of [Gitea](https://github.com/go-gitea/gitea). It keeps
Gitea's architecture, data model and configuration, so existing Gitea
deployments can be migrated with minimal changes.

As it is written in Go, it works across **all** the platforms and
architectures supported by Go, including Linux, macOS, FreeBSD/OpenBSD and
Windows on x86, amd64, ARM, RISC-V 64 and PowerPC architectures.

## Documentation

BranchSpire is based on Gitea, and the Gitea documentation applies to it:

- [Installation](https://docs.gitea.com/installation/install-from-binary)
- [Administration](https://docs.gitea.com/administration/config-cheat-sheet)
- [Usage](https://docs.gitea.com/usage/quick-start)

Instance setup starts in [custom/conf/app.example.ini](custom/conf/app.example.ini),
which documents every supported configuration option.

## Building

See [docs/build-setup.md](docs/build-setup.md) for prerequisites
and [docs/development.md](docs/development.md) for setting up a local development environment, linting, and testing.

If you'd like to build from source or make a distribution package, see [docs/build-source.md](docs/build-source.md) for more information.

After building, you can run `./branchspire web` to start the server, or `./branchspire help` to see all available commands.

## Contributing

Expected workflow is: Fork -> Patch -> Push -> Pull Request

> [!NOTE]
>
> 1. **YOU MUST READ THE [CONTRIBUTORS GUIDE](CONTRIBUTING.md) BEFORE STARTING TO WORK ON A PULL REQUEST.**
> 2. New to the codebase? The [development guide](docs/development.md) walks through setting up a local environment and building from source.
> 3. In `options/locale`, only edit `locale_en-US.json`; other locales are synced from upstream automatically.

## License

This project is licensed under the MIT License.
See the [LICENSE](https://github.com/go-gitea/gitea/blob/main/LICENSE) file
for the full license text.

Copyright and attribution to the original Gitea authors is preserved in the
source files and the upstream repository.
