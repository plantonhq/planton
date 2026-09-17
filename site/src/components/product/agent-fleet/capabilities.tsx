'use client';

import { Box, Typography } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  FeatureTitle,
  BodyText,
  FeatureCard,
  Grid,
} from '@/components/marketing';
import {
  Storefront as MarketplaceIcon,
  AutoFixHigh as SkillsIcon,
  AccountTree as OrchestrationIcon,
  Cable as McpIcon,
  BugReport as TestingIcon,
  Stream as SessionsIcon,
} from '@mui/icons-material';
import {
  MetricsStrip,
  CodeTabs,
  BentoGrid,
  BentoItem,
  AnimatedTerminal,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { TerminalLine, CodeTab, MetricItem } from '@/components/product/shared';

const metrics: MetricItem[] = [
  { value: '100%', label: 'Real Execution' },
  { value: '3-tier', label: 'Skill Ownership' },
  { value: 'MCP', label: 'Tool Protocol' },
  { value: '100%', label: 'Auditable' },
];

const skillTabs: CodeTab[] = [
  {
    label: 'Skill',
    code: `# Postgres Failover Runbook
#
# When replication lag exceeds 30s:
# 1. Check replica sync status
# 2. Verify data consistency
# 3. Promote replica to primary
# 4. Update connection strings
# 5. Notify on-call channel
#
# Ownership: platform (built-in)
# Scope: organization`,
  },
  {
    label: 'Agent Session',
    code: `▶ Agent: pipeline-debugger
  Skills: ci-pipeline-analysis, resource-scaling
  Tools: kubectl, planton-api, log-search

  Analyzing CI failure on main...
  → Build step 3/5: OOM at 512Mi
  → Peak memory (7d): 480Mi
  → Recommendation: increase to 768Mi`,
  },
];

const sessionLines: TerminalLine[] = [
  { text: '▶ Agent: security-auditor', className: 'text-[#b0b0b0]' },
  { text: '  Session: ses-d83b14', className: 'text-[#666]' },
  { text: '  Trigger: weekly security scan', className: 'text-[#666]' },
  { text: '' },
  { text: '⚙ Tool: list_cloud_resources', className: 'text-white' },
  { text: '  → Scanning 47 resources across 3 providers', className: 'text-[#666]' },
  { text: '' },
  { text: '⚙ Tool: check_encryption_at_rest', className: 'text-white' },
  { text: '  → 2 S3 buckets missing encryption', className: 'text-[#f59e0b]' },
  { text: '  → staging-uploads, dev-artifacts', className: 'text-[#666]' },
  { text: '' },
  { text: '⚙ Tool: check_public_access', className: 'text-white' },
  { text: '  → All resources private ✓', className: 'text-[#10b981]' },
  { text: '' },
  { text: '📋 Report: 2 findings, 0 critical', className: 'text-[#10b981]' },
];

const SkillsShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <SkillsIcon />
            </Box>
            <FeatureTitle>Skills</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Agents learn your organization&apos;s patterns. Create skills that encode your
            operational knowledge into repeatable, automatable procedures.
          </BodyText>
          <Box className="space-y-3">
            {[
              'Encode runbooks, naming conventions, and compliance requirements as skills',
              'Your practices, not generic best practices — tailored to your stack',
              'Version-controlled skill definitions alongside your infrastructure code',
              'Composable skills that agents combine for complex scenarios',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <CodeTabs tabs={skillTabs} title="Agent Skills" />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const FeatureBentoSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Purpose-built agents for every infrastructure challenge</SectionTitle>
        <SectionSubtitle className="mx-auto">
          From marketplace discovery to production orchestration &mdash; Agent Fleet gives your team AI-powered infrastructure operations.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <MarketplaceIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Marketplace</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Browse and deploy agents for specific tasks. Pipeline troubleshooting, database management, security hardening, drift detection.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['pipeline-debugger', 'security-auditor', 'drift-detector', 'db-optimizer'].map((agent) => (
                <Box key={agent} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {agent}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <OrchestrationIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Sub-Agents &amp; Orchestration</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Complex tasks broken into specialized sub-tasks. A deployment agent invokes security, testing, and notification agents automatically.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['delegate', 'retry', 'context-pass', 'approval-gate'].map((cap) => (
                <Box key={cap} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {cap}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem span="wide">
            <Box className="flex flex-col md:flex-row gap-5">
              <Box className="flex-1">
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <McpIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">MCP Integration</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Agents connect through Model Context Protocol. Real tool execution against your infrastructure, not simulated responses.
                </BodyText>
              </Box>
              <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                <pre className="whitespace-pre-wrap">{`⚙ Tool: kubectl_get_pods
  namespace: backend
  cluster: prod-us-east-1

→ 3 pods running, 0 pending
→ No restarts in 24h`}</pre>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <TestingIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Testing</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Test suites for agent validation. Verify behavior before deploying to production with structured test scenarios.
            </BodyText>
            <Box className="flex items-center gap-3">
              <Box className="flex-1 p-3 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-[#10b981] mb-0.5">12 passed</Typography>
                <Typography className="text-[10px] text-[#666]">Test suite</Typography>
              </Box>
              <Box className="flex-1 p-3 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white mb-0.5">0 failed</Typography>
                <Typography className="text-[10px] text-[#666]">Regressions</Typography>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const SessionsDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Full visibility into every agent action</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Real-time streaming of tool calls, reasoning steps, and results. Every action is auditable and replayable.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={sessionLines}
        title="agent session — ses-7f2a9c"
        lineDelay={400}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<SessionsIcon />}
            title="Live Streaming"
            description="Every tool call, decision, and result streamed in real time to the console."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<TestingIcon />}
            title="Session Replay"
            description="Replay any session for debugging and post-incident review. Full timeline preserved."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<MarketplaceIcon />}
            title="Exportable Logs"
            description="Export session logs for compliance reporting and external audit trail integration."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

export const AgentFleetCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <SkillsShowcase />
    <FeatureBentoSection />
    <SessionsDeepDive />
    {/* SCREENSHOT OPPORTUNITY: Agent Fleet — Agent Session Chat Interface
       Show: The console's agent chat interface during a live session with:
       - The chat thread showing agent messages, tool call results, and recommendations
       - The sidebar showing session metadata (agent name, trigger, duration, tool calls count)
       - Ideally a canvas/flow view showing agent → sub-agent orchestration
       Value: This is the single most impactful screenshot for the Agent Fleet page — it proves the product exists and is polished.
       Suggested: 16:9 aspect ratio, dark theme. Use a realistic scenario like pipeline debugging or security audit.
       Placement: Add an <img> tag wrapped in a <Section> component between SessionsDeepDive and CTA.
    */}
    {/* SCREENSHOT OPPORTUNITY: Agent Fleet — Marketplace Browse View
       Show: The agent marketplace page with 6-8 agent cards showing:
       - Agent name, description, skills count, and category tags
       - A search/filter bar at the top
       Value: Demonstrates the marketplace is populated and the agent ecosystem is real, not vaporware.
       Suggested: 16:9 aspect ratio, dark theme.
       Placement: Could replace or accompany the Marketplace bento card in FeatureBentoSection.
    */}
  </>
);
