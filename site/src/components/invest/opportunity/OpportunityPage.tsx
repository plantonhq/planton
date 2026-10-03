'use client';

import { FC } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import {
  Section,
  Container,
  SectionTitle,
  SectionSubtitle,
  BodyText,
  SmallText,
  Card,
  CardTitle,
  CardText,
  Callout,
  Badge,
  Grid,
  List,
  Table,
} from '../explainer/shared';
import { Hero, Footer } from '../explainer/layout';

// ============================================================================
// EXTERNAL LINKS
// ============================================================================

const YC_URL = 'https://www.ycombinator.com/';

// ============================================================================
// ALTERNATIVE DATA STRUCTURE
// ============================================================================

interface Alternative {
  name: string;
  url: string;
  logo: string;
  tagline: string;
  funding: string;
  fundingSource: string;
  founded: string;
  ycBatch?: string; // e.g., "S20", "W22"
  positioning: string;
  whatWeRespect: string;
  wherePlantonDiffers: string;
}

// Alternatives in the market: Porter.run, Qovery.com, MassDriver, Flightcontrol.dev, Harness
// Future context: Railway, Render (managed infrastructure providers)
const ALTERNATIVES: Alternative[] = [
  {
    name: 'Harness',
    url: 'https://harness.io',
    logo: 'harness.png',
    tagline: 'Enterprise CI/CD Platform',
    funding: '$400M+ raised',
    fundingSource: 'https://www.crunchbase.com/organization/harness-io',
    founded: '2017',
    positioning: 'Full enterprise DevOps platform targeting Fortune 500. Comprehensive CI/CD, feature flags, cloud cost management. Enterprise pricing starts at $100K+/year.',
    whatWeRespect: 'Proven enterprise success with major companies. Comprehensive feature set covering full DevOps lifecycle. Strong engineering team.',
    wherePlantonDiffers: 'Mid-market focus with transparent pricing. Truly multi-cloud (not just K8s). Self-host ready from day one. 100% open-source IaC modules.',
  },
  {
    name: 'Porter',
    url: 'https://porter.run',
    logo: 'porter-dot-run.png',
    tagline: 'Kubernetes-only PaaS',
    funding: '$20M Series A (Jan 2026)',
    fundingSource: 'https://www.porter.run/blog/effortless-app-infrastructure-in-any-cloud-porters-20m-series-a',
    founded: '2020',
    ycBatch: 'S20',
    positioning: 'Deploy applications on your own cloud with Kubernetes abstraction. Markets as "multi-cloud" but only deploys Kubernetes clusters—no other infra components.',
    whatWeRespect: 'YC-backed, strong Kubernetes abstraction, excellent for teams transitioning from Heroku to K8s. Recent $20M raise validates market demand.',
    wherePlantonDiffers: 'Truly multi-cloud: deploy any infrastructure (storage, queues, databases), not just K8s. Full transparency with open-source Pulumi modules. Porter shows no deployment visibility.',
  },
  {
    name: 'Qovery',
    url: 'https://qovery.com',
    logo: 'qovery.png',
    tagline: 'Kubernetes-only IDP',
    funding: '~$17M total ($4M 2022, $13M Sep 2025)',
    fundingSource: 'https://tech.eu/2025/09/30/qovery-raises-13m-to-redefine-devops-automation/',
    founded: '2019',
    // Techstars Paris, not YC
    positioning: 'Internal Developer Platform on AWS/GCP/Azure. 10,000+ developers across 120+ countries. Markets as multi-cloud but limited to Kubernetes deployments only.',
    whatWeRespect: 'Large developer community, G2 Momentum Leader, enterprise-ready. Recent $13M raise shows strong market validation.',
    wherePlantonDiffers: 'Truly multi-cloud: deploy storage buckets, message queues, databases—anything beyond K8s. Open-source Pulumi modules provide full deployment transparency. Qovery only shows limited Terraform runs.',
  },
  {
    name: 'MassDriver',
    url: 'https://massdriver.cloud',
    logo: 'massdriver.png',
    tagline: 'Infrastructure Platform',
    funding: '~$12M raised',
    fundingSource: 'https://www.crunchbase.com/organization/massdriver',
    founded: '2021',
    positioning: 'Visual infrastructure platform for deploying infra components. Drag-and-drop interface for infrastructure.',
    whatWeRespect: 'Visual approach to infrastructure makes it accessible. Strong focus on developer experience.',
    wherePlantonDiffers: 'Service Hub: Vercel-like experience for backend in the customer\'s own cloud. Self-host ready from day one. 100% open-source Pulumi modules.',
  },
  {
    name: 'Flightcontrol',
    url: 'https://flightcontrol.dev',
    logo: 'flight-control.png',
    tagline: 'AWS-only PaaS',
    funding: 'Undisclosed seed (Apr 2022)',
    fundingSource: 'https://www.crunchbase.com/organization/flightcontrol',
    founded: '2021',
    ycBatch: 'W22',
    positioning: 'Deploy to your own AWS account with Heroku-like simplicity. Manages $3M+ AWS resources for customers. AWS-only.',
    whatWeRespect: 'AWS-native focus, bootstrapper mentality from founder. Deploys directly to customer AWS accounts, no vendor lock-in.',
    wherePlantonDiffers: 'True multi-cloud (AWS, GCP, Azure, Kubernetes). Open-source modules provide full transparency into every deployment.',
  },
];

