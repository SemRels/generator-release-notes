// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-release-notes Authors

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	plugin "github.com/SemRels/generator-release-notes/internal/plugin"
)

const pluginSchemaVersion = 1

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv))
}

func run(stdout, stderr io.Writer, getenv func(string) string) int {
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)
	ctx, err := releaseContextFromEnv(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "generator-release-notes:", err)
		return 1
	}

	options := plugin.DefaultGenerateOptions()
	options.Signature = envBool(getenv, "SEMREL_PLUGIN_SIGNATURE", false)
	options.AIDisclosure = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE", false)
	options.AIDisclosureSection = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE_SECTION", false)
	if badge := strings.TrimSpace(getenv("SEMREL_PLUGIN_AI_DISCLOSURE_BADGE")); badge != "" {
		options.AIDisclosureBadge = badge
	}

	if _, err := io.WriteString(stdout, plugin.New().Generate(ctx, options)); err != nil {
		fmt.Fprintln(stderr, "generator-release-notes:", err)
		return 1
	}

	return 0
}

func releaseContextFromEnv(getenv func(string) string) (plugin.ReleaseContext, error) {
	raw := getenv("SEMREL_COMMITS")

	var commits []string
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &commits); err != nil {
			return plugin.ReleaseContext{}, fmt.Errorf("invalid SEMREL_COMMITS JSON: %w", err)
		}
	}

	return plugin.ReleaseContext{
		Version:        firstNonEmpty(getenv("SEMREL_VERSION"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_NEXT_VERSION")),
		CurrentVersion: strings.TrimSpace(getenv("SEMREL_CURRENT_VERSION")),
		Branch:         strings.TrimSpace(getenv("SEMREL_BRANCH")),
		Commits:        commits,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func envBool(getenv func(string) string, key string, defaultValue bool) bool {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
