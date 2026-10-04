# Website Copywriting Overhaul: From 2025 Marketing Copy to Current Product Reality

**Date**: March 25, 2026
**Type**: Content | Refactoring
**Components**: Navigation, Features Pages, Solutions Pages, Pricing Page, Design System, UI Components

## Summary

Complete rewrite of all marketing pages on planton.ai to reflect the current product as of March 2026. The website had not been updated since March 2025 while the product evolved through 1,828 changelog entries. Every feature page, solutions page, and navigation element was rewritten with problem-first, visitor-centric copywriting grounded in what the platform actually does today. Old feature labels (Self-Service DevOps, Auditable Intelligence, Planton Copilot) were replaced with real product module names (InfraHub, ServiceHub, Runner, Security, Agent Fleet).

## Problem Statement / Motivation

A friend navigated through every page on planton.ai via the sitemap and reported that all content was completely outdated. The website still described "Plantora" (a deprecated AI chatbot), "Self-Service DevOps" (a vague label that didn't map to any product module), and "Planton Copilot" (disabled in December 2025). Meanwhile, the actual product had grown to include InfraHub with Infra Charts and DAG pipelines, a complete Runner architecture for self-hosted execution, an Agent Fleet with marketplace and skills, a Security module with multi-backend secrets, and 150+ catalog kinds.

### Pain Points

- Feature pages used arbitrary marketing labels ("Self-Service DevOps", "Auditable Intelligence") instead of actual product module names
- "Plantora" / "Planton Copilot" referenced on 20+ component files despite being deprecated since December 2025
- No pages existed for Connect (integrations), Runner (self-hosted execution), Security (secrets/IAM/audit), or OpenMCF (open source)
- Multiple CTA buttons had no `href` attribute
- Solutions page copy mentioned "your startup's growth" on the Enterprises page
- Internal implementation details (Temporal, NATS, mTLS, Konnectivity, Neo4j, OpenFGA, Tekton) leaked into marketing copy throughout
- Inconsistent brand naming: Planton, Planton, Plantora, Planton Aura, Project Planterm

## Solution / What's New

### New Information Architecture

The top-level navigation changed from "Features" (6 old labels) to "Product" (7 real modules):

| Old | New |
|-----|-----|
| Agent Fleet | Agent Fleet |
| Self-Service DevOps | *(removed)* |
| Service Hub | ServiceHub |
| IaC Workflows | InfraHub |
| Kubernetes Dashboard | *(removed)* |
| Auditable Intelligence | *(removed)* |
| *(none)* | Runner |
| *(none)* | Security |
| *(none)* | Open Source |
| *(none)* | CLI |

Solutions menu updated: "ChatOps" use case replaced with "Self-Hosted DevOps", "DevOps" role renamed to "Platform Engineers".

### 7 Product Pages Created or Rewritten

1. **Product Overview** (`/features`) — Hub page with 7 module cards, architecture diagram showing SaaS/Runner split, and trust bar
2. **InfraHub** (`/features/infra-hub`) — Infra Components, Infra Charts, Infra Pipelines, Infra Jobs, Presets, Multi-Provisioner
3. **ServiceHub** (`/features/service-hub`) — Git-to-deploy, multi-env promotion, deploy anywhere, Kustomize config, ingress, pod access
4. **Runner** (`/features/runner`) — Architecture, CloudOps, IaC execution, security model, deployment options, secure tunnel
5. **Security** (`/features/security`) — Secrets management, multi-backend secrets, Runner trust model, identity & access, audit trails, connection security, zero-trust architecture
6. **Agent Fleet** (`/agents`) — Marketplace, skills, sub-agents, MCP integration, testing, sessions & streaming
7. **CLI** (`/cli`) — Manifest-driven, Infra Job operations, connection management, kubernetes access, environment config
8. **Open Source** (`/features/open-source`) — OpenMCF, portable manifests, Infra Charts, Forge workflow

### 11 Solutions Pages Rewritten

- **Hub** (`/solutions`) — Clean grid with 3 use cases, 3 sizes, 4 roles
- **By Use Case**: Internal Developer Platform, Multi-Cloud, Self-Hosted DevOps (new)
- **By Size**: Startups, Growing Teams, Enterprises
- **By Role**: Developers, Platform Engineers (new, was DevOps), Startup Founders, Engineering Leaders

### 7 Redirect Pages

Old routes preserved with client-side redirects:
- `/features/iac-workflows` → `/features/infra-hub`
- `/features/self-service-devops` → `/features`
- `/features/planton-copilot` → `/agents`
- `/features/auditable-intelligence` → `/features`
- `/features/kubernetes-dashboard` → `/features`
- `/solutions/by-use-case/chat-ops` → `/solutions`
- `/solutions/by-role/devops` → `/solutions/by-role/platform-engineers`

### Connect Page Removed

Initially created a Connect page, then removed it after recognizing that credential/integration management is plumbing, not a value proposition. The meaningful security aspects were already covered by the Security page.

### Implementation Details Stripped

Systematic sweep to remove internal technology names from all marketing pages:
- Temporal → "reliable execution", "automated"
- NATS → "real-time streaming"
- mTLS → "encrypted", "secure"
- Konnectivity → removed, described outcome instead
- Neo4j → "dependency graph"
- Tekton → "managed pipelines"
- OpenFGA → "fine-grained access control"
- Certificate CN → "verified identity"
- Port numbers, protocol details → removed entirely

## Implementation Details

### Component Architecture

Each product page follows a consistent 3-component pattern:

```
src/components/product/<module>/
  ├── index.tsx     (barrel export)
  ├── hero.tsx      (badge, title, problem, solution, CTAs)
  ├── capabilities.tsx  (alternating sections with code examples)
  └── cta.tsx       (bottom call-to-action)
```

Solutions pages use a single component per page (simpler structure).

All pages use the v3 design system from `shared.tsx`: Section, SectionTitle, Card, FeatureCard, Badge, PrimaryButton, SecondaryButton.

### Copywriting Framework

Every page follows a consistent arc:
1. **Pain** — Start with the visitor's frustration
2. **Bridge** — Show the contrast
3. **Solution** — How Planton specifically solves this
4. **Proof** — Code examples, CLI output, YAML snippets
5. **Depth** — Capabilities with problem-first framing
6. **Action** — Clear CTA

### Key Decision: Connect Removed

After initial implementation, removed the Connect product page. Rationale: credential management is table stakes, not a differentiator. Visitors expect a platform to connect to their clouds — they don't need a dedicated page selling them on it. The meaningful security aspects (SecretRef, JIT resolution, multi-backend) already live on the Security page.

### Key Decision: No Implementation Details

All internal technology choices (Temporal for orchestration, NATS for messaging, Konnectivity for tunneling, Neo4j for dependency graphs) were stripped from marketing copy. Visitors care about outcomes ("reliable execution", "real-time progress"), not implementation choices. This also protects intellectual property.

## Benefits

- Every marketing page now reflects the actual product as of March 2026
- Visitors can understand what each product module does and how it solves their problem
- Navigation maps directly to real product modules, not arbitrary marketing labels
- Zero deprecated terminology (Plantora, Planton, Planton Copilot, Planton Aura)
- All CTAs have valid hrefs pointing to console signup or demo booking
- Old URLs preserved via redirect pages — no broken links from external sources
- Implementation details protected — no Temporal, NATS, mTLS, or Konnectivity exposure

## Impact

### Pages Affected

- **7 product pages** created or rewritten
- **11 solutions pages** rewritten
- **7 redirect pages** created for deprecated routes
- **Navigation** (header + footer) restructured
- **Features sub-nav** updated
- **1 page removed** (/features/connect)
- **Pricing metadata** updated

### Files Changed

~80 files across:
- `src/components/product/` (new directory with 10 subdirectories)
- `src/components/product/solutions/` (11 solution page components)
- `src/components/layout/header/header.tsx`
- `src/components/layout/footer.tsx`
- `src/app/(root)/features/` (layout + page files)
- `src/app/(root)/solutions/` (page files)
- `src/app/(root)/agents/page.tsx`
- `src/app/(root)/cli/page.tsx`
- `src/components/common/redirect-page.tsx`

## Related Work

- Monochrome theme redesign (same branch) — applied black-and-white visual identity
- Homepage v3 update (January 2026) — already reflected current product
- Product documentation (planton monorepo) — source of truth for all capabilities

---

**Status**: ✅ Live
**Timeline**: Single session
