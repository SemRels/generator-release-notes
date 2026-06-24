# generator-release-notes

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
  --certificate-identity-regexp 'https://github.com/SemRels/generator-release-notes/.github/workflows/release.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - name: generator-release-notes
    path: ~/.semrel/plugins/semrel-plugin-generator-release-notes
    env:
      SEMREL_PLUGIN_TEMPLATE: ".semrel/templates/release-notes.tmpl"
      SEMREL_PLUGIN_MAX_COMMITS: "50"
      SEMREL_PLUGIN_INCLUDE_BODY: "false"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_TEMPLATE` | Optional | Path to a custom template file. | Built-in template |
| `SEMREL_PLUGIN_MAX_COMMITS` | Optional | Maximum number of commits to include. | 50 |
| `SEMREL_PLUGIN_INCLUDE_BODY` | Optional | Include full commit bodies in the generated notes. | false |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CURRENT_VERSION` | Current version before the new release is applied. |
| `SEMREL_BRANCH` | Git branch associated with the current release run. |

## Example behavior

The plugin renders release notes text that can be published by provider plugins or reused in notifications.

## License

Apache-2.0
