import { SlideConfig } from '@/components/deck';
import type { MeetingDeckConfig } from '../index';

// Import all Rahul Gulati slides
import S01Cover from './slides/S01Cover';
import S02YouWatchedThisWall from './slides/S02YouWatchedThisWall';
import S03TheGapAfterTheGreenCheck from './slides/S03TheGapAfterTheGreenCheck';
import S04WhatIsPlanton from './slides/S04WhatIsPlanton';
import S05NotAnotherCodingAgent from './slides/S05NotAnotherCodingAgent';
import S06PavedRoadInOneSitting from './slides/S06PavedRoadInOneSitting';
import S07BuiltOnGitHubNotAroundIt from './slides/S07BuiltOnGitHubNotAroundIt';
import S08DemoHandoff from './slides/S08DemoHandoff';
import S09TheEngineBehindTheAssistant from './slides/S09TheEngineBehindTheAssistant';
import S10BornFromPlanton from './slides/S10BornFromPlanton';
import S11WhereWeAre from './slides/S11WhereWeAre';

// ============================================================================
// RAHUL GULATI SLIDE CONFIGURATION
// ============================================================================
//
// Audience: GTM / partnerships / growth (ex-GitHub for Startups APAC,
// ex-Microsoft for Startups). Not a technical buyer, not a prospect —
// likely evaluating Planton as an advisory or channel opportunity.
//
// Notes render as HTML (PresenterNotes uses dangerouslySetInnerHTML), so
// <strong> works but any literal angle bracket must be escaped.
//
// Slides 3 and 7 are the designated cut points: dropping either leaves the
// narrative intact if the room wants to reach the demo sooner.

