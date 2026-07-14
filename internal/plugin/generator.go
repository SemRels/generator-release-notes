// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-release-notes Authors

package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

// aiTrailerPatterns matches known AI co-author trailers in commit messages.
var aiTrailerPatterns = []struct {
	pattern *regexp.Regexp
	label   string
}{
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*copilot`), "GitHub Copilot"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*github-copilot`), "GitHub Copilot"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*claude`), "Claude (Anthropic)"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*chatgpt`), "ChatGPT (OpenAI)"},
	{regexp.MustCompile(`(?i)(?m)^ai-assisted:\s*true`), "AI"},
	{regexp.MustCompile(`(?i)(?m)^generated-by:`), "AI"},
}

// detectAIAuthors returns deduplicated AI tool labels found in commit trailers.
func detectAIAuthors(commitMsg string) []string {
	seen := map[string]struct{}{}
	var labels []string
	for _, pat := range aiTrailerPatterns {
		if pat.pattern.MatchString(commitMsg) {
			if _, ok := seen[pat.label]; !ok {
				seen[pat.label] = struct{}{}
				labels = append(labels, pat.label)
			}
		}
	}
	return labels
}

const (
	featuresSection     = "Features"
	bugFixesSection     = "Bug Fixes"
	otherChangesSection = "Other"
)

// SectionRule maps a conventional-commit type to a custom changelog section
// heading, mirroring the "presetConfig.types" mapping supported by
// @semantic-release/release-notes-generator. When Hidden is true, commits of
// this Type are omitted from the release notes entirely (Section is ignored).
type SectionRule struct {
	// Type is the conventional-commit type this rule applies to (e.g. "feat",
	// "fix", "chore"). Matching is case-insensitive.
	Type string `json:"type"`
	// Section is the heading commits of this Type are grouped under (e.g.
	// "Features", "Dependencies"). Ignored when Hidden is true.
	Section string `json:"section,omitempty"`
	// Hidden, when true, drops commits of this Type instead of assigning them
	// to a section.
	Hidden bool `json:"hidden,omitempty"`
}

type ReleaseContext struct {
	Version        string
	CurrentVersion string
	Branch         string
	RepositoryURL  string
	Commits        []string
}

// Contributor represents a person who contributed to a release.
type Contributor struct {
	// Name is the display name (git author name or GitHub login).
	Name string `json:"name"`
	// Login is the optional GitHub/Gitea username (e.g. "alice"). When set,
	// the entry is rendered as @alice; otherwise Name is used as plain text.
	Login string `json:"login,omitempty"`
	// PR is the optional pull-request number associated with this contributor's
	// first contribution in the release. Zero means no PR is available.
	PR int `json:"pr,omitempty"`
}

type GenerateOptions struct {
	Signature           bool
	AIDisclosure        bool
	AIDisclosureBadge   string
	AIDisclosureSection bool
	// NewContributors controls whether a "New Contributors" section is appended.
	// The section is only rendered when Contributors is non-empty.
	NewContributors bool
	// MVP controls whether a "🏆 MVP" section is rendered after the contributor
	// list. Requires NewContributors=true and a non-empty Contributors slice.
	MVP bool
	// MVPMetric determines how the MVP is chosen: "commits" (default) picks the
	// contributor with the highest commit count; "impact" weights breaking/feat
	// commits more heavily.
	MVPMetric string
	// Contributors is the pre-computed list of first-time contributors for this
	// release. It is populated by the caller (e.g. from SEMREL_PLUGIN_CONTRIBUTORS_JSON).
	// When empty the section is silently skipped regardless of NewContributors.
	Contributors []Contributor
	// Sections, when non-empty, overrides the built-in feat/fix→section
	// mapping used to group release-note entries. Types without a matching
	// rule fall back to the "Other" section.
	Sections []SectionRule
}

type Generator struct{}

var conventionalHeaderPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)

func New() *Generator {
	return &Generator{}
}

func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		Signature:           false,
		AIDisclosure:        false,
		AIDisclosureBadge:   "🤖",
		AIDisclosureSection: false,
		NewContributors:     true,
		MVP:                 false,
		MVPMetric:           "commits",
	}
}

