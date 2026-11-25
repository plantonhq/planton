import { FREE_TIER_SEATS } from './pricing.ts';

/**
 * Home, one section per record, in the order the page tells the story (the
 * story's spine, v3.0: a company with no DevOps team runs its own cloud).
 * Each section names the chapter it mirrors; components render these records
 * and state nothing. Home is built one or two sections per session, so the
 * records here are the sections that exist today, in page order.
 */
export const HOMEPAGE = {
  title: 'Run Your Cloud Without a DevOps Team',
  description:
    // ch 1 (a new cloud account, and no DevOps team): the stand, then the first action.
    'Run your cloud without a DevOps team. Paste one prompt into your coding agent, and it takes your cloud account to a running app in your own cloud.',
  // ch 1: the stand. The team the diagram beside it draws is the rest of the hero's story.
  headline: 'Run your cloud without a DevOps team.',
  intro:
    // ch 1 (one prompt to a running app), then ch 6 (every push ships through one path).
    'Paste one prompt into your coding agent, and your app runs in your own cloud account, even a brand-new one. Then every push ships through one path, with your team’s approvals.',
  // ch 12 (start): the price answered at the first door. The free tier never asks for a card (the claims ledger, the price row).
  caption: `Free for up to ${FREE_TIER_SEATS} seats. Never asks for a card.`,
  /**
   * ch 1: the founder's measure of the product, proven by the setup prompt
   * itself. Paste to running app is proven on Google Cloud Run (the claims
   * ledger, the Setup Prompt row), so no other cloud is named beside it.
   */
  firstApp: {
    // A non-breaking hyphen keeps "brand-new" whole when the heading wraps.
    title: 'From a brand\u2011new cloud account to a running app.',
    intro:
      'These are the exact words your coding agent reads. Planton writes them for you when you sign up, with your own one\u2011time code.',
    promptLabel: 'An example prompt, for Neha at Slotwren, a demo company',
    you: {
      title: 'You',
      steps: [
        'Open your app’s repository in Claude Code, Codex or Cursor, and paste the prompt.',
        'Say yes before anything changes your cloud or costs money.',
        'Install Planton’s GitHub App, so every push after this ships the same way.',
      ],
    },
    agent: {
      title: 'Your Agent',
      steps: [
        'Installs the Planton CLI and signs in as you.',
        'Connects your cloud without storing a key.',
        'Gets your app running in your own cloud account.',
        'Tells you what it set up, with a link to each thing.',
      ],
    },
  },
  // ch 12 (start): what the visitor keeps, then the first act.
  close: {
    title: 'Your app, running in your own cloud, on your own bill.',
    // ch 8 (runs where you decide): what the agent builds stays in the customer's account (the claims ledger, the own-cloud row).
    text: 'What your agent builds stays in your cloud account, and your cloud provider bills you directly.',
    note: `Free for up to ${FREE_TIER_SEATS} seats. Never asks for a card.`,
  },
} as const;

/**
 * The setup prompt a new organization's Home hands over, as the server writes
 * it (`SetupAgentPrompt.render` in the platform repository), for Slotwren, a
 * demo company. The page shows the words and never a working prompt: the real
 * one carries a personal single-use sign-in code, so `{code}` renders as a
 * marked placeholder and there is no copy button. When the server's wording
 * changes, this example changes with it.
 */
export const SETUP_PROMPT_EXAMPLE = {
  code: '<her one-time code>',
  /** The sentence the page sets in full ink: the agent asks before it costs anything. */
  emphasis: 'Ask me before anything that changes my cloud or costs money; connecting a cloud account counts.',
  lines: [
    "Set up Planton for this repository and take it to its first deployment, in my organization slotwren on planton.ai. I'm neha@slotwren.com.",
    '',
    '1. Install the Planton CLI and sign it in as me. The code works once, within 15 minutes:',
    '   curl -fsSL https://planton.ai/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh',
    '   ~/.planton/bin/planton login --grant {code} --instance planton.ai',
    '   Call the CLI by that full path in every command.',
    '2. Fetch https://planton.ai/agents.md with curl, read it as written, and follow it. Skip what you already did in step 1.',
    "3. Follow the planton skill's first-deployment guide for this repository in slotwren. Ask me before anything that changes my cloud or costs money; connecting a cloud account counts. Tick each Getting Started step as you finish it with ~/.planton/bin/planton setup done <step> (run it with no step to list them).",
    '4. When you finish, tell me what you set up, with a link to each thing.',
  ],
} as const;

/** Source-scoped booking copy; the form payload and existing calendar stay unchanged. */
export const DEMO_COPY = {
  eyebrow: 'LET’S MAKE IT CONCRETE',
  title: 'Your cloud.\nYour team.\nYour walkthrough.',
  intro:
    'See how Planton could connect infrastructure setup and application delivery for your team. Tell us a little about yourself, then choose a time to talk.',
  agendaTitle: 'What we’ll walk through',
  agenda: [
    'Create a reusable cloud environment.',
    'Take a service from repository to deployment.',
    'Review access, approvals, and the change record.',
  ],
  note: 'Bring your current stack and the parts of delivery you want to improve.',
  steps: ['Your details', 'Choose a time'],
  formTitle: 'Start with your details',
  submit: 'Continue to available times',
  schedulerTitle: 'Choose a time that works for you.',
  schedulerText:
    'Your details have been submitted. Select and confirm a slot below to book your meeting.',
  loading: 'Loading available times…',
  calendarFailure:
    'The calendar could not load here. Your details are saved; open the scheduling page to choose a time.',
  fallback: 'Having trouble with the calendar?',
  calendarLink: 'Open the scheduling page',
  calLink: 'swarup-donepudi/60min',
  calUrl: 'https://cal.com/swarup-donepudi/60min',
  confirmation: 'Your meeting is booked.',
  confirmationText:
    'Look for the calendar invitation in your email. We look forward to learning about your team.',
} as const;
