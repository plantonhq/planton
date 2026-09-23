import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, CenteredCards, Doors, PageCard, PageHero, PersonaCard, Section } from '@/components/marketing';
import { START_DOORS } from '@/data/doors';
import { PERSONAS } from '@/data/personas';
import { POSITIONING } from '@/data/positioning';
import { chapter } from '@/data/story';

/**
 * The door to the Solutions section: chapter 9 as the headline and the
 * promise (the platform engineer is the user, the engineering leader is who
 * signs), then the five people the story is told to, each as a door to
 * their page, and one more door for the developer who arrived with a coding
 * agent open: their way in is the agent's, so their card is the Coding
 * Agents page (the story never headlines "built for developers"). For the
 * visitor deciding whether Planton is for someone like them. The index adds
 * no sentence of its own.
 */
export const SolutionsIndex: FC = () => {
  const who = chapter('who-it-is-for');
  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Solutions' }} title={who.title} kicker={POSITIONING.umbrella.tagline} lede={who.claim} forWhom="Pick the page written for your role.">
        <Doors {...START_DOORS} className="justify-center mt-2" />
        <BodyText className="text-sm text-fg-secondary mt-2">
          Writing code with a coding agent open?{' '}
          <Link href="/product/coding-agents" className="underline underline-offset-4 hover:text-white">
            {'Start with how your agent reaches Planton \u2192'}
          </Link>
        </BodyText>
      </PageHero>
      <Section>
        <CenteredCards className="max-w-6xl mx-auto">
          {PERSONAS.map((persona) => (
            <PersonaCard key={persona.slug} slug={persona.slug} />
          ))}
          <PageCard path="/product/coding-agents" linkLabel="How Your Agent Reaches Planton" />
        </CenteredCards>
        <Doors {...START_DOORS} className="justify-center mt-12" />
      </Section>
    </main>
  );
};