export const rahulGulatiSlides: SlideConfig[] = [
  {
    id: 'cover',
    name: 'Cover',
    component: S01Cover,
    presenterNotes: [
      'Before this slide: 10 minutes of HIM first. "What pulled you into this? What are you building post-GitHub?" — which scenario he is (advisor / connector / curious) decides the close.',
      'Bio in TWO minutes maximum: Zillow — built Zodiac, self-service infrastructure for 2,000 engineers — founding question: "why can\'t small teams have that?"',
      'Set the frame: <strong>five minutes of context, then I type one sentence and the product talks</strong>.',
    ],
  },
  {
    id: 'the-wall',
    name: 'The Wall You Watched',
    component: S02YouWatchedThisWall,
    presenterNotes: [
      '"I want to start with your world, not mine — correct me early if I have it wrong."',
      '"You gave startups repos, CI, Copilot. They could build from day one. Shipping still needed a platform engineer they couldn\'t hire."',
      'Land the callout, then ask: <strong>"Is that a fair read of what you saw at GitHub?"</strong> — and actually wait. Cheapest moment in the meeting to get corrected, and it tells you which scenario he is.',
      'Engage with his experience, not his LinkedIn. Do not recite his job history back at him.',
    ],
  },
  {
    id: 'the-gap',
    name: 'The Gap',
    component: S03TheGapAfterTheGreenCheck,
    presenterNotes: [
      '"Actions gets you a green check. It doesn\'t get you a running URL."',
      '"Coding agents made writing code cheap. All that did was expose where the real queue was."',
      'The third card sets up the demo: cost, permissions, compliance are answered after the fact, if ever. The demo answers them BEFORE.',
      '<strong>CUT THIS SLIDE</strong> if he is already nodding — slide 4 stands alone.',
    ],
  },
  {
    id: 'what-is-planton',
    name: 'What Is Planton',
    component: S04WhatIsPlanton,
    presenterNotes: [
      '"Two halves. Infra Hub creates the platform. Service Hub deploys applications onto it."',
      'If asked for an exact component count: <strong>"north of six hundred and climbing"</strong> — never a precise number.',
      'The failed-copilot story goes here or on the next slide: "we spent a year and a half building a generic DevOps copilot. It hallucinated. We killed it. That failure is why the AI never writes infrastructure code."',
      'Do not skip the bottom callout — runs in THEIR cloud, open source underneath, no lock-in.',
    ],
  },
  {
    id: 'not-a-coding-agent',
    name: 'Not Another Coding Agent',
    component: S05NotAnotherCodingAgent,
    presenterNotes: [
      'This is the positioning slide — slow down.',
      '"Claude Code and I aren\'t competing — I use it every day; it built half this product. But its output is code one expert understands. Planton\'s output is an organizational capability."',
      '"A coding agent speeds up the person using it. Planton turns that work into something the whole team reuses."',
      'The closing line is the one to land: <strong>"the second developer needs only the template, not the AI."</strong>',
    ],
  },
  {
    id: 'paved-road',
    name: 'The Self-Service Loop',
    component: S06PavedRoadInOneSitting,
    presenterNotes: [
      'Walk the flow left to right — this is EXACTLY the sequence of the demo he is about to watch.',
      '"The bill, the compliance posture, and the IAM policy exist before the infrastructure does. Verified data, not model memory — every price cites the provider document it came from."',
      '"One platform person paves the road in a sitting. From then on, the hundredth deployment is a form fill."',
    ],
  },
  {
    id: 'built-on-github',
    name: 'Built On GitHub',
    component: S07BuiltOnGitHubNotAroundIt,
    presenterNotes: [
      'This slide exists to defuse the Actions worry BEFORE he raises it.',
      '"GitHub owns everything up to the merge — repos, reviews, checks, Copilot. We own the mile after. That mile is where the startups you onboarded stalled."',
      '"Teams that love Actions keep Actions — our CLI runs inside a workflow. And even Service Hub writes its results back into GitHub\'s own UI."',
      'Never say "GitHub Actions replacement" — not on this slide, not anywhere.',
      '<strong>CUT THIS SLIDE</strong> if running long; the sentiment survives in Q&amp;A.',
    ],
  },
  {
    id: 'demo-handoff',
    name: 'Demo Handoff',
    component: S08DemoHandoff,
    presenterNotes: [
      '"Let me stop talking and show you."',
      '<strong>Confirm he can see the screen before the first demo sentence.</strong>',
      'Read the three watch-for items aloud, then switch to the desktop app.',
      'Demo sequence: prompt — canvas reveal (say nothing) — Cloud Bill Preview — Runner Policy sheet — RDS posture / HIPAA fold — Deploy (watch first layers go green, 60-90s, move on) — Publish — switch hats: "now I\'m my developer" — open the template as a form fill.',
      '<strong>The production console stays closed.</strong> Everything shown lives on the desktop build.',
      '<strong>If the demo breaks:</strong> do not debug in front of him. Cut to the pre-run artifacts calmly.',
    ],
  },
  {
    id: 'stigmer',
    name: 'Runs on Stigmer',
    component: S09TheEngineBehindTheAssistant,
    presenterNotes: [
      'The segue, said exactly: <strong>"One more thing before we talk about where we are. The conversation you just watched — the chat, the streaming, the approval prompts — that is not Planton code. Those are Stigmer\'s drop-in components, and Stigmer\'s runtime behind them. Planton built the canvas; Stigmer runs the agent. We wrote zero agent infrastructure."</strong>',
      'The frontend card is the differentiator to dwell on: building UI for AI agents is genuinely nuanced — streaming, tool-call rendering, approvals — and Stigmer ships it as drop-in components for web, desktop, and mobile.',
      'Tenant operations: "your customers become tenants, with usage caps and spend limits built in — the layer everyone builds badly, provided."',
      'Facts if probed: Apache-2.0 open source with a managed cloud (the Temporal distribution model). All Planton agent execution is delegated through official Stigmer SDKs.',
    ],
  },
  {
    id: 'born-from-planton',
    name: 'Born From Planton',
    component: S10BornFromPlanton,
    presenterNotes: [
      'The disclosed framing, verbatim: <strong>"Stigmer came out of our own AI research — my co-founder spun it out as a separate company in January. Planton is its first customer."</strong> Never say "we own Stigmer."',
      'Why the split helps both: "the separation lets Planton stay purely a cloud and DevOps product, while Stigmer improves agent capabilities full-time."',
      'The Rahul angle: two products, one GTM brain — if the advisory conversation lands, Stigmer widens what he would be attached to.',
      'If investment comes up: Stigmer is a separate company and a separate opportunity — mention only if HE raises it; this meeting\'s ask is GTM.',
      '<strong>Boundary:</strong> in public written channels (Discord, GitHub) the voice is enthusiastic adopter, never co-founder-adjacent. If Rahul later engages the Stigmer community, that is the voice he will see — and that is correct.',
      '<strong>CUT THIS SLIDE</strong> if time is short — the previous slide carries the value proposition.',
    ],
  },
  {
    id: 'where-we-are',
    name: 'Where We Are',
    component: S11WhereWeAre,
    presenterNotes: [
      'Full honesty here — it is the setup for the ask. "Every customer came inbound. I\'ve never done a day of outbound. The launch plan is written and hasn\'t fired."',
      'Then HIS questions: ask the rubric — <strong>"strengths, risks, gaps, and what would you focus on first?"</strong>',
      'If the read is positive, the offer: "the free tier is already shaped like the startup-program offer you used to run, with zero program around it. If building that interests you — advisor equity, or a revenue share on the channel you build — is that a conversation worth having?"',
      'If asked about the raise: <strong>"$500K on a SAFE at a $7M cap, same terms since May. Burn is about $8K a month. Half the round is growth and distribution — which is this conversation."</strong> Never volunteer runway status.',
      'Close with a DATED next step before leaving the room. No loose promises — everything becomes the follow-up agenda.',
    ],
  },
];

// ============================================================================
// RAHUL GULATI GUEST CONFIG
// ============================================================================

export const rahulGulatiConfig: MeetingDeckConfig = {
  slides: rahulGulatiSlides,
  guest: 'rahul-gulati',
  meetingDate: '2026-08-17-1700',
  company: 'Rahul Gulati',
  location: 'Virtual',
};
