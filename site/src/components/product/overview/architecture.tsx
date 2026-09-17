'use client';

import { Box, Typography } from '@mui/material';
import { Section, SectionTitle, SectionSubtitle, BodyText, Card } from '@/components/marketing';

const ArchitectureLayer = ({
  label,
  items,
  accent = false,
}: {
  label: string;
  items: string[];
  accent?: boolean;
}) => (
  <Box
    className={`
      rounded-lg border p-4 md:p-5
      ${accent ? 'border-white/20 bg-white/5' : 'border-[#2a2a2a] bg-[#151515]'}
    `}
  >
    <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-3">
      {label}
    </Typography>
    <Box className="flex flex-wrap gap-2">
      {items.map((item) => (
        <Box
          key={item}
          className="px-3 py-1.5 rounded-md bg-white/5 border border-[#2a2a2a] text-sm text-[#b0b0b0]"
        >
          {item}
        </Box>
      ))}
    </Box>
  </Box>
);

export const ProductArchitecture = () => {
  return (
    <Section>
      <Box className="text-center mb-10">
        <SectionTitle>How it all fits together</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Planton separates orchestration (SaaS) from execution (your cloud). You get the
          convenience of a managed platform with the security of self-hosted infrastructure.
        </SectionSubtitle>
      </Box>

      <Card hover={false} className="!p-6 md:!p-8 max-w-4xl mx-auto">
        <Box className="space-y-4">
          <ArchitectureLayer
            label="Your developers"
            items={['Console UI', 'CLI', 'CI/CD Pipelines', 'Agent Fleet']}
          />

          <Box className="flex justify-center">
            <Box className="w-px h-6 bg-[#2a2a2a]" />
          </Box>

          <ArchitectureLayer
            label="Planton control plane (SaaS)"
            items={['Infra Hub', 'Service Hub', 'Security', 'Orchestration']}
            accent
          />

          <Box className="flex justify-center">
            <Box className="flex flex-col items-center">
              <Box className="w-px h-3 bg-[#2a2a2a]" />
              <Box className="px-3 py-1 rounded-full border border-[#2a2a2a] bg-[#151515] text-xs text-[#666]">
                Encrypted Tunnel
              </Box>
              <Box className="w-px h-3 bg-[#2a2a2a]" />
            </Box>
          </Box>

          <ArchitectureLayer
            label="Your cloud (Runner)"
            items={['IaC Execution', 'CloudOps', 'Provider APIs', 'Your Resources']}
          />
        </Box>

        <Box className="mt-6 flex flex-col md:flex-row gap-4">
          <Box className="flex-1 p-4 rounded-lg bg-white/5 border border-[#2a2a2a]">
            <BodyText className="!text-[#666]">
              <strong className="text-[#b0b0b0]">Credentials stay in your cloud.</strong>{' '}
              Runner authenticates at execution time using your cloud provider&apos;s native IAM.
              Those credentials never cross into the control plane.
            </BodyText>
          </Box>
          <Box className="flex-1 p-4 rounded-lg bg-white/5 border border-[#2a2a2a]">
            <BodyText className="!text-[#666]">
              <strong className="text-[#b0b0b0]">Open source foundation.</strong>{' '}
              Every resource definition is a KRM YAML manifest powered by Planton open source.
              Export and run anywhere with the standalone CLI.
            </BodyText>
          </Box>
        </Box>
      </Card>
      {/* SCREENSHOT OPPORTUNITY: Product Overview — Console Dashboard
         Show: The main Planton console dashboard after login showing:
         - The organization sidebar with modules (InfraHub, ServiceHub, Connect, etc.)
         - A summary view with recent deployments, active resources count, and runner status
         - The onboarding wizard for new users (if possible, show a partially completed state)
         Value: The overview page is the first product page visitors see. A real console screenshot
         proves this is a shipped, polished product — not a landing page for vaporware.
         Suggested: 16:9 aspect ratio, dark theme, with realistic org name and 3-4 recent activities.
         Placement: Add an <img> tag wrapped in a <Section> after this architecture card and before the CTA.
      */}
    </Section>
  );
};