// Future infrastructure competitors for context
const INFRASTRUCTURE_CONTEXT = [
  { name: 'Railway', description: 'Instant deployments with zero config' },
  { name: 'Render', description: 'Unified cloud for apps and websites' },
];

// ============================================================================
// MARKET SIZE DATA
// ============================================================================

interface MarketData {
  category: string;
  size2024: string;
  size2030: string;
  cagr: string;
  source: string;
}

const MARKET_DATA: MarketData[] = [
  {
    category: 'DevOps Platform Market',
    size2024: '$10B',
    size2030: '$25B+',
    cagr: '~16%',
    source: 'Gartner/IDC estimates',
  },
  {
    category: 'Cloud PaaS',
    size2024: '$172B',
    size2030: '$208B (2025)',
    cagr: '21.6%',
    source: 'Gartner Nov 2024',
  },
  {
    category: 'Internal Developer Platforms',
    size2024: '$2.1B',
    size2030: '$15.2B (2033)',
    cagr: '22.4%',
    source: 'Dataintelo IDP Report',
  },
];

// ============================================================================
// MOMENTUM DATA (What we did, what's next)
// ============================================================================

interface MomentumItem {
  label: string;
  description: string;
}

const LAST_MONTH: MomentumItem[] = [
  { label: 'Paying customers', description: 'Active revenue from IT consulting firms in India' },
  { label: 'US pipeline', description: 'Demos with 20+ year American consulting companies (Indiana, San Jose)' },
  { label: 'Investor infrastructure', description: 'Carta + Mercury setup, SAFE documents ready' },
];