func (g *Generator) Generate(ctx ReleaseContext, options ...GenerateOptions) string {
	generateOptions := DefaultGenerateOptions()
	if len(options) > 0 {
		generateOptions = options[0]
	}

	sections := map[string][]string{}
	rendered := 0
	type aiEntry struct {
		header string
		labels []string
	}
	var aiEntries []aiEntry

	for _, commit := range ctx.Commits {
		section, line := classifyCommit(commit, generateOptions)
		if section == "" || line == "" {
			continue
		}
		if generateOptions.AIDisclosure {
			if labels := detectAIAuthors(commit); len(labels) > 0 {
				badge := generateOptions.AIDisclosureBadge
				if badge == "" {
					badge = "🤖"
				}
				line = line + " " + badge
				if generateOptions.AIDisclosureSection {
					aiEntries = append(aiEntries, aiEntry{firstLine(commit), labels})
				}
			}
		}
		sections[section] = append(sections[section], line)
		rendered++
	}

	var builder strings.Builder
	if version := displayVersion(ctx.Version); version != "Unreleased" {
		fmt.Fprintf(&builder, "## %s\n\n", version)
	}
	builder.WriteString("## What's Changed")

	for _, section := range sectionOrder(generateOptions) {
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

	if generateOptions.NewContributors && len(generateOptions.Contributors) > 0 {
		builder.WriteString("\n\n### New Contributors\n")
		for _, c := range generateOptions.Contributors {
			builder.WriteString("* ")
			builder.WriteString(formatContributorEntry(c, ctx.RepositoryURL))
			builder.WriteString(" made their first contribution")
			if c.PR > 0 && ctx.RepositoryURL != "" {
				repoURL := strings.TrimRight(strings.TrimSpace(ctx.RepositoryURL), "/")
				fmt.Fprintf(&builder, " in [#%d](%s/pull/%d)", c.PR, repoURL, c.PR)
			}
			builder.WriteString("\n")
		}

		if generateOptions.MVP {
			mvp := pickMVP(generateOptions.Contributors, ctx.Commits, generateOptions.MVPMetric)
			if mvp != nil {
				builder.WriteString("\n### 🏆 MVP\n")
				fmt.Fprintf(&builder, "%s led the contributors this release.\n", formatContributorEntry(*mvp, ctx.RepositoryURL))
			}
		}
	}

	if generateOptions.AIDisclosure && generateOptions.AIDisclosureSection && len(aiEntries) > 0 {
		builder.WriteString("\n\n<details>\n<summary>🤖 AI-Assisted Contributions</summary>\n\nThe following changes were co-authored with an AI assistant:\n\n")
		for _, e := range aiEntries {
			fmt.Fprintf(&builder, "- %s — Co-authored with **%s**\n", e.header, strings.Join(e.labels, ", "))
		}
		builder.WriteString("\n_Disclosed in accordance with project AI-usage policy (L-08 §8)._\n</details>")
	}

	if generateOptions.Signature {
		builder.WriteString("\n\n---\n*Generated by [semrel.io](https://semrel.io)*")
	}

	return builder.String()
}

func classifyCommit(commit string, options GenerateOptions) (string, string) {
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

	commitType := strings.ToLower(matches[1])

	if rule, ok := findSectionRule(options.Sections, commitType); ok {
		if rule.Hidden {
			return "", ""
		}
		section := strings.TrimSpace(rule.Section)
		if section == "" {
			return "", ""
		}
		return section, line
	}

	if len(options.Sections) > 0 {
		// A custom mapping is configured but doesn't cover this type; bucket it
		// with the other unmapped commits instead of applying the built-in
		// feat/fix defaults.
		return otherChangesSection, line
	}

	switch commitType {
	case "feat":
		return featuresSection, line
	case "fix", "perf", "revert":
		return bugFixesSection, line
	default:
		return otherChangesSection, line
	}
}

// findSectionRule returns the first rule whose Type matches commitType
// case-insensitively.
func findSectionRule(rules []SectionRule, commitType string) (SectionRule, bool) {
	for _, rule := range rules {
		if strings.EqualFold(strings.TrimSpace(rule.Type), commitType) {
			return rule, true
		}
	}
	return SectionRule{}, false
}

// sectionOrder returns the release-note section headings in the order they
// should be rendered. Without a custom Sections mapping this is the built-in
// fixed order. With a custom mapping, each distinct configured section is
// rendered in declaration order, followed by "Other" for any commit types
// the mapping doesn't cover.
func sectionOrder(options GenerateOptions) []string {
	if len(options.Sections) == 0 {
		return []string{featuresSection, bugFixesSection, otherChangesSection}
	}

	var order []string
	seen := map[string]bool{}
	for _, rule := range options.Sections {
		if rule.Hidden {
			continue
		}
		section := strings.TrimSpace(rule.Section)
		if section == "" || seen[section] {
			continue
		}
		seen[section] = true
		order = append(order, section)
	}
	if !seen[otherChangesSection] {
		order = append(order, otherChangesSection)
	}
	return order
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

// formatContributorEntry returns a Markdown-formatted mention for a contributor.
// When Login is set it renders as @login (linked when a repositoryURL is available);
// otherwise it falls back to plain Name.
func formatContributorEntry(c Contributor, repositoryURL string) string {
	login := strings.TrimSpace(c.Login)
	if login == "" {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			return "unknown"
		}
		return name
	}

	repositoryURL = strings.TrimRight(strings.TrimSpace(repositoryURL), "/")
	if repositoryURL == "" {
		return "@" + login
	}

	// Derive the base host URL from the repository URL (e.g. https://github.com).
	// We can't assume a specific host; strip the path to get the root.
	base := repositoryURL
	if idx := strings.Index(repositoryURL, "//"); idx >= 0 {
		rest := repositoryURL[idx+2:]
		if slash := strings.Index(rest, "/"); slash >= 0 {
			base = repositoryURL[:idx+2+slash]
		}
	}
	return fmt.Sprintf("[@%s](%s/%s)", login, base, login)
}

// pullRequestPattern matches "(#123)" style PR references used for MVP scoring.
var pullRequestPattern = regexp.MustCompile(`\(#(\d+)\)`)

func pickMVP(contributors []Contributor, commits []string, mvpMetric string) *Contributor {
	if len(contributors) == 0 {
		return nil
	}
	if len(contributors) == 1 {
		return &contributors[0]
	}

	// Without per-commit author attribution we rank contributors by the number
	// of PR references in commit messages that match their PR number.
	scores := make(map[string]int, len(contributors))
	for _, commit := range commits {
		prs := pullRequestPattern.FindAllStringSubmatch(commit, -1)
		for _, match := range prs {
			prNum := match[1]
			for _, c := range contributors {
				if c.PR > 0 && fmt.Sprintf("%d", c.PR) == prNum {
					key := contributorKey(c)
					if mvpMetric == "impact" {
						scores[key] += impactWeight(commit)
					} else {
						scores[key]++
					}
				}
			}
		}
	}

	// Fall back to first contributor in list (ordering preserved from caller).
	best := &contributors[0]
	bestScore := scores[contributorKey(contributors[0])]
	for i := range contributors[1:] {
		c := &contributors[i+1]
		if s := scores[contributorKey(*c)]; s > bestScore {
			bestScore = s
			best = c
		}
	}
	return best
}

func contributorKey(c Contributor) string {
	if c.Login != "" {
		return c.Login
	}
	return c.Name
}

func impactWeight(commit string) int {
	lower := strings.ToLower(commit)
	if strings.Contains(lower, "breaking change") {
		return 3
	}
	if conventionalHeaderPattern.MatchString(firstLine(commit)) {
		matches := conventionalHeaderPattern.FindStringSubmatch(firstLine(commit))
		if len(matches) > 3 && matches[3] == "!" {
			return 3
		}
		if len(matches) > 1 && matches[1] == "feat" {
			return 2
		}
	}
	return 1
}
