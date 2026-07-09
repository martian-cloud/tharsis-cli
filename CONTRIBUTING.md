# Contributing to Martian Cloud projects.

There are many ways you can contribute to Martian Cloud projects.  This document describes some of those ways.  It also describes a few things we request you do as part of making any code or documentation contributions.

## Prerequisites

* **Go >= 1.26** ( [https://golang.org/dl/](https://golang.org/dl/) or [https://golang.org/doc/install](https://golang.org/doc/install) )

## Ways to Contribute

- Report bugs.
- Raise security issues.
- Suggest features or enhancements.
- Make and submit changes to fix bugs or add/enhance functionality.
- Write documentation.
- Answer questions other users ask or might have.
- Write tests.

## Reporting Bugs

- Search the existing GitLab issues to see if someone else already reported it.
- Make sure you're using the latest version of the appropriate project(s) in case it might have already been fixed in a later version.
- If no existing issue matches, file a new GitLab issue.
- Please use [this GitLab-supplied template.](https://gitlab.com/gitlab-org/gitlab/-/blob/master/.gitlab/issue_templates/Bug.md)
- Make sure the following items are clearly described in your bug report:
    - the versions of Martian Cloud projects you're using
    - steps to reproduce the problem
    - how actual results differ from expected results
- If you can include a patch as a proposed fix, please do so.

## Security Issues

If you have discovered something that appears to be a security issue, please report it to the email address listed in the README.md file.

## Suggesting Features or Enhancements

Suggestions for features and enhancements are not required to include a code contribution.  To avoid wasting your time, if you plan to contribute code to implement the feature or enhancement you are suggesting, please file an issue before doing substantial work on the code contribution.  If you are submitting only a suggestion, please make your suggestion as complete and precise as reasonably possible.

## Making Changes (bug fixes or enhancements)

If the project in question has existing unit and/or integration tests, before submitting a code contribution (whether it is a bug fix or an enhancement) make sure to run all available tests:

    make test

    make integration

If the existing tests don't pass with your code contribution, your contribution cannot be accepted until that problem has been resolved.

### Formatting and Style

Please respect the formatting of the project codebase:

- tabs rather than spaces for indentation (we set our IDE to display two spaces for a tab)
- standard Go formatting and error scanning:

    make fmt

    make vet

- we generally try to follow the guidelines in this guide: [Uber's Go styling](https://github.com/uber-go/guide/blob/master/style.md)

## Changelog entries

This project uses [changie](https://changie.dev/) to manage changelog entries. Every MR that introduces a user-facing change must include a changelog fragment, and CI (`changelog-check`) enforces this.

Install changie:

```
# macOS / Linux (Homebrew)
brew install changie

# Windows
winget install miniscruff.changie

# Go
go install github.com/miniscruff/changie@latest
```

See the [changie installation docs](https://changie.dev/guide/installation/) for all options.

Add an entry from the repository root:

```
changie new
```

Pick a kind (Added, Changed, Fixed, Deprecated, Removed, Security) and write a short description. This creates a fragment file in `.changes/unreleased/` — commit it with your MR.

If a change genuinely does not need a changelog entry (e.g. a CI tweak or a docs-only fix), add the `skip-changelog` label to the MR to bypass the check.

> **Note:** the label is evaluated at pipeline creation time, not on retry. If you add `skip-changelog` after a pipeline has already run, trigger a new pipeline to pick it up — go to **Build → Pipelines → Run pipeline** on your branch (no commit needed).

### Creating a release (maintainers)

Releases are cut from the accumulated fragments — no manual tagging or hand-edited changelog. There are two ways to do it; the CI job is the preferred path.

#### Via the CI `Create Release` job (preferred)

This cuts a release with no MR — the unreleased fragments were already reviewed when their own MRs merged.

1. Go to **Build → Pipelines** and either open the latest pipeline on `main`, or click **Run pipeline** with the branch set to `main`.
2. (Optional) Set pipeline variables when running:
   - `RELEASE_VERSION` — the version to cut (e.g. `v1.2.3` for a final, `v1.2.3-alpha.1` for a prerelease). Auto-computed from the fragments' bump levels if unset. Once a prerelease version file exists, pass this explicitly to avoid changie over-bumping.
   - `RELEASE_REF` — the commit, branch, or tag to release from. Defaults to `main` HEAD. Set to an older tag for backport releases (e.g. `v0.35.0`).
3. In the pipeline, click the manual **`Create Release`** job (it only appears on `main`).

The job creates a short-lived temp branch off `RELEASE_REF`, commits the changelog there, and tags that commit. This ensures the tagged commit always has the correct changelog. Then:
- **Prerelease** (hyphen in version): the temp branch is deleted; `main` is untouched. Fragments are kept so they roll into the eventual final release.
- **Final release**: if `RELEASE_REF` is on `main`'s history (normal case), the temp branch is merged into `main` and deleted. If not (backport), the temp branch is deleted without merging.

The tag triggers the build/upload/release pipeline, which publishes the binaries and creates the GitLab release with the changelog notes as its description.

> **One-time setup:** this job requires a `RELEASE_TOKEN` CI/CD variable — a Maintainer-role project access token with the `api` and `write_repository` scopes, configured as **masked** and **protected** — whose bot user is added to `main`'s **"Allowed to push and merge"** list (Settings → Repository → Protected branches). Project access tokens expire, so rotate it before it lapses or releases will start failing.

#### Cutting a prerelease (alpha)

To cut a prerelease, run the `Create Release` CI job with `RELEASE_VERSION=v0.36.0-alpha.1`. The job commits the prerelease changelog on a temp branch and tags that commit — so the tagged commit has the correct changelog — then deletes the temp branch without touching `main`. Fragments are kept so the final release inherits them.

When you later cut the final `v0.36.0`, the temp branch is merged to `main`, the prerelease sections are removed, and the fragments are consumed into the final changelog.

#### Locally (alternative)

Use this if you need to prepare or hand-tweak the changelog before releasing:

1. From the repository root, batch the unreleased fragments into the changelog:

   ```
   make release-prep VERSION=vX.Y.Z
   ```

   Omit `VERSION` to let changie auto-compute the next version from the fragments' bump levels. To cut a prerelease, pass the full version including the suffix (e.g. `VERSION=v0.36.0-alpha.1`) — the hyphen signals prerelease mode. Pass `VERSION` explicitly once any prerelease exists for the cycle.

2. Review the resulting `CHANGELOG.md` change, commit it, and open an MR with the `skip-changelog` label.


3. Once the MR is merged to `main`, the next `Create Release` job run will detect the changelog and include these changes in the release.

## Writing Documentation

If your talents lean more toward writing documentation than code, your contributions of documentation are welcome.  Please make sure your contribution of documentation is accurate.  Also, please try to make it consistent in style with the existing documentation.  There may be other guidelines for documentation style published elsewhere in the project.

## Submitting Changes

- submit your changes to the GitLab project: https://gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli  (Do NOT submit changes to the mirrored GitHub project.)
- do your development in a feature or bug-fix branch based on "main"
- please submit your contribution of code or documentation as a Git pull request
- please respond as promptly as you can to feedback regarding your contribution (in order to save your time and ours)

## Answering Questions

If your talents include answering questions asked by other users, we encourage you to do so in considerate and helpful ways.  In time, we may establish a discussion forum or other official place to discuss use of Martian Cloud projects.

## Testing

If you are adding significant new features or functionality, please include unit tests in your contribution.  For larger contributions, you are welcome to include integration tests.

When writing unit tests, please use mocks where appropriate.

## Contributor License Agreement (CLA)

If we have published a Contributor License Agreement prior to the time you submit a contribution, make sure to sign and submit the agreement before or along with your contribution.

## Licensing of Your Contributions

Your contributions will become licensed under the [Mozilla Public License v2.0](https://www.mozilla.org/en-US/MPL/2.0/)
