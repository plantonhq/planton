'use client';

import { FC, ReactNode } from 'react';
import Link from 'next/link';
import { Box, Stack, Typography } from '@mui/material';
import {
  Badge,
  CheckIcon,
  CloudIcon,
  CodeIcon,
  CpuIcon,
  PrimaryButton,
  SecondaryButton,
} from '@/components/marketing';
import { useMarket } from '@/components/market';
import { useHandoffEmail, withHandoffEmail } from '@/lib/console-handoff';
import { SITE } from '@/data/site-pages';
import {
  EVALUATION_DAYS,
  EVALUATION_URL,
  COMMUNITY_SEAT_LIMIT,
  FREE_TIER_SEATS,
} from '@/data/pricing';

/**
 * The unified plan grid: the whole self-serve story in one view — both
 * homes (Planton.ai runs it, or your own infrastructure), four plans, each
 * card carrying its own truthful AI line. Card bullets are the value
 * matrix's headline rows; the matrix below is the complete reference.
 * Prices come from the market-aware pricing-truth module, so a card can
 * never disagree with the catalog-authored charge.
 */

interface PlanCardData {
  name: string;
  badge?: { label: string; variant: 'default' | 'success' };
  priceMain: string;
  priceUnit?: string;
  priceSub?: string;
  tagline: string;
  /** Plain text, or text that is itself a door (a linked bullet renders as a link). */
  bullets: Array<string | { text: string; href: string }>;
  /**
   * The card's AI story — per-plan data, never a shared constant. The
   * assistant is deployment-capability-gated: Planton.ai plans include it
   * today on prepaid credits; self-hosted deployments do not have it yet
   * (it will run in-cluster with the customer's own LLM provider key), so
   * their cards say Coming Soon and make no credits claim.
   */
  ai: { title: string; sub: string };
  cta: { label: string; href: string; primary: boolean; external?: boolean };
  highlighted?: boolean;
}

const GroupHeader: FC<{ icon: ReactNode; title: string; sublabel: string }> = ({
  icon,
  title,
  sublabel,
}) => (
  <Box className="flex items-center gap-2.5 px-1 pb-1">
    <Box className="text-[#595959]">{icon}</Box>
    <Typography className="text-xs font-semibold uppercase tracking-wider text-[#595959]">
      {title}
    </Typography>
    <Typography className="text-xs text-[#616161]">{sublabel}</Typography>
  </Box>
);

const AiLine: FC<{ ai: PlanCardData['ai'] }> = ({ ai }) => (
  <Box className="flex items-start gap-2 border-t border-[#c7c7c4] pt-3 mt-1">
    <Box className="text-[#595959] mt-0.5 [&_svg]:w-4 [&_svg]:h-4">
      <CpuIcon />
    </Box>
    <Box>
      <Typography className="text-xs font-semibold text-[#171717]">
        {ai.title}
      </Typography>
      <Typography className="text-xs text-[#616161]">
        {ai.sub}
      </Typography>
    </Box>
  </Box>
);

const PlanCard: FC<PlanCardData> = ({
  name,
  badge,
  priceMain,
  priceUnit,
  priceSub,
  tagline,
  bullets,
  ai,
  cta,
  highlighted,
}) => (
  <Box
    className={`relative flex flex-col h-full rounded-xl bg-[#eeeeeb] border p-5 transition-all duration-300 ${
      highlighted
        ? 'border-[#999996] hover:border-[#777774]'
        : 'border-[#c7c7c4] hover:border-[#999996]'
    }`}
  >
    <Stack className="gap-3 flex-1">
      <Box className="flex items-center justify-between gap-2">
        <Typography className="text-base font-semibold text-[#171717]">{name}</Typography>
        {badge && (
          <Badge variant={badge.variant} className="!px-2 !py-0.5 !text-[11px]">
            {badge.label}
          </Badge>
        )}
      </Box>
      {/* The price zone and the tagline reserve fixed room (min-heights sized
          for their worst wrap: a price line plus a two-line sub, and a
          two-line tagline) so the bullet lists — and their check marks —
          start at the same height on every card at every breakpoint. */}
      <Box className="min-h-[4.75rem]">
        <Box className="flex items-baseline gap-1">
          <Typography className="text-3xl font-bold text-[#171717]">{priceMain}</Typography>
          {priceUnit && (
            <Typography className="text-xs text-[#595959]">{priceUnit}</Typography>
          )}
        </Box>
        <Typography className={`text-xs mt-0.5 ${priceSub ? 'text-[#616161]' : 'text-transparent select-none'}`}>
          {priceSub ?? '·'}
        </Typography>
      </Box>
      <Typography className="text-sm text-[#454545] min-h-[2.5rem]">{tagline}</Typography>
      <Stack className="gap-2 flex-1">
        {bullets.map((bullet) => {
          const text = typeof bullet === 'string' ? bullet : bullet.text;
          return (
            <Box key={text} className="flex items-start gap-2">
              <Box className="mt-1 flex-shrink-0">
                <CheckIcon />
              </Box>
              <Typography className="text-xs text-[#454545] leading-relaxed">
                {typeof bullet === 'string' ? (
                  bullet
                ) : (
                  <Link href={bullet.href} className="underline underline-offset-2 hover:text-[#171717]">
                    {bullet.text}
                  </Link>
                )}
              </Typography>
            </Box>
          );
        })}
      </Stack>
      <AiLine ai={ai} />
      <Link href={cta.href} target={cta.external ? '_blank' : '_self'} className="mt-2">
        {cta.primary ? (
          <PrimaryButton className="w-full">{cta.label}</PrimaryButton>
        ) : (
          <SecondaryButton className="w-full">{cta.label}</SecondaryButton>
        )}
      </Link>
    </Stack>
  </Box>
);

