import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import {
  BodyText,
  Card,
  ChapterSection,
  Doors,
  FeatureCard,
  FeatureTitle,
  Grid,
  PageArtifact,
  PageCard,
  PageHero,
  PlatformCounts,
  ProofSentences,
  Section,
  SectionSubtitle,
  SectionTitle,
  TestimonialCard,
} from '@/components/marketing';
import { artifactOf } from '@/data/artifacts';
import { persona as personaRecord, personaPagePath, type Beat, type Persona, type PersonaSlug } from '@/data/personas';
import { POSITIONING } from '@/data/positioning';
import { sitePage } from '@/data/site-pages';
import { chapter } from '@/data/story';
import { testimonial } from '@/data/testimonials';

/**
 * One template renders every persona page from its record in
 * src/data/personas.ts. The order is this person's order: the promise in
 * their words and the wall they hit, the two or three proof points that
 * decide it, the chapters they need most as full sections (the persona's
 * angle over the chapter's proof, beside the record the proving page shows),
 * the rest of the story as doors, the questions they will ask answered where
 * they arise, the people and the numbers, then two other roles and the
 * doors. The template adds no sentence of its own.
 */
export const PersonaPage: FC<{ slug: PersonaSlug }> = ({ slug }) => {
  const persona = personaRecord(slug);
  const lead = persona.beats.filter((b) => b.weight === 'lead');
  const supporting = persona.beats.filter((b) => b.weight === 'supporting');
  const proof = chapter('proof-it-works');

  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Solutions', href: '/solutions' }} title={persona.headline} kicker={`${POSITIONING.umbrella.tagline}, for ${persona.plural}`} lede={persona.wall} forWhom={persona.who}>
        <Doors {...persona.doors} className="justify-center mt-2" />
        <BodyText className="text-sm text-fg-secondary max-w-2xl mt-2">{POSITIONING.umbrella.sentence}</BodyText>
      </PageHero>

      <Section>
        <Box className="max-w-6xl mx-auto flex flex-col items-center gap-8">
          <SectionTitle>What Decides It</SectionTitle>
          <Grid cols={3} className="w-full">
            {persona.decidingProof.map((point) => (
              <FeatureCard key={point.label} title={point.label} description={point.text} />
            ))}
          </Grid>
        </Box>
      </Section>

      {lead.map((beat, i) => (
        <LeadBeat key={beat.chapter} beat={beat} reverse={i % 2 === 1} />
      ))}

      <Section>
        <Box className="max-w-6xl mx-auto flex flex-col items-center gap-8">
          <Box className="text-center flex flex-col items-center">
            <SectionTitle>The Rest of the Story</SectionTitle>
            <SectionSubtitle className="mx-auto">The chapters that matter next, each with the page that tells it in full.</SectionSubtitle>
          </Box>
          <Grid cols={supporting.length % 3 === 0 ? 3 : 2} className="w-full">
            {supporting.map((beat) => (
              <SupportingBeat key={beat.chapter} beat={beat} />
            ))}
          </Grid>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-6xl mx-auto flex flex-col items-center gap-8">
          <SectionTitle>You Will Ask</SectionTitle>
          <Grid cols={3} className="w-full">
            {persona.objections.map((objection) => (
              <ObjectionCard key={objection.question} objection={objection} persona={persona} />
            ))}
          </Grid>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-6xl mx-auto flex flex-col items-center gap-8">
          <Box className="text-center flex flex-col items-center">
            <SectionTitle>{proof.title}</SectionTitle>
            {persona.testimonials.length ? <SectionSubtitle className="mx-auto">{proof.claim}</SectionSubtitle> : null}
          </Box>
          <PlatformCounts className="w-full" />
          {persona.testimonials.length ? null : (
            <BodyText className="text-center text-fg-secondary max-w-2xl">
              {proof.proof[2]}{' '}
              <Link href="/product/open-source" className="underline underline-offset-4 hover:text-white">
                {'Open-Source Modules \u2192'}
              </Link>
            </BodyText>
          )}
          {persona.testimonials.length ? (
            <Grid cols={persona.testimonials.length === 1 ? 1 : 2} className={persona.testimonials.length === 1 ? 'w-full max-w-2xl' : 'w-full max-w-5xl'}>
              {persona.testimonials.map((name) => {
                const t = testimonial(name);
                return <TestimonialCard key={t.name} name={t.name} role={t.role} company={t.company} location={t.location} quote={t.quote} />;
              })}
            </Grid>
          ) : null}
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Read It for Another Role</SectionTitle>
          <Grid cols={2} className="w-full">
            {persona.siblings.map((sibling) => (
              <PageCard key={sibling} path={personaPagePath(sibling)} linkLabel="How Planton Fits Your Work" />
            ))}
          </Grid>
          <Doors {...persona.doors} className="mt-4" />
        </Box>
      </Section>
    </main>
  );
};

/**
 * A lead beat: the chapter's title, the persona's angle as the subtitle, the
 * chapter's chosen proof sentences, the record the proving page shows beside
 * them, and a door to that page. Sides alternate so the page does not read
 * as one shape repeated. A proving page with no record (an index) stacks.
 */
const LeadBeat: FC<{ beat: Beat; reverse: boolean }> = ({ beat, reverse }) => {
  const ch = chapter(beat.chapter);
  const artifact = artifactOf(beat.provenAt);
  const proven = sitePage(beat.provenAt);
  const sentences = beat.proof.map((i) => ch.proof[i]);
  return (
    <ChapterSection
      chapter={ch}
      subtitle={beat.angle}
      layout={artifact ? (reverse ? 'split-reverse' : 'split') : 'center'}
      aside={artifact ? <ProofSentences items={sentences} /> : undefined}
      readMore={{ href: beat.provenAt, label: proven.title }}
    >
      {artifact ? (
        <PageArtifact artifact={artifact} />
      ) : (
        sentences.map((text) => (
          <BodyText key={text} className="text-center text-fg-secondary max-w-2xl mx-auto">
            {text}
          </BodyText>
        ))
      )}
    </ChapterSection>
  );
};

/** A supporting beat: the chapter's title, the persona's angle, and a door to the page that tells it in full. */
const SupportingBeat: FC<{ beat: Beat }> = ({ beat }) => {
  const ch = chapter(beat.chapter);
  const proven = sitePage(beat.provenAt);
  return (
    <Card>
      <Box className="flex flex-col gap-3 h-full text-left">
        <FeatureTitle className="text-balance">{ch.title}</FeatureTitle>
        <BodyText className="flex-1">{beat.angle}</BodyText>
        <Link href={beat.provenAt} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
          {`${proven.title} \u2192`}
        </Link>
      </Box>
    </Card>
  );
};

/**
 * An objection and its answer, with a door to the page that tells the
 * answering chapter for this person when one of their beats does.
 */
const ObjectionCard: FC<{ objection: Persona['objections'][number]; persona: Persona }> = ({ objection, persona }) => {
  const href = objection.readMore ?? persona.beats.find((b) => b.chapter === objection.chapter)?.provenAt;
  const proven = href ? sitePage(href) : undefined;
  return (
    <Card>
      <Box className="flex flex-col gap-3 h-full text-left">
        <FeatureTitle className="text-balance">{objection.question}</FeatureTitle>
        <BodyText className="flex-1">{objection.answer}</BodyText>
        {href && proven ? (
          <Link href={href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${proven.title} \u2192`}
          </Link>
        ) : null}
      </Box>
    </Card>
  );
};
