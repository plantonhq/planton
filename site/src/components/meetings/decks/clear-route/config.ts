import { SlideConfig } from '@/components/deck';
import type { MeetingDeckConfig } from '../index';

// Import all ClearRoute slides
import S01Cover from './slides/S01Cover';
import S02HowYouWork from './slides/S02HowYouWork';
import S03BottleneckIsRouteToLive from './slides/S03BottleneckIsRouteToLive';
import S04WhatIsPlanton from './slides/S04WhatIsPlanton';
import S05WhyFitsClearRoute from './slides/S05WhyFitsClearRoute';
import S06RunsInClientKubernetes from './slides/S06RunsInClientKubernetes';
import S07AINativeDeterministic from './slides/S07AINativeDeterministic';
import S08DemoHandoff from './slides/S08DemoHandoff';
import S09NextSteps from './slides/S09NextSteps';

// ============================================================================
// CLEARROUTE SLIDE CONFIGURATION
// ============================================================================
//
// Notes render as HTML (PresenterNotes uses dangerouslySetInnerHTML), so
// <strong> works but any literal angle bracket must be escaped.
//
// Slides 3 and 7 are the designated cut points: dropping either leaves the
// narrative intact if the room wants to reach the demo sooner.

export const clearRouteSlides: SlideConfig[] = [
  {
    id: 'cover',
    name: 'Cover',
    component: S01Cover,
    presenterNotes: [
      '"Good morning everyone, and thank you for making the time — I know this is a big group."',
      '"I am Swarup, founder of Planton. I have about ten minutes of context, then I want to get into the product."',
      'Set the frame early: <strong>this is a working session, not a pitch</strong>.',
      'Keep it to 30 seconds. Do not read the slide.',
    ],
  },
  {
    id: 'how-you-work',
    name: 'How You Work',
    component: S02HowYouWork,
    presenterNotes: [
      '"I want to start with how I understand your work, so you can correct me early if I have it wrong."',
      '"QCE spans Quality Engineering, Cloud Platforms and Developer Experience — because delivery breaks at the seams between them. Planton lands on two of those three."',
      '"Route to Live is your language. I am going to keep using it, because it is exactly the surface we compress."',
      '"And your clients are banks, insurers, national infrastructure. So anything you introduce has to clear their security review, not just yours." — this sets up the Kubernetes slide later.',
      'Land the closing line deliberately: "the question is not whether Planton is useful to ClearRoute — it is whether it is useful in your clients\' hands."',
      'Then ask: "Is that a fair read?" — and <strong>actually wait for an answer</strong>. This is the cheapest moment in the meeting to get corrected.',
      '<strong>Engage with their ideas, not their dossier.</strong> Do not recite their client names, headcounts or awards back at them — it reads as surveillance rather than interest.',
    ],
  },
  {
    id: 'bottleneck',
    name: 'The Bottleneck',
    component: S03BottleneckIsRouteToLive,
    presenterNotes: [
      '"This quote is from your own homepage. Not mine."',
      'Read it out loud, slowly. Then <strong>pause</strong>.',
      '"You already diagnosed the problem. AI made writing code cheap, and all that did was expose where the real queue was."',
      '"Every one of these three is a wait, not a work item. Nobody is typing during any of them."',
      'The third card is the one that matters commercially: "the learning does not compound across engagements."',
      '<strong>CUT THIS SLIDE</strong> if the room is impatient or already nodding — slide 4 stands alone.',
    ],
  },
  {
    id: 'what-is-planton',
    name: 'What Is Planton',
    component: S04WhatIsPlanton,
    presenterNotes: [
      '"Two halves. Infra Hub creates the platform. Service Hub deploys applications onto it."',
      '"600+ components across 17 clouds, and every one ships both a Pulumi and an OpenTofu module — the customer picks, without changing the manifest."',
      'Infra Charts: "Helm charts, but for cloud infrastructure. An Infra Chart is to an Infra Project what a Helm chart is to a Helm release."',
      'If asked for an exact component count, say <strong>"north of six hundred and climbing"</strong>. Do not quote a precise number — three internal sources count differently.',
      'If someone says our GitHub README claims 400 components and 49 charts: "The README is stale. We clean-slated the chart catalog deliberately and rebuilt it as a curated set of 17."',
      '<strong>Do not skip the bottom callout.</strong> This room is full of platform engineers waiting to catch us claiming to abstract the clouds. Say it plainly: "we are not an abstraction layer."',
    ],
  },
  {
    id: 'why-fits-clearroute',
    name: 'Why This Fits You',
    component: S05WhyFitsClearRoute,
    presenterNotes: [
      '"This is the slide I actually care about. Everything else is preamble."',
      'Walk the five points, but spend your time on the last two.',
      '"You leave an asset, not a dependency" — "when the engagement ends, the client keeps a working platform. That is a better renewal conversation than a handover document."',
      '"Best practice becomes enforceable" — <strong>this is the one to dwell on</strong>. Their clients have large DevOps teams already; the problem is not headcount, it is that the standard lives in a wiki nobody reads.',
      '"Your platform team encodes the paved road once, and every product team gets it by default." That is the pitch to a bank, not "you need fewer engineers."',
      '<strong>Never say a client needs no DevOps team.</strong> These are enterprises with big platform groups. Planton is leverage for that group, not a replacement for it.',
      'No dollar figures on this slide, and do not improvise any. We do not have validated savings numbers.',
      'Close with the callout, then stop talking: "ClearRoute will stand up self-service DevOps in your cloud." Let them react.',
    ],
  },
  {
    id: 'runs-in-client-kubernetes',
    name: 'Runs In Their Kubernetes',
    component: S06RunsInClientKubernetes,
    presenterNotes: [
      '"For a bank or an insurer, nothing matters until it runs inside their own boundary. So — one command."',
      '"Publicly pullable. No registry login, no account, no license key, no image pull secret. You could run that on the train home."',
      '<strong>The security point is the one that wins the room:</strong> "connections are secret-free. Identity comes ambiently from the runner pod through IRSA or workload identity. We never copy or store your client\'s credentials."',
      'Now be honest, on purpose: "self-hosted is in <strong>active preview</strong>. The install works today. The polish is not there yet, and our first design partner has not deployed."',
      'Why this helps: it converts the conversation from vendor-and-buyer into partner-and-partner. That is the ask on the last slide.',
      'Do not oversell here. If they put a rough preview in front of a bank on our word, we lose them permanently.',
    ],
  },
  {
    id: 'ai-native',
    name: 'AI-Native',
    component: S07AINativeDeterministic,
    presenterNotes: [
      '"Everybody says AI-native right now, so let me be specific about what we mean and what we refuse to do."',
      '"In coding, a wrong answer is caught in review. In infrastructure, a misconfigured production VPC has no undo."',
      '<strong>The key line:</strong> "so the AI never writes infrastructure code. It selects and configures from a typed catalog, and deterministic modules execute."',
      'On the comparison: "generated Terraform is plausible. Plausible is not the bar for production infrastructure — typed and validated before it runs is the bar."',
      '<strong>CUT THIS SLIDE</strong> if you are running long — the demo makes the same argument better.',
    ],
  },
  {
    id: 'demo-handoff',
    name: 'Demo Handoff',
    component: S08DemoHandoff,
    presenterNotes: [
      '"Let me stop talking and show you."',
      'Frame the desktop app as the destination, <strong>not an apology</strong>: "four surfaces — desktop, web, CLI, mobile. Desktop is where the AI teammate lives today, so it is the one we put in front of engineers."',
      'On web parity: "the web console catches up at the end of this week. We ship weekly." Say it as a release train, matter-of-fact. Do not elaborate unless asked.',
      'Read the three watch-for items aloud before you switch windows — it tells the room where to look.',
      '<strong>Scope the demo explicitly:</strong> "this is Infra Hub. Service Hub runs on the hosted and self-hosted consoles and reaches desktop in a coming release." Say it up front so nobody asks for it mid-demo.',
      'If someone asks to see a service deployed anyway: "I can show you that on the hosted console after this, or in the technical deep dive."',
      '<strong>If the demo breaks:</strong> do not debug in front of them. "This is exactly why I am not showing you a recorded demo — let me pick it up on the deep-dive call." Then go to the last slide.',
    ],
  },
  {
    id: 'next-steps',
    name: 'Next Steps',
    component: S09NextSteps,
    presenterNotes: [
      '"Here is what I would actually like, and it is not a purchase order."',
      '"Be our design partner for self-hosted, on one live engagement, with our team embedded."',
      'Then <strong>stop</strong>. Do not fill the silence. This is the moment you learn whether it landed.',
      '',
      '<strong>Q&amp;A PREP — the three questions that will come:</strong>',
      '<strong>"How many customers do you have?"</strong> — Be straight: a small number, and name them if useful. Then pivot: "the more useful fact is that Planton and Stigmer both run entirely on Planton. We are our own hardest customer."',
      '<strong>"How is this different from Backstage or Port?"</strong> — "Those are portals. You still have to build the backend they point at. Planton brings the execution layer with it — the catalog, the modules, the runner."',
      '<strong>"What if Planton goes away?"</strong> — "Open source owns how to deploy: the schemas, the modules, the CLI, the self-hosting Helm charts, all Apache-2.0. Planton Platform owns workflow and governance. You would lose the surrounding system, never the ability to keep deploying."',
      'If asked about pricing: "we are reworking packaging right now and I would rather not quote you a number I have to walk back. Let me come back to you with it."',
      'Follow up within 24 hours: thank-you note, design-partner proposal, deep-dive scheduling.',
    ],
  },
];

// ============================================================================
// CLEARROUTE GUEST CONFIG
// ============================================================================

export const clearRouteConfig: MeetingDeckConfig = {
  slides: clearRouteSlides,
  guest: 'clear-route',
  meetingDate: '2026-08-12-1100',
  company: 'ClearRoute',
  location: 'Virtual',
};