const NEXT_MONTH: MomentumItem[] = [
  { label: 'Customer onboarding', description: 'Converting US pipeline to signed contracts' },
  { label: 'Self-hosting hardening', description: 'Enterprise security requirements' },
  { label: 'SOC 2 prep', description: 'Investing in compliance infrastructure' },
];

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export const OpportunityPage: FC = () => {
  return (
    <div className="min-h-screen bg-[#0a0a0a] pt-16">

      <main>
        {/* Hero Section */}
        <Hero
          title="The Opportunity"
          subtitle="Where Planton sits in the market, who we're compared to, and why we believe this is worth building."
        />

        {/* Challenge Statement */}
        <Section>
          <Container>
            <Card variant="highlight" className="border border-[#3a3a3a]">
              <CardTitle color="#ededed" className="text-2xl mb-4">
                We Challenge You to Find Another
              </CardTitle>
              <BodyText className="text-lg mb-6">
                Planton is the <strong>only platform</strong> that provides truly multi-cloud
                deployments—not just Kubernetes, but <em>any</em> infra component across AWS,
                GCP, and Azure—backed by 100% open-source infrastructure as code.
              </BodyText>
              <List
                items={[
                  { icon: '→', iconColor: 'pink', text: 'Truly multi-cloud: deploy storage, databases, queues, serverless—not just K8s' },
                  { icon: '→', iconColor: 'pink', text: <><a href="https://planton.dev/" target="_blank" rel="noopener noreferrer" className="text-white hover:underline">Planton open source</a> backbone: every single line of infrastructure as code is open source</> },
                  { icon: '→', iconColor: 'pink', text: 'Customizable modules: retain self-service benefits while extending functionality' },
                  { icon: '→', iconColor: 'pink', text: 'Full SDLC lifecycle: beyond infrastructure into CI/CD with Service Hub\'s Vercel-like experience' },
                  { icon: '→', iconColor: 'pink', text: 'Enterprise-grade auditability: built into the product from day one' },
                ]}
              />
            </Card>

            {/* The Vision - Postman Parallel */}
            <Card className="mt-6">
              <CardTitle className="mb-3">The Vision</CardTitle>
              <BodyText className="mb-4">
                Postman made API development effortless and became India&apos;s first globally loved developer tool—
                30 million users, 98% of Fortune 500 companies, built from Bangalore.
              </BodyText>
              <BodyText className="mb-4">
                <strong>Planton will make DevOps effortless.</strong> That&apos;s the mission. DevOps is a $10B+
                market growing 16% annually. Even capturing a small slice means building something meaningful.
              </BodyText>
              <SmallText>
                <a
                  href="https://donepudi.me/blog/planton-will-become-the-next-postman/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-white hover:underline"
                >
                  Read the full declaration →
                </a>
              </SmallText>
            </Card>
          </Container>
        </Section>

        {/* Market Context */}
        <Section>
          <Container>
            <SectionTitle gradient>The Platform Engineering Wave</SectionTitle>
            <SectionSubtitle className="mb-8">
              Every company is becoming a software company. DevOps is becoming platform engineering.
              The market is growing, and we&apos;re positioned to capture a meaningful piece.
            </SectionSubtitle>

            <Callout variant="highlight" icon="📈" title="Market Tailwinds" className="mb-8">
              <BodyText>
                Gartner predicts that by 2026, <strong>80% of software engineering organizations</strong> will
                establish platform teams as internal providers of reusable services, components, and tools.
                We&apos;re building for this shift.
              </BodyText>
            </Callout>

            {/* Market Segmentation */}
            <SectionTitle className="text-xl mb-6">Where Planton Fits</SectionTitle>
            <Grid cols={3} gap="md" className="mb-8">
              {/* Segment 1: Startups */}
              <Card>
                <Badge variant="warning" className="mb-3">Startups</Badge>
                <CardTitle className="text-lg mb-2">No Cloud Complexity</CardTitle>
                <BodyText className="mb-3 text-sm">
                  Railway, Render, Fly.io, Heroku
                </BodyText>
                <div className="space-y-2 text-sm">
                  <div className="text-[#a0a0a0]">
                    <strong className="text-[#a0a0a0]">Trade-off:</strong> Give up control for simplicity
                  </div>
                  <div className="text-[#a0a0a0]">
                    <strong className="text-[#ef4444]/70">Not viable for:</strong> Compliance, data sovereignty, custom infra
                  </div>
                </div>
              </Card>

              {/* Segment 2: Enterprise */}
              <Card>
                <Badge className="mb-3">Enterprise</Badge>
                <CardTitle className="text-lg mb-2">1000+ Employees</CardTitle>
                <BodyText className="mb-3 text-sm">
                  Harness, Cloudbees, Spacelift
                </BodyText>
                <div className="space-y-2 text-sm">
                  <div className="text-[#a0a0a0]">
                    <strong className="text-white">Trade-off:</strong> Expensive, long sales cycles, complex
                  </div>
                  <div className="text-[#a0a0a0]">
                    <strong className="text-[#ef4444]/70">Still lacks:</strong> True multi-cloud beyond K8s
                  </div>
                </div>
              </Card>

              {/* Segment 3: Mid-Market (Planton's Target) */}
              <Card>
                <Badge className="mb-3">Mid-Market Gap</Badge>
                <CardTitle className="text-lg mb-2">50-500 Employees</CardTitle>
                <BodyText className="mb-3 text-sm text-[#a0a0a0]">
                  ← Planton&apos;s Target
                </BodyText>
                <div className="space-y-2 text-sm">
                  <div className="text-[#a0a0a0]">
                    <strong className="text-white">Problem:</strong> Need own cloud accounts, can&apos;t use PaaS, can&apos;t afford enterprise
                  </div>
                  <div className="text-[#a0a0a0]">
                    <strong className="text-[#10b981]">Planton solves:</strong> Self-service platform in your cloud, affordable
                  </div>
                </div>
              </Card>
            </Grid>

            {/* Market Size Table - with placeholder data */}
            <Card className="mb-8">
              <CardTitle className="mb-4">Market Size Estimates</CardTitle>
              <SmallText className="mb-4">
                These figures are directional. Exact market sizing depends on how you define the category.
              </SmallText>
              <Table
                columns={[
                  { header: 'Category', accessor: 'category' },
                  { header: '2024', accessor: 'size2024' },
                  { header: '2030 (Est.)', accessor: 'size2030' },
                  { header: 'CAGR', accessor: 'cagr' },
                ]}
                data={MARKET_DATA.map((m) => ({
                  category: m.category,
                  size2024: m.size2024,
                  size2030: m.size2030,
                  cagr: <span className="text-[#10b981]">{m.cagr}</span>,
                }))}
              />
            </Card>
          </Container>
        </Section>

        {/* Alternatives in the Market */}
        <Section background="gradient-subtle">
          <Container>
            <SectionTitle gradient>Alternatives in the Market</SectionTitle>
            <SectionSubtitle className="mb-4">
              Companies solving similar problems, all with different approaches.
              We have immense respect for everyone building in this space.
            </SectionSubtitle>

            <Callout className="mb-8">
              <BodyText>
                <strong>A note on comparisons:</strong> These companies are doing incredible work.
                We&apos;re not here to diminish their achievements—we&apos;re here to provide
                perspective on where we fit in this evolving landscape.
              </BodyText>
            </Callout>

            {/* Alternative Cards */}
            <div className="space-y-6 mb-8">
              {ALTERNATIVES.map((alternative) => (
                <Card key={alternative.name}>
                  <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-4">
                    <div>
                      <div className="flex items-center gap-3 mb-2">
                        {/* Logo */}
                        <div className="w-10 h-10 rounded-lg bg-[#1a1a1a] flex items-center justify-center p-1.5 flex-shrink-0">
                          <Image
                            src={`/_site/images/alternatives/${alternative.logo}`}
                            alt={alternative.name}
                            width={32}
                            height={32}
                            className={`object-contain ${alternative.name === 'Qovery' ? 'bg-white rounded p-0.5' : ''}`}
                          />
                        </div>
                        <CardTitle>{alternative.name}</CardTitle>
                        <Badge variant="purple">{alternative.tagline}</Badge>
                        {alternative.ycBatch && (
                          <a
                            href={YC_URL}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-[#2a2a2a] border border-[#3a3a3a] hover:border-white transition-colors duration-300"
                          >
                            <Image
                              src="/_site/images/logos/yc.svg"
                              alt="Y Combinator"
                              width={16}
                              height={16}
                              className="rounded-sm"
                            />
                            <span className="text-xs font-medium text-[#a0a0a0]">{alternative.ycBatch}</span>
                          </a>
                        )}
                      </div>
                      <SmallText>
                        <a
                          href={alternative.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-white hover:underline"
                        >
                          {alternative.url} ↗
                        </a>
                        {alternative.founded && (
                          <span className="text-[#666] ml-2">• Founded {alternative.founded}</span>
                        )}
                      </SmallText>
                    </div>
                    <div className="text-right">
                      <div className="text-lg font-semibold text-white">{alternative.funding}</div>
                      {alternative.fundingSource && !alternative.fundingSource.includes('PLACEHOLDER') && (
                        <SmallText>
                          <a
                            href={alternative.fundingSource}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-white hover:underline"
                          >
                            Source ↗
                          </a>
                        </SmallText>
                      )}
                    </div>
                  </div>

                  <BodyText className="mb-4">{alternative.positioning}</BodyText>

                  <Grid cols={2} gap="sm">
                    <div className="p-3 rounded-lg bg-[#10b981]/10 border border-[#10b981]/20">
                      <SmallText className="text-[#10b981] font-medium mb-1">What We Respect</SmallText>
                      <CardText>{alternative.whatWeRespect}</CardText>
                    </div>
                    <div className="p-3 rounded-lg bg-[#1a1a1a] border border-[#2a2a2a]">
                      <SmallText className="text-white font-medium mb-1">Where Planton Differs</SmallText>
                      <CardText>{alternative.wherePlantonDiffers}</CardText>
                    </div>
                  </Grid>
                </Card>
              ))}
            </div>

            {/* Future Infrastructure Context */}
            <Card variant="warning">
              <CardTitle color="#a0a0a0" className="mb-3">
                Future Competitive Context
              </CardTitle>
              <BodyText className="mb-4">
                As we grow, we&apos;ll also compete with managed infrastructure providers:
              </BodyText>
              <div className="flex flex-wrap gap-3">
                {INFRASTRUCTURE_CONTEXT.map((item) => (
                  <div
                    key={item.name}
                    className="px-3 py-2 rounded-lg bg-[#151515] border border-[#2a2a2a]"
                  >
                    <span className="font-medium text-white">{item.name}</span>
                    <span className="text-[#666]"> — {item.description}</span>
                  </div>
                ))}
              </div>
            </Card>
          </Container>
        </Section>

        {/* Why Planton */}
        <Section>
          <Container>
            <SectionTitle gradient>Why Planton</SectionTitle>
            <SectionSubtitle className="mb-8">
              Our positioning in this landscape—not a roast, just our perspective.
            </SectionSubtitle>

            <Card variant="highlight" className="mb-8">
              <CardTitle color="#ededed" className="mb-4">
                Our Thesis
              </CardTitle>
              <BodyText>
                {/* PLACEHOLDER: Replace with your actual thesis */}
                Platform engineering shouldn&apos;t require a dedicated team of 5+ engineers.
                DevOps automation should be accessible to every company, not just well-funded startups
                and enterprises. We&apos;re building the platform that makes this possible.
              </BodyText>
            </Card>

            <Grid cols={2} className="mb-8">
              <Card variant="success">
                <CardTitle color="#10b981" className="mb-3">What We&apos;ve Built</CardTitle>
                <List
                  items={[
                    { icon: '✓', iconColor: 'emerald', text: 'Production-ready platform with paying customers' },
                    { icon: '✓', iconColor: 'emerald', text: '100% multi-cloud: deploy any infrastructure, not just K8s' },
                    { icon: '✓', iconColor: 'emerald', text: 'Built-in CI/CD: Service Hub\'s Vercel-like experience for backend' },
                    { icon: '✓', iconColor: 'emerald', text: 'Self-host ready from day one for enterprise' },
                  ]}
                />
              </Card>

              <Card variant="cyan">
                <CardTitle color="#ededed" className="mb-3">Key Differentiators</CardTitle>
                <List
                  items={[
                    { icon: '→', iconColor: 'cyan', text: 'Open-source Pulumi modules — full deployment transparency' },
                    { icon: '→', iconColor: 'cyan', text: 'Deploy storage, queues, databases — not just Kubernetes' },
                    { icon: '→', iconColor: 'cyan', text: 'Alternatives show nothing; we show everything' },
                    { icon: '→', iconColor: 'cyan', text: '3-person team building what 20+ engineers struggle to' },
                  ]}
                />
              </Card>
            </Grid>

            <Callout icon="💡" title="The Underdog Advantage">
              <BodyText>
                We&apos;re lean by design, not by limitation. Lower burn rate means longer runway.
                Longer runway means more time to find product-market fit. We don&apos;t need to
                raise $50M to prove this works.
              </BodyText>
            </Callout>
          </Container>
        </Section>

        {/* Our Niche - Target Market */}
        <Section background="gradient-subtle">
          <Container>
            <SectionTitle gradient>Our Niche</SectionTitle>
            <SectionSubtitle className="mb-8">
              We found our initial customers in a specific segment that traditional solutions can&apos;t serve well.
            </SectionSubtitle>

            <Grid cols={2} className="mb-8">
              <Card variant="highlight">
                <CardTitle color="#ededed" className="mb-4">
                  IT Consulting Companies
                </CardTitle>
                <BodyText className="mb-4">
                  Our primary early adopters are IT consulting companies serving mid-market clients
                  (50-500 employees). These firms have a unique problem:
                </BodyText>
                <List
                  items={[
                    { icon: '→', iconColor: 'pink', text: 'They deploy to client cloud environments, not their own PaaS' },
                    { icon: '→', iconColor: 'pink', text: 'Clients require compliance and data sovereignty' },
                    { icon: '→', iconColor: 'pink', text: 'Can\'t justify Harness pricing ($100K+) for mid-market deals' },
                    { icon: '→', iconColor: 'pink', text: 'Need repeatable, auditable deployments across multiple clients' },
                  ]}
                />
              </Card>

              <Card variant="success">
                <CardTitle color="#10b981" className="mb-4">
                  Why They Choose Planton
                </CardTitle>
                <BodyText className="mb-4">
                  Planton fills the gap perfectly for these consulting firms:
                </BodyText>
                <List
                  items={[
                    { icon: '✓', iconColor: 'emerald', text: 'Deploy to any client\'s AWS, GCP, or Azure account' },
                    { icon: '✓', iconColor: 'emerald', text: 'Open-source IaC modules they can audit and customize' },
                    { icon: '✓', iconColor: 'emerald', text: 'Self-host option for clients with strict security requirements' },
                    { icon: '✓', iconColor: 'emerald', text: 'Affordable pricing that works for mid-market engagements' },
                  ]}
                />
              </Card>
            </Grid>

            <Callout variant="highlight" icon="🎯" title="The Market Gap">
              <BodyText>
                Organizations with 50-500 employees can&apos;t use Railway/Render (compliance issues),
                can&apos;t afford Harness (enterprise pricing), and can&apos;t build their own platform
                (no dedicated platform team). <strong>Planton is built for them.</strong>
              </BodyText>
            </Callout>
          </Container>
        </Section>

        {/* Current Momentum */}
        <Section background="gradient-subtle">
          <Container>
            <SectionTitle gradient>Current Momentum</SectionTitle>
            <SectionSubtitle className="mb-8">
              What we did, what we&apos;re doing, and where we&apos;re headed.
            </SectionSubtitle>

            <Grid cols={2} className="mb-8">
              {/* Last Month */}
              <Card>
                <div className="flex items-center gap-2 mb-4">
                  <Badge variant="success">Last Month</Badge>
                  <SmallText>January 2026</SmallText>
                </div>
                <div className="space-y-3">
                  {LAST_MONTH.map((item, index) => (
                    <div key={index} className="flex items-start gap-2">
                      <span className="text-[#10b981] mt-0.5">✓</span>
                      <div>
                        <span className="font-medium text-white">{item.label}</span>
                        {item.description && (
                          <SmallText className="block">{item.description}</SmallText>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </Card>

              {/* Next Month */}
              <Card>
                <div className="flex items-center gap-2 mb-4">
                  <Badge variant="purple">This Month</Badge>
                  <SmallText>February 2026</SmallText>
                </div>
                <div className="space-y-3">
                  {NEXT_MONTH.map((item, index) => (
                    <div key={index} className="flex items-start gap-2">
                      <span className="text-white mt-0.5">→</span>
                      <div>
                        <span className="font-medium text-white">{item.label}</span>
                        {item.description && (
                          <SmallText className="block">{item.description}</SmallText>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            </Grid>

            <Callout variant="highlight" icon="📊" title="Investor Updates">
              <BodyText>
                We publish regular investor updates with detailed metrics, wins, challenges, and asks.
                Transparency is part of how we operate.
              </BodyText>
              <div className="mt-4">
                <Link
                  href="/legal/investor-updates"
                  className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-[#2a2a2a] border border-[#3a3a3a] text-white hover:border-white hover:bg-white/5 transition-all duration-300"
                >
                  View Investor Updates →
                </Link>
              </div>
            </Callout>
          </Container>
        </Section>

        {/* CTA Section */}
        <Section>
          <Container>
            <div className="text-center py-8">
              <SectionTitle gradient className="text-center mb-4">
                Interested in Learning More?
              </SectionTitle>
              <SectionSubtitle className="text-center mx-auto mb-8">
                Explore the investment terms, understand the process, or reach out directly.
              </SectionSubtitle>

              <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                <Link
                  href="/invest/and-you-get"
                  className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-[#fff] text-black font-medium text-sm hover:bg-gray-200 transition-all duration-300 hover:-translate-y-0.5"
                >
                  What You Get →
                </Link>
                <Link
                  href="/invest/process"
                  className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg border border-[#3a3a3a] text-white font-medium text-sm hover:border-white hover:bg-white/5 transition-all duration-300"
                >
                  Investment Process
                </Link>
                <Link
                  href="/invest/if-you-are"
                  className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg border border-[#2a2a2a] text-[#a0a0a0] font-medium text-sm hover:border-[#3a3a3a] hover:bg-white/5 transition-all duration-300"
                >
                  What We Look For
                </Link>
              </div>

              <SmallText className="mt-8">
                Questions? Email{' '}
                <a
                  href="mailto:swarup@planton.ai"
                  className="text-white hover:underline"
                >
                  swarup@planton.ai
                </a>
              </SmallText>
            </div>
          </Container>
        </Section>
      </main>

      <Footer />
    </div>
  );
};

export default OpportunityPage;
