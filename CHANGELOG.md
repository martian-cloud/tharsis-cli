## v0.37.0 - 2026-10-28 (9 changes)
### Added
- plan, apply and destroy follow every stage of a run, including pre-plan, post-plan, pre-apply and post-apply policy stages, and show each stage's policy findings.
- When a policy soft-fails, plan, apply and destroy let you approve the gate (or override it when it has no eligible approvers), discard the run, or leave it waiting.
- New -wait and -wait-timeout flags on plan, apply and destroy. With -wait, when a plan approval or policy gate decision can't be asked for (-input=false or no interactive terminal), the CLI keeps following the run while the decision is made in the Tharsis UI. -wait-timeout (default 1h) limits how long a decision is waited on, at the terminal or with -wait in the Tharsis UI, before leaving the run waiting and exiting 1. plan also gains the -input flag.
### Changed
- Plan approval in apply and destroy now takes 'yes' to apply, 'no' to discard the run, or Enter to leave it waiting; previously any answer other than 'yes' left the run planned. Leaving a run waiting at a prompt exits 0; stopping because -wait-timeout passed exits 1.
- With -input=false (or without an interactive terminal), apply and destroy still leave the run waiting for plan approval in the Tharsis UI and exit 0, but first follow any policy stages before it; if a policy gate needs a decision, the run is left waiting at the gate and the CLI exits 1. Pass -wait to wait for these decisions in the Tharsis UI instead.
- With -auto-approve, the server starts the apply as soon as the plan and any post-plan policy stage clear, instead of the CLI applying the run.
- Ctrl-C during plan, apply and destroy asks at an interactive terminal whether to cancel a run that is still running. On SIGTERM, without a terminal, or with -input=false, the CLI stops following and leaves the run going (exit 130, or 143 for SIGTERM); it never cancels or discards a run without being asked. A second interrupt exits immediately.
### Fixed
- Print streamed job logs verbatim so a '%' in the output is no longer interpreted as a format verb and corrupted (e.g. 80%!o(MISSING)f).
### Security
- Upgrade google.golang.org/grpc to v1.83.1 to address grpc-go vulnerabilities.
## v0.36.0 - 2026-10-06 (5 changes)
### Added
- Create Release job supports cutting prereleases (via hyphen in RELEASE_VERSION, e.g. v1.2.3-alpha.1) that tag directly without pushing to main, keeping fragments so they roll into the final release; the job also only runs after a green pipeline.
- Support run annotations via a repeatable --annotation flag on plan, apply, and destroy, and via the MCP create_run tool, so runs can be traced back to what created them (e.g. commit, ref, merge request, pipeline).
### Fixed
- Remove accidentally committed .go/ CI GOPATH cache from repository
- Reconnect log stream on connection errors to prevent jobs from hanging
- Detect duplicate endpoint profiles during SSO login and fail with a clear error instead of non-deterministic auth failures
## v0.35.0 - 2026-06-23 (1 changes)
### ✨ Added
- Changelog automation and release tagging via changie: changelog fragments enforced per MR, automated version tagging from CHANGELOG.md, and a manual job to cut releases.
## v0.34.0 - 2026-05-19 (0 changes)