// The two truthful AI stories. Hosted plans include the credit-funded
// assistant today; self-hosted installs do not ship an assistant engine yet,
// so their line is Coming Soon and names the bring-your-own-key model that
// will serve them — never a Planton-credits claim.
const AI_INCLUDED: PlanCardData['ai'] = {
  title: 'AI Assistant Included',
  sub: 'Prepaid credits — spend protection on by default',
};
const AI_COMING_SELF_HOSTED: PlanCardData['ai'] = {
  title: 'AI Assistant — Coming Soon',
  sub: 'Runs in your cluster with your own LLM provider key',
};

export const PlanGrid: FC = () => {
  const { market } = useMarket();
  // A console that already knows the buyer hands their email to this tab;
  // both doors into the store carry it so the form opens prefilled.
  const handoffEmail = useHandoffEmail();

  const sizeLow = market.licenses[0];
  const sizeHigh = market.licenses[market.licenses.length - 1];

  const plantonAiPlans: PlanCardData[] = [
    {
      name: 'Free',
      priceMain: `${market.symbol}0`,
      tagline: 'Everything you need to ship — no card, ever.',
      bullets: [
        'The full platform — every deployment surface',
        `Up to ${FREE_TIER_SEATS} seats — everything else unlimited`,
        'Pauses at its limit — never bills, never deletes',
      ],
      ai: AI_INCLUDED,
      cta: { label: 'Start Free', href: `${SITE.url}/login`, primary: false },
    },
    {
      name: 'Team',
      priceMain: `${market.symbol}${market.teamSeatMonthly.toLocaleString()}`,
      priceUnit: '/seat/month',
      priceSub: `or ${market.symbol}${market.teamSeatAnnual.toLocaleString()}/seat/year — two months free`,
      tagline: 'The whole platform for your whole team.',
      bullets: [
        'Everything in Free, without the seat cap',
        'Pay per seat — grow and shrink as you go',
        'Unlimited environments, resources & automation',
        'Cancel anytime, no forms, no calls',
      ],
      ai: AI_INCLUDED,
      cta: { label: 'Get Started', href: `${SITE.url}/login`, primary: true },
      highlighted: true,
    },
  ];

  const yourInfraPlans: PlanCardData[] = [
    {
      name: 'Community',
      badge: { label: 'Free Forever', variant: 'success' },
      priceMain: `${market.symbol}0`,
      tagline: 'The full core platform in your own cluster.',
      bullets: [
        `Up to ${COMMUNITY_SEAT_LIMIT} seats — no license key, no time limit`,
        'Runs fully offline; nothing ever expires',
        // The evaluation is a door, not a promise: the bullet itself opens
        // the claim section on the console's public buy page.
        {
          text: `${EVALUATION_DAYS}-day full-experience evaluation key — no card, no call`,
          href: withHandoffEmail(EVALUATION_URL, handoffEmail),
        },
      ],
      ai: AI_COMING_SELF_HOSTED,
      cta: { label: 'Talk to us', href: '/book-demo', primary: false },
    },
    {
      name: 'Licensed',
      priceMain: `From ${sizeLow.perYearCard}`,
      priceUnit: '/year',
      priceSub:
        `${sizeLow.perYear} — ${sizeLow.seatCeiling} seats · ` +
        `${sizeHigh.perYear} — ${sizeHigh.seatCeiling} seats`,
      tagline: 'Org-scale capabilities on your install.',
      bullets: [
        'Works fully offline — expiry never breaks anything',
        'Key arrives by email the moment payment completes',
        'No sales call, no account required',
      ],
      ai: AI_COMING_SELF_HOSTED,
      cta: { label: 'Talk to us', href: '/book-demo', primary: true },
    },
  ];

  return (
    <Box className="w-full px-4 md:px-8 pt-8 pb-4 bg-[#f6f6f3]">
      <Box className="max-w-7xl mx-auto">
        {/* Two labeled groups with a subtle hairline between them: vertical
            when side by side at desktop, horizontal when stacked. */}
        {/* The two groups are flex siblings, so they already stretch to one
            shared height; each group is a column whose grid fills the space
            below the header (flex-1), so all FOUR cards equalize — the row
            reads as one story, never two mismatched pairs. */}
        <Box className="flex flex-col lg:flex-row">
          <Box className="flex-1 lg:pr-6 flex flex-col">
            <GroupHeader icon={<CloudIcon />} title="Planton.ai" sublabel="we run it" />
            <Box className="grid grid-cols-1 sm:grid-cols-2 gap-5 mt-3 items-stretch flex-1">
              <PlanCard {...plantonAiPlans[0]} />
              <PlanCard {...plantonAiPlans[1]} />
            </Box>
          </Box>
          <Box className="hidden lg:block w-px self-stretch bg-gradient-to-b from-transparent via-[#2f2f2f] to-transparent" />
          <Box className="lg:hidden my-6 w-full h-px bg-gradient-to-r from-transparent via-[#2f2f2f] to-transparent" />
          <Box className="flex-1 lg:pl-6 flex flex-col">
            <GroupHeader
              icon={<CodeIcon />}
              title="Your Infrastructure"
              sublabel="your cluster, your control"
            />
            <Box className="grid grid-cols-1 sm:grid-cols-2 gap-5 mt-3 items-stretch flex-1">
              <PlanCard {...yourInfraPlans[0]} />
              <PlanCard {...yourInfraPlans[1]} />
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};
