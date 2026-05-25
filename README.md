# generator-release-notes

Release notes generator plugin for Semantic Release.

Generates release notes content from Semantic Release metadata.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/generator-release-notes>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: generator-release-notes
    type: generator
    config:
      format: markdown
      include_commits: true
      include_compare_url: true
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.