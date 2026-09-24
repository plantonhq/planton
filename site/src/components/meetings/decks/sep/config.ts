import { SlideConfig } from '@/components/deck';
import type { MeetingDeckConfig } from '../index';

// Import all SEP slides
import S01Cover from './slides/S01Cover';
import S02WeKnowSEP from './slides/S02WeKnowSEP';
import S03SEPChallenge from './slides/S03SEPChallenge';
import S07WhatIsPlanton from './slides/S07WhatIsPlanton';
import S11DemoIntro from './slides/S11DemoIntro';
import S16OpenSource from './slides/S16OpenSource';
import S17Security from './slides/S17Security';
import S19ROI from './slides/S19ROI';
import S20Customers from './slides/S20Customers';
import S22NextSteps from './slides/S22NextSteps';
import S24ThankYou from './slides/S24ThankYou';

// ============================================================================
// SEP SLIDE CONFIGURATION
// ============================================================================

export const sepSlides: SlideConfig[] = [
  {
    id: 'cover',
    name: 'Cover',
    component: S01Cover,
    presenterNotes: [
      'Smile, make eye contact',
      '"Good afternoon, everyone. Thank you for having me here today."',
      '"I\'m excited to show you how Planton can help SEP scale the amazing work you\'re already doing."',
      'Keep it brief: 30 seconds max',
    ],
  },
  {
    id: 'we-know-sep',
    name: 'We Know SEP',
    component: S02WeKnowSEP,
    presenterNotes: [
      '"Before we jump into what Planton does, I want to show you we understand SEP."',
      'Read through key facts confidently',
      '"You\'re not just any consulting firm—you specialize in the hard problems where failure is expensive."',
      '"That \'do it right\' ethos? That\'s why I think Planton is a great fit."',
      'Pause: "Does this resonate with your team?"',
    ],
  },
  {
    id: 'sep-challenge',
    name: 'The Challenge',
    component: S03SEPChallenge,
    presenterNotes: [
      '"This is from your own published case study on your website."',
      'Emphasize the numbers: "<strong>Thirty days</strong>... to <strong>thirty minutes</strong>. That\'s incredible work."',
      'Pause, let it sink in',
      '"Here\'s the challenge: That was custom-built for one animal health pharma client using Terraform, Azure DevOps, GitHub Actions, Docker, and Ansible."',
      '"What if you could make that same outcome repeatable?"',
    ],
  },
  {
    id: 'what-is-planton',
    name: 'What Is Planton',
    component: S07WhatIsPlanton,
    presenterNotes: [
      '"Planton is the Self-Service Cloud Platform — it turns your own cloud account into a self-service platform."',
      '"Its Service Hub is \'Vercel for Backend, In Your Own Cloud\': just like Vercel abstracts frontend deployment complexity, Service Hub abstracts backend CI/CD."',
      '"But unlike Vercel—where you have no infrastructure control—Planton runs in <strong>your</strong> AWS, Azure, or GCP account."',
      '"Everything is open source. All Terraform code is on GitHub. No vendor lock-in."',
    ],
  },
  {
    id: 'customers',
    name: 'Customers',
    component: S20Customers,
    presenterNotes: [
      '"100% retention since 2023. No one has churned."',
      '"These are companies that <strong>could</strong> build this themselves—but chose Planton."',
      '"Why? Because rebuilding the same infrastructure scaffolding for every project is wasteful."',
    ],
  },
  {
    id: 'demo-intro',
    name: 'Demo Introduction',
    component: S11DemoIntro,
    presenterNotes: [
      '"Let me show you a realistic scenario."',
      '"Imagine SEP just signed a healthcare client. They need a HIPAA-compliant AWS environment."',
      '"Starting point: empty AWS account."',
      '"Watch how fast we go from zero to production-ready."',
      '"This is not a toy demo. This is how our customers onboard real client projects today."',
    ],
  },
  {
    id: 'open-source',
    name: 'Open Source',
    component: S16OpenSource,
    presenterNotes: [
      '"SEP\'s brand is \'do it right.\' We respect that."',
      '"All our infrastructure code is open source on GitHub."',
      '"You can audit everything. Fork it. Customize it for SEP\'s needs."',
      '"And if you ever outgrow us or need to leave? Export everything. No lock-in."',
      '"We\'re confident enough in our platform\'s value to make the code public."',
    ],
  },
  {
    id: 'security',
    name: 'Security',
    component: S17Security,
    presenterNotes: [
      '"Two ways to deploy Planton."',
      '<strong>SaaS</strong>: "Fastest to start. We connect to your cloud account with limited IAM permissions."',
      '<strong>Self-Hosted</strong>: "For maximum security and compliance."',
      '"Terraform execution happens entirely within your AWS boundary."',
      '"For SEP\'s healthcare clients (HIPAA) or aerospace clients (safety-critical), self-hosted is ideal."',
    ],
  },
  {
    id: 'roi',
    name: 'ROI',
    component: S19ROI,
    presenterNotes: [
      '"Let\'s talk ROI using SEP\'s own numbers."',
      'Walk through the calculation slowly',
      '"If you take on 20 new clients per year... and Planton saves 4 weeks per client..."',
      '"At $200/hour billable rate, that\'s $640,000 in opportunity cost recovered."',
      '"That\'s real money left on the table due to repetitive DevOps work."',
    ],
  },
  {
    id: 'next-steps',
    name: 'Next Steps',
    component: S22NextSteps,
    presenterNotes: [
      '"I\'d love to propose a pilot."',
      '"2 weeks. 1-2 use cases. Planton team fully supporting you."',
      '"We validate it works for SEP\'s specific workflow."',
      '<strong>Ask:</strong> "What do you think? Does this make sense?"',
      '<strong>Listen carefully</strong>—this is where you gauge interest',
    ],
  },
  {
    id: 'thank-you',
    name: 'Thank You',
    component: S24ThankYou,
    presenterNotes: [
      '"Thank you for your time today."',
      '"I\'m genuinely excited about the possibilities here."',
      '"SEP\'s work on standardizing DevOps is exactly what Planton makes scalable."',
      '"Let\'s schedule that pilot and technical deep-dive."',
      '<strong>Follow-up within 24 hours</strong>: Thank-you email, pilot proposal, next steps',
    ],
  },
];

// ============================================================================
// SEP GUEST CONFIG
// ============================================================================

export const sepConfig: MeetingDeckConfig = {
  slides: sepSlides,
  guest: 'sep',
  meetingDate: '2026-01-23-1400',
  location: 'Westfield, Indiana',
};
