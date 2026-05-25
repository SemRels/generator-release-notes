// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-release-notes Authors

package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	featuresSection     = "Features"
	bugFixesSection     = "Bug Fixes"
	otherChangesSection = "Other"
)

type ReleaseContext struct {
	Version        string
	CurrentVersion string
	Branch         string
	Commits        []string
}

type Generator struct{}

var conventionalHeaderPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(ctx ReleaseContext) string {
	sections := map[string][]string{}
	rendered := 0
	for _, commit := range ctx.Commits {
		section, line := classifyCommit(commit)
		if section == "" || line == "" {
			continue
		}
		sections[section] = append(sections[section], line)
		rendered++
	}

	var builder strings.Builder
	if version := displayVersion(ctx.Version); version != "Unreleased" {
		fmt.Fprintf(&builder, "## %s\n\n", version)
	}
	builder.WriteString("## What's Changed")

	for _, section := range []string{featuresSection, bugFixesSection, otherChangesSection} {
		lines := sections[section]
		if len(lines) == 0 {
			continue
		}

		builder.WriteString("\n\n### ")
		builder.WriteString(section)
		builder.WriteString("\n")
		for index, line := range lines {
			if index > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("- ")
			builder.WriteString(line)
		}
	}

	if rendered == 0 {
		builder.WriteString("\n\n- No notable changes")
	}

	switch {
	case strings.TrimSpace(ctx.CurrentVersion) != "" && strings.TrimSpace(ctx.Version) != "":
		fmt.Fprintf(&builder, "\n\n**Full Changelog**: %s...%s", displayVersion(ctx.CurrentVersion), displayVersion(ctx.Version))
	case strings.TrimSpace(ctx.Branch) != "":
		fmt.Fprintf(&builder, "\n\n_Target branch: %s_", strings.TrimSpace(ctx.Branch))
	}

	return builder.String()
}

func classifyCommit(commit string) (string, string) {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return "", ""
	}

	header := firstLine(trimmed)
	matches := conventionalHeaderPattern.FindStringSubmatch(header)
	if len(matches) == 0 {
		if breaking, ok := breakingChangeText(trimmed); ok {
			return otherChangesSection, breaking
		}
		return otherChangesSection, header
	}

	line := header
	if breaking, ok := breakingChangeText(trimmed); ok {
		line = breaking
	} else if matches[3] == "!" {
		line = "BREAKING: " + header
	}

	switch strings.ToLower(matches[1]) {
	case "feat":
		return featuresSection, line
	case "fix", "perf", "revert":
		return bugFixesSection, line
	default:
		return otherChangesSection, line
	}
}

func breakingChangeText(message string) (string, bool) {
	for _, line := range strings.Split(message, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BREAKING CHANGE:") {
			return trimmed, true
		}
	}
	return "", false
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	parts := strings.SplitN(message, "\n", 2)
	return strings.TrimSpace(parts[0])
}

func displayVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "Unreleased"
	}
	if strings.HasPrefix(strings.ToLower(version), "v") {
		return version
	}
	return "v" + version
}
