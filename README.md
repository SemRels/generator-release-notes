# generator-release-notes

[![Latest Release](https://img.shields.io/github/v/release/SemRels/generator-release-notes?label=version&color=blue)](https://github.com/SemRels/generator-release-notes/releases/latest)
[![CI](https://github.com/SemRels/generator-release-notes/actions/workflows/ci.yml/badge.svg)](https://github.com/SemRels/generator-release-notes/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/SemRels/generator-release-notes)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SemRels/generator-release-notes/badge)](https://scorecard.dev/viewer/?uri=github.com/SemRels/generator-release-notes)
[![REUSE status](https://api.reuse.software/badge/github.com/SemRels/generator-release-notes)](https://api.reuse.software/info/github.com/SemRels/generator-release-notes)

Generates release notes text for providers and notification hooks.

This plugin is distributed as the standalone Go binary `semrel-plugin-generator-release-notes`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/generator-release-notes/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/generator-release-notes:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/generator-release-notes:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/generator-release-notes/.github/workflows/container.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - uses: "@semrel/generator-release-notes"
    phase: generator
    args:
      template: ".semrel/templates/release-notes.tmpl"
      max_commits: "50"
      include_body: "false"
```

## Options

Every option can be set as an entry under `args` in `.semrel.yaml`. semrel passes each entry to the plugin as the environment variable shown in the table, so you can also set that variable directly (for example in your CI environment) instead of using `args`.

| Argument | Environment variable | Required | Description | Default |
| --- | --- | --- | --- | --- |
| `template` | `SEMREL_PLUGIN_TEMPLATE` | Optional | Path to a custom template file. | Built-in template |
| `max_commits` | `SEMREL_PLUGIN_MAX_COMMITS` | Optional | Maximum number of commits to include. | 50 |
| `include_body` | `SEMREL_PLUGIN_INCLUDE_BODY` | Optional | Include full commit bodies in the generated notes. | false |
| `signature` | `SEMREL_PLUGIN_SIGNATURE` | Optional | Append an opt-in footer line linking to `semrel.io` after the generated release notes output. | `false` |
| `ai_disclosure` | `SEMREL_PLUGIN_AI_DISCLOSURE` | Optional | Detect AI co-author trailers and append a badge to each AI-assisted commit line. | `false` |
| `ai_disclosure_badge` | `SEMREL_PLUGIN_AI_DISCLOSURE_BADGE` | Optional | Badge appended to AI-assisted lines. | `🤖` |
| `ai_disclosure_section` | `SEMREL_PLUGIN_AI_DISCLOSURE_SECTION` | Optional | Append a collapsible "🤖 AI-Assisted Contributions" section (requires `SEMREL_PLUGIN_AI_DISCLOSURE=true`). | `false` |
| `new_contributors` | `SEMREL_PLUGIN_NEW_CONTRIBUTORS` | Optional | Render a "New Contributors" section listing first-time contributors for this release (requires contributor data via `SEMREL_PLUGIN_CONTRIBUTORS_JSON`). | `true` |
| `contributors_json` | `SEMREL_PLUGIN_CONTRIBUTORS_JSON` | Optional | JSON array of first-time contributors for this release, e.g. `[{"name":"Alice","login":"alice","pr":42}]`. Can be populated from semrel core's `SEMREL_CONTRIBUTORS` (filter to `firstContribution:true`). | — |
| `mvp` | `SEMREL_PLUGIN_MVP` | Optional | Render a "🏆 MVP" section highlighting the top contributor for this release (requires `SEMREL_PLUGIN_NEW_CONTRIBUTORS=true` and non-empty contributors). | `false` |
| `mvp_metric` | `SEMREL_PLUGIN_MVP_METRIC` | Optional | MVP ranking metric: `commits` (default, PR-reference count) or `impact` (weights breaking changes/features more heavily). | `commits` |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CURRENT_VERSION` | Current version before the new release is applied. |
| `SEMREL_BRANCH` | Git branch associated with the current release run. |
| `SEMREL_REPOSITORY_URL` | Repository URL used to render contributor links and MVP mentions. |

## Example behavior

The plugin renders release notes text that can be published by provider plugins or reused in notifications.

Set `SEMREL_PLUGIN_SIGNATURE=true` to append this footer after the generated notes:

```md
---
*Generated by [semrel.io](https://semrel.io)*
```

## License

Apache-2.0
