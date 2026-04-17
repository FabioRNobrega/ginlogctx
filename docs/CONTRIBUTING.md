# Contributing to `ginlogctx`

Thanks for contributing to `ginlogctx`.

This document describes the Git workflow conventions used in this repository, including:
- branch naming
- commit message format
- tag and release naming

## Git Guidelines

### Branch Names

Use one of these prefixes for branches:

- `feat/branch-name`
- `hotfix/branch-name`
- `poc/branch-name`

Examples:

- `feat/custom-fields-api`
- `feat/readme-improvements`
- `hotfix/webserver-user-id-binding`
- `poc/context-propagation-experiment`

### Commit Prefixes

Use conventional-style commit prefixes:

- `chore(scope): message`
- `feat(scope): message`
- `fix(scope): message`
- `refactor(scope): message`
- `tests(scope): message`
- `docs(scope): message`

Examples:

```text
feat(middleware): support custom request-scoped fields
fix(webserver): bind user_id before request logging
docs(readme): add installation and usage examples
tests(hook): cover custom field resolution
refactor(config): simplify default field behavior
chore(release): prepare v0.1.0
```

## Suggested Workflow

1. Create a branch using one of the approved prefixes.
2. Make your changes in small, reviewable commits.
3. Use the commit prefixes above.
4. Open a pull request with a clear summary of the change.

Example:

```bash
git checkout -b feat/custom-fields-api
git add .
git commit -m "feat(middleware): support configurable custom fields"
git push origin feat/custom-fields-api
```

## Tags and Releases

Git tags are used to mark versions of the module.

Recommended format:

- `v0.1.0`
- `v0.1.1`
- `v0.2.0`
- `v1.0.0`

Examples:

- `v0.1.0` for the first public release
- `v0.1.1` for a backward-compatible bug fix
- `v0.2.0` for a backward-compatible feature release
- `v1.0.0` for the first stable major release

### Create a Tag from the Console

You can do everything from the terminal. No GitHub UI setup is required.

Annotated tag:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

Lightweight tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Annotated tags are recommended because they carry a message and are more explicit for releases.

## Typical Release Flow

```bash
git checkout master
git pull origin master
git add .
git commit -m "chore(release): prepare v0.1.0"
git push origin master
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

## Notes for Go Modules

If you want the new version to appear on `pkg.go.dev`, make sure:

- the repository is pushed
- `go.mod` is at the repository root
- the tag is pushed to GitHub
- the tag follows semantic versioning such as `v0.1.0`

After the tag is pushed, `pkg.go.dev` can discover it automatically.
