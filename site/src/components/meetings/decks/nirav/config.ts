import { SlideConfig } from '@/components/deck';
import type { MeetingDeckConfig } from '../index';

import S01Cover from './slides/S01Cover';
import S02Problem from './slides/S02Problem';
import S03CuriosityHook from './slides/S03CuriosityHook';
import S04DeterminismInsight from './slides/S04DeterminismInsight';
import S05Journey from './slides/S05Journey';
import S06Solution from './slides/S06Solution';
import S07Demo from './slides/S07Demo';
import S08AITeammates from './slides/S08AITeammates';
import S09Enterprise from './slides/S09Enterprise';
import S10Security from './slides/S10Security';
import S11Stigmer from './slides/S11Stigmer';
import S12Traction from './slides/S12Traction';
import S13Team from './slides/S13Team';
import S14TheAsk from './slides/S13TheAsk';
import S15WhyNow from './slides/S14WhyNow';
import S16Close from './slides/S15Close';

export const niravSlides: SlideConfig[] = [
  {
    id: 'cover',
    name: 'Cover',
    component: S01Cover,
    presenterNotes: [
      '"Good evening, Nirav. Thank you for taking the time — Raghav speaks very highly of you."',
      '"I\'m Swarup, founder of Planton. I want to share a story about what we\'ve been building for the last 3 years."',
      'Keep it warm, brief — 30 seconds max',
    ],
  },
  {
    id: 'problem',
    name: 'The Problem',
    component: S02Problem,
    presenterNotes: [
      '"Two years ago, we set out to solve what every growing tech company struggles with."',
      '"DevOps talent is hard to find and expensive. You know this better than anyone — managing 60+ IT staff and 40 contractors at Republic Airways."',
      '"Every company rebuilds the same infrastructure patterns from scratch. It\'s wasteful."',
      'Connect to his experience: "At MISO, with 500+ people, the same challenge exists at scale."',
    ],
  },
  {
    id: 'curiosity-hook',
    name: 'The Question',
    component: S03CuriosityHook,
    presenterNotes: [
      '"AI writes code brilliantly now — application code, Terraform, all of it."',
      '"But writing infrastructure code is one thing. Operating infrastructure — orchestrating deployments, managing state, handling secrets, maintaining audit trails, central visibility — that is where AI has NOT been adopted."',
      'PAUSE. Let the distinction land.',
      '"We spent 2 years and $500K discovering why. The answer changed everything about how we build Planton."',
      'DO NOT REVEAL THE ANSWER YET. Move to next slide.',
    ],
  },
  {
    id: 'determinism-insight',
    name: 'The Insight',
    component: S04DeterminismInsight,
    presenterNotes: [
      '"Here\'s what we learned."',
      '"In coding, non-deterministic AI is acceptable. Code review catches mistakes. The cost of bad code is manageable."',
      '"But in infrastructure — you misconfigure a production VPC, you bring down a service. There\'s no undo."',
      '"Ops teams are deeply anxious about letting AI touch their infrastructure. And they should be."',
      'Connect to Nirav\'s CISO role: "As someone who\'s been CISO, you understand this risk better than most."',
      'KEY REVEAL: "AI must be paired with deterministic tools to earn trust."',
    ],
  },
  {
    id: 'journey',
    name: 'Our Journey',
    component: S05Journey,
    presenterNotes: [
      '"Let me be honest about our journey."',
      '"In early 2025, we built a co-pilot. No matter what we tried, it was never deterministic enough."',
      '"That failure taught us the critical insight: AI should not write infrastructure code on the fly."',
      '"AI should be the orchestration layer — selecting and configuring the right deterministic tools."',
      'Be authentic about the failure — it builds credibility.',
    ],
  },
  {
    id: 'solution',
    name: 'The Solution',
    component: S06Solution,
    presenterNotes: [
      '"This is the architecture we arrived at."',
      '"We modeled 370+ cloud resource kinds using Protocol Buffers — structured data that AI excels at working with."',
      '"The AI teammate understands what the team wants, selects from this catalog, and kicks off Terraform or Helm."',
      '"The tooling is deterministic — what you write is what you get."',
      'Keep it conceptual, not deep-tech. Save details for Q&A.',
    ],
  },
  {
    id: 'demo',
    name: 'Live Demo',
    component: S07Demo,
    presenterNotes: [
      '"Let me show you the product."',
      'TAB OUT to planton.ai — show the console',
      'Walk through: connecting a cloud provider, browsing the catalog, deploying a resource',
      'Show the AI teammate chat if available',
      'Keep the demo under 5 minutes',
    ],
  },
  {
    id: 'ai-teammates',
    name: 'AI Teammates',
    component: S08AITeammates,
    presenterNotes: [
      '"This is how we ultimately solve the hiring problem."',
      '"Instead of growing an ops team, teams add AI teammates."',
      '"They have specialized skills — AWS, GCP, Kubernetes, security."',
      '"They know your org\'s context — connections, environments, secrets."',
      '"And they\'re available on desktop, mobile, web — like Slack, wherever you are."',
      '"But every action is backed by deterministic tooling. No hallucinated Terraform."',
    ],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    component: S09Enterprise,
    presenterNotes: [
      '"Having worked in enterprises myself, I know what makes a solution enterprise-ready."',
      '"We built a Planton Runner — a Kubernetes operator that deploys completely in your private network."',
      '"For organizations that want more flexibility, we have a hybrid model — SaaS control plane, private data plane."',
      '"Runners only need outbound connections. No inbound. Works in locked-down environments."',
      '"And we built desktop and mobile apps — because DevOps teammates should be available on the go, like Slack."',
      'Connect: "At MISO\'s scale, this hybrid model is exactly what you\'d need."',
    ],
  },
  {
    id: 'security',
    name: 'Security',
    component: S10Security,
    presenterNotes: [
      '"We took security extremely seriously."',
      '"All sensitive data is encrypted at rest and on the wire."',
      '"Organizations can bring their own vaults — HashiCorp Vault, AWS Secrets Manager."',
      'KEY POINT: "Secrets are resolved just-in-time, right before execution, inside the runner."',
      '"The secrets are NEVER decrypted in the control plane. Only in the runner, which runs in YOUR network."',
      'Nirav was CISO — emphasize: "This is not bolted on. This is the architecture."',
    ],
  },
  {
    id: 'stigmer',
    name: 'Stigmer',
    component: S11Stigmer,
    presenterNotes: [
      '"Raghav wanted me to share this part of the story."',
      '"All of our AI agent research led us to extract the agent infrastructure into its own product — Stigmer."',
      '"It\'s open source, free to use, self-deployable. Has its own Kubernetes operator."',
      '"Stigmer is registered as a separate Delaware Corp. Planton is its first customer."',
      '"We have 2 pilot users in Hyderabad building agentic experiences in their own products."',
      '"They don\'t want to deal with the complexity of building robust AI agents. That\'s Stigmer."',
      '"It\'s resonating instantly with people we talk to."',
    ],
  },
  {
    id: 'traction',
    name: 'Traction',
    component: S12Traction,
    presenterNotes: [
      '"Let me give you the numbers."',
      '"3 customers actively on the platform — one Pro, one Plus, one Free. First paying customer came on about 8 months ago. Zero churn."',
      '"But here\'s what I\'m most proud of: both Planton and Stigmer run entirely on Planton for all DevOps and infrastructure. 100% dogfooding."',
      '"$500K+ self-funded — this is our conviction, our skin in the game."',
      '"370+ cloud resource kinds across 14 providers. Complete platform — web, desktop, mobile, CLI."',
    ],
  },
  {
    id: 'team',
    name: 'Team',
    component: S13Team,
    presenterNotes: [
      '"Let me introduce the team."',
      '"Suresh and I are college batchmates — we\'ve known each other since 2007."',
      '"Small team, but deeply committed. 3+ years building together, $500K+ of our own money."',
      '"Irshad has been doing all our design for 5 years. Satish built the entire web console. Avinash runs all non-technical operations."',
      '"And increasingly, AI coding agents are a core part of our engineering team — we built the entire desktop, mobile, and large parts of the web platform with them."',
    ],
  },
  {
    id: 'the-ask',
    name: 'The Ask',
    component: S14TheAsk,
    presenterNotes: [
      '"We\'ve invested $500K of our own money — salaries, cloud, office, tooling. The product is in production, customers are retained, and the roadmap is clear."',
      '"We are NOT raising because we need to build. We are raising because we need to GROW."',
      '"Our burn is lean — about $8K a month. The new capital goes to distribution."',
      '"50% to growth — developer advocate, community building, meetups, event sponsorships."',
      '"30% to enterprise sales — a B2B sales hire to convert the pipeline we already have."',
      '"20% to continued product reliability and agent expansion."',
      '"Stigmer is a separate opportunity at the same $7M cap."',
      'Be direct, confident. The product-is-built framing is your strongest position.',
    ],
  },
  {
    id: 'why-now',
    name: 'Why Now',
    component: S15WhyNow,
    presenterNotes: [
      '"The timing is critical."',
      '"AI models just crossed the capability threshold for reliable tool orchestration."',
      '"Our insight — pair AI with deterministic tools — is our moat. It took 2 years and $500K to discover."',
      '"Enterprise demand is immediate. This is not a future problem."',
      '"And you get two bets — Planton the product, and Stigmer the platform."',
    ],
  },
  {
    id: 'close',
    name: 'Close',
    component: S16Close,
    presenterNotes: [
      '"The future of DevOps is not more engineers. It is better teammates."',
      '"Thank you, Nirav. And thank you to Raghav for making this introduction."',
      '"I\'m genuinely excited about what we\'ve built. I\'d love to hear your thoughts."',
      'LISTEN. Let Nirav respond. Don\'t fill the silence.',
      'If he asks about Stigmer investment separately, explain both are available.',
    ],
  },
];

export const niravConfig: MeetingDeckConfig = {
  slides: niravSlides,
  guest: 'nirav',
  meetingDate: '2026-05-08-2030',
  company: 'MISO',
  location: 'Virtual',
};
