---
title: "Coding Agents"
description: "Use Planton from Cursor, Claude Code, Codex, and any other coding agent: install the Planton skills, pair them with the CLI, and let your agent compose, validate, and deploy infrastructure from your repository"
icon: terminal
order: 3
tags:
  - Coding Agents
  - Cursor
  - Claude Code
  - Skills
  - Getting Started
---

# Coding Agents

Planton ships its assistant's knowledge as **agent skills**: instruction files in the open [Agent Skills](https://agentskills.io) format that Cursor, Claude Code, Codex, Gemini CLI, and dozens of other coding agents load on demand. Install them once and your agent knows how to compose cloud infrastructure, validate it, wire resources together, register and deploy services, and where the lines are that it never crosses.

## Why install the skills

Without them, a coding agent writes infrastructure from memory: field names it half-remembers, resources it forgets to wire together, no idea what anything costs. With them, the agent reads every schema fact from a reference page at answer time, wires resources by reference instead of pasting IDs, validates before it hands anything over, tells you the monthly cost, and asks before it applies. The same skills power the Planton Assistant in Planton Desktop and the web console, so your agent and the platform's own assistant share one craft.

Two skills ship together:

| Skill | What it carries |
|-------|-----------------|
| `planton` | The working craft: Infra Charts and manifest sets, the compile loop, deployed projects, service registration, push-to-deploy, CI/CD, and the boundaries (no mutation without consent, never outside your repository). |
| `multi-cloud-catalog` | The kind reference pack, shipped inside the skill: one page per catalog kind across every supported provider, the catalog-wide reference graph, and verified fact sheets for cost, security posture, and runner permissions. |

## Install the skills

### One command, any agent

The [skills CLI](https://github.com/vercel-labs/skills) detects the agents on your machine and installs into each one's skills directory, at project or global scope:

```bash
npx skills add plantonhq/skills
```

Target a specific agent, or install globally so every project sees the skills:

```bash
npx skills add plantonhq/skills --agent cursor
npx skills add plantonhq/skills --agent claude-code --global
```

### Claude Code marketplace

Inside Claude Code:

```
/plugin marketplace add plantonhq/skills
/plugin install planton@planton
```

### Manual

Clone [github.com/plantonhq/skills](https://github.com/plantonhq/skills) and copy or symlink `skills/planton` and `skills/multi-cloud-catalog` into your agent's skills directory: `~/.cursor/skills/`, `~/.claude/skills/`, `~/.codex/skills/`, or `.agents/skills/` inside a project. `SKILL.md` must sit directly inside each skill folder.

## Install and sign in to the CLI

The skills do their best work with the `planton` CLI on your PATH. Schema lookups (`planton explain`) and manifest validation (`planton validate`) run fully offline with no account; one sign-in unlocks the compile loop, lookups against your organization, and deploys.

```bash
brew install plantonhq/tap/planton
planton login
planton context set --org your-org -e dev
```

Other platforms: see [CLI Installation](/docs/cli).

## Your first request

Open a repository in your coding agent and ask for what you need in your own words:

> I need a Postgres database for this service in dev.

The agent creates an `infrastructure/` folder at the repository root, writes one manifest for the database grounded in the kind's reference page, validates it with `planton validate`, and replies with what it built, what it costs per month, and the assumptions it made. It does not apply anything until you say so; when you do, it runs `planton apply -f infrastructure/` with one confirmation and narrates the preflight report and the deploy.

Ask for more and the folder grows the same way: several wired resources are applied together as one dependency-ordered set, and a full application platform becomes an Infra Chart in its own subfolder. Everything lands under `infrastructure/`, ready to diff and commit; the agent never touches your application code or files outside the repository.

## Optional: the platform's tools over MCP

Your agent can also reach the platform's own operations directly (build, apply, deploy, read your organization's estate) through the hosted Planton MCP server at `https://mcp.planton.ai/`. This works alongside the CLI or without it.

Create an API key and export it:

```bash
planton api-key new --name cursor --scope infrastructure=read-write
export PLANTON_API_KEY=<the key printed above>
```

Then connect your agent:

**Claude Code**

```
/plugin install planton-platform-tools@planton
```

**Cursor** (`~/.cursor/mcp.json`, or `.cursor/mcp.json` in a project):

```json
{
  "mcpServers": {
    "planton": {
      "type": "http",
      "url": "https://mcp.planton.ai/",
      "headers": { "Authorization": "Bearer ${env:PLANTON_API_KEY}" }
    }
  }
}
```

API keys carry your own permissions, scoped to the product areas you name; see [Authentication and Authorization](/docs/security/authentication-and-authorization).

## Versions and updates

Every commit in [plantonhq/skills](https://github.com/plantonhq/skills) is a Planton release, tagged with the release's tag. The skill files are byte-identical to the checksummed archives that release published, and they are what Planton Desktop and the hosted assistant run at that version. Update with:

```bash
npx skills update
```

In Claude Code, `/plugin` shows available plugin updates.

The skills are authored in [plantonhq/planton](https://github.com/plantonhq/planton) under `skills/`, linted against the Agent Skills specification on every pull request, and copied to the distribution repository by the release lane. Improvements are welcome there.
