'use client';

import Link from 'next/link';
import { Box } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  PrimaryButton,
  SecondaryButton,
  Badge,
  BodyText,
} from '@/components/marketing';
import { AnimatedTerminal, ScrollReveal } from '@/components/product/shared';
import type { TerminalLine } from '@/components/product/shared';

const heroLines: TerminalLine[] = [
  { text: '▶ Agent: pipeline-debugger', className: 'text-[#b0b0b0]' },
  { text: '  Trigger: CI failure on main', className: 'text-[#666]' },
  { text: '' },
  { text: '⚙ Tool: fetch_pipeline_logs', className: 'text-white' },
  { text: '  → Build step 3/5 failed: OOM at 512Mi', className: 'text-[#f59e0b]' },
  { text: '' },
  { text: '⚙ Tool: get_resource_spec', className: 'text-white' },
  { text: '  → Current memory limit: 512Mi', className: 'text-[#666]' },
  { text: '  → Peak usage (7d): 480Mi', className: 'text-[#666]' },
  { text: '' },
  { text: '💡 Recommendation:', className: 'text-[#10b981]' },
  { text: '  Increase memory limit to 768Mi', className: 'text-[#10b981]' },
  { text: '  → planton apply -f updated-spec.yaml', className: 'text-[#666]' },
];

export const AgentFleetHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Agent Fleet</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            AI agents, purpose-built for your infrastructure
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Your DevOps team is firefighting, not building. Pipeline failures at 2&nbsp;AM.
            Database incidents requiring tribal knowledge. Security audits that take weeks
            of manual review.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Agent Fleet: Specialized AI agents that understand your infrastructure.
            Not generic chatbots &mdash; purpose-built agents with real access to your
            Planton resources, trained on your specific stack.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Explore Agent Fleet</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="agent session"
            lineDelay={350}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Box>
    </Section>
  );
};
