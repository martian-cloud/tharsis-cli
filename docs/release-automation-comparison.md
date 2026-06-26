# Release Automation: Changie Prerelease Approach Comparison

Analysis of how other OSS projects using changie handle prereleases and release automation,
in response to the rough edges identified in MR !98.

## Universal Pattern

Every project commits the changelog, then tags the commit that contains it.
None tag a commit without the changelog — which is exactly what causes our "tag shows the old
version" edge. The projects differ only in how much branching machinery wraps that.

| Approach | How it works | Keeps `main` pristine for prereleases? | Rough effort |
|---|---|---|---|
| **Light** ([codegen-openapi](https://github.com/hashicorp/terraform-plugin-codegen-openapi/blob/main/.github/workflows/release.yml)) | Commit changelog to the branch → tag it → release; notes read from `.changes/<version>.md` | No (changelog commit lands on the branch) | ~half a day |
| **Medium** ([dbt-core](https://github.com/dbt-labs/dbt-core/blob/main/.github/workflows/release.yml)) | Temp `release-prep` branch holds the changelog commit; prereleases tag it without merging, finals merge to `main`; `RELEASE_REF` input picks the commit | Yes | ~1.5–2 days |
| **Heavy** ([terraform-provider-aws](https://hashicorp.github.io/terraform-provider-aws/design-decisions/changie-migration/)) | `release/N.x` branches + `beta/`/`ga/` subdirs, consolidated on merge to `main` | Yes | multi-day + ongoing process |

Medium's temp-branch logic lives in dbt's reusable workflows:
[release-prep.yml](https://github.com/dbt-labs/dbt-release/blob/main/.github/workflows/release-prep.yml) +
[release-prep-base.yml](https://github.com/dbt-labs/dbt-release/blob/main/.github/workflows/release-prep-base.yml).
dbt commits prerelease changelogs too — see
[`.changes/`](https://github.com/dbt-labs/dbt-core/tree/main/.changes) (e.g. `2.0.0-alpha.1.md`).
changie's own prerelease→final recipe (`--move-dir`/`--include`/`--remove-prereleases`) is in
[discussion #237](https://github.com/miniscruff/changie/discussions/237).

## Rough Edge Mapping

- **Edge 1 (stale changelog at the tag):** fixed by any option — they all commit the changelog before tagging.
- **Edge 2 (release a previous commit):** a small `RELEASE_REF`/ref input (Light/Medium) or inherent in release branches (Heavy).

## Time Rationale

The implementation in each case is modest; the cost is dominated by our testing friction
(we still can't exercise `create-release` from an MR because `RELEASE_TOKEN` is a protected
variable) plus review rounds.

- **Light** is mostly "commit before tagging" + read notes from the version file.
- **Medium** adds temp-branch creation, merge-vs-tag logic, a ref input, and retiring
  `auto-tag-release` + the interim `make prerelease` — hence the jump to ~2 days.
- **Heavy** adds release-branch process overhead that's overkill for a CLI our size.

## Deciding Question

**Do prereleases need to keep `main` pristine?**

- **No** → Light (~half a day). Alpha entries temporarily appear on `main` and are cleaned up at the final release.
- **Yes** → Medium (~1.5–2 days). Alpha changelog commits live only on the temp branch; `main` only ever reflects final releases.

## Reference Links

- [changie itself — release.yml](https://github.com/miniscruff/changie/blob/main/.github/workflows/release.yml)
- [changie discussion #237 — prerelease→final recipe](https://github.com/miniscruff/changie/discussions/237)
- [dbt-core — release.yml](https://github.com/dbt-labs/dbt-core/blob/main/.github/workflows/release.yml)
- [dbt-core — .changes/ (committed prerelease files)](https://github.com/dbt-labs/dbt-core/tree/main/.changes)
- [dbt-release — release-prep.yml](https://github.com/dbt-labs/dbt-release/blob/main/.github/workflows/release-prep.yml)
- [dbt-release — release-prep-base.yml](https://github.com/dbt-labs/dbt-release/blob/main/.github/workflows/release-prep-base.yml)
- [terraform-plugin-codegen-openapi — release.yml](https://github.com/hashicorp/terraform-plugin-codegen-openapi/blob/main/.github/workflows/release.yml)
- [terraform-provider-aws — changie migration design doc](https://hashicorp.github.io/terraform-provider-aws/design-decisions/changie-migration/)
