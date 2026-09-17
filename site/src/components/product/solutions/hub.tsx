'use client';

import Link from 'next/link';
import { Box } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  BodyText,
  FeatureTitle,
  Card,
  Badge,
  Grid,
  ArrowRightIcon,
} from '@/components/marketing';

interface SolutionCardProps {
  title: string;
  problem: string;
  href: string;
}

const SolutionCard = ({ title, problem, href }: SolutionCardProps) => (
  <Link href={href} className="no-underline">
    <Card className="h-full group cursor-pointer">
      <FeatureTitle className="mb-2">{title}</FeatureTitle>
      <BodyText className="mb-4">{problem}</BodyText>
      <Box className="flex items-center text-sm text-[#a0a0a0] group-hover:text-white transition-colors">
        Learn more <ArrowRightIcon />
      </Box>
    </Card>
  </Link>
);

const useCases: SolutionCardProps[] = [
  {
    title: 'Internal Developer Platform',
    problem: 'Stop building your IDP from scratch — get self-service infrastructure, CI/CD, and access control out of the box.',
    href: '/solutions/by-use-case/internal-developer-platform',
  },
  {
    title: 'Multi-Cloud',
    problem: 'One workflow for AWS, GCP, and Azure — without lowest-common-denominator abstraction.',
    href: '/solutions/by-use-case/multi-cloud',
  },
  {
    title: 'Self-Hosted DevOps',
    problem: 'Enterprise security with SaaS convenience — your credentials never leave your cloud.',
    href: '/solutions/by-use-case/self-hosted-devops',
  },
];

const bySize: SolutionCardProps[] = [
  {
    title: 'Startups',
    problem: 'Ship infrastructure from day one without growing your ops team.',
    href: '/solutions/by-size/startups',
  },
  {
    title: 'Growing Teams',
    problem: 'Scale infrastructure processes without scaling ops headcount.',
    href: '/solutions/by-size/growing-teams',
  },
  {
    title: 'Enterprises',
    problem: 'Standardize across teams, clouds, and compliance requirements.',
    href: '/solutions/by-size/enterprises',
  },
];

const byRole: SolutionCardProps[] = [
  {
    title: 'Developers',
    problem: 'Provision infrastructure and deploy code without waiting for ops.',
    href: '/solutions/by-role/developers',
  },
  {
    title: 'Platform Engineers',
    problem: 'Build golden paths and guardrails instead of fielding tickets.',
    href: '/solutions/by-role/platform-engineers',
  },
  {
    title: 'Startup Founders',
    problem: 'Production-grade infrastructure without growing your ops team.',
    href: '/solutions/by-role/startup-founders',
  },
  {
    title: 'Engineering Leaders',
    problem: 'Visibility into infrastructure costs, velocity, and compliance.',
    href: '/solutions/by-role/engineering-leader',
  },
];

export const SolutionsHub = () => {
  return (
    <>
      {/* Hero */}
      <Section className="pt-24 md:pt-32">
        <Box className="text-center max-w-3xl mx-auto">
          <Badge className="mb-6">Solutions</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Infrastructure that works for your team
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-2xl">
            Whether you&apos;re a startup shipping your first service or an enterprise standardizing
            across clouds, Planton meets you where you are.
          </SectionSubtitle>
        </Box>
      </Section>

      {/* By Use Case */}
      <Section>
        <Box className="mb-8">
          <SectionTitle className="mb-2">By use case</SectionTitle>
          <SectionSubtitle>Common infrastructure challenges and how Planton solves them.</SectionSubtitle>
        </Box>
        <Grid cols={3}>
          {useCases.map((card) => (
            <SolutionCard key={card.title} {...card} />
          ))}
        </Grid>
      </Section>

      {/* By Size */}
      <Section>
        <Box className="mb-8">
          <SectionTitle className="mb-2">By team size</SectionTitle>
          <SectionSubtitle>Right-sized infrastructure workflows for every stage of growth.</SectionSubtitle>
        </Box>
        <Grid cols={3}>
          {bySize.map((card) => (
            <SolutionCard key={card.title} {...card} />
          ))}
        </Grid>
      </Section>

      {/* By Role */}
      <Section>
        <Box className="mb-8">
          <SectionTitle className="mb-2">By role</SectionTitle>
          <SectionSubtitle>How Planton helps each member of the engineering organization.</SectionSubtitle>
        </Box>
        <Grid cols={4}>
          {byRole.map((card) => (
            <SolutionCard key={card.title} {...card} />
          ))}
        </Grid>
      </Section>
    </>
  );
};
