import { Box } from '@mui/material';
import type { FC } from 'react';
import { SlideFrame } from '@/components/deck';
import {
  BodyText,
  Card,
  Doors,
  FeatureTitle,
  Grid,
  PageArtifact,
  PlatformCounts,
  ProofSentences,
  SectionSubtitle,
  SectionTitle,
  TestimonialCard,
} from '@/components/marketing';
import { artifactOf } from '@/data/artifacts';
import { POSITIONING } from '@/data/positioning';
import type { Beat, Persona } from '@/data/personas';
import { ROADMAP_DISCLOSURE, chapter, type Chapter } from '@/data/story';
import { testimonial } from '@/data/testimonials';

/**
 * The six slides a persona deck is made of, each a generic component fed its
 * content by the deck's registry. They compose the marketing primitive
 * library inside the deck's slide frame, so a slide and the page section it
 * mirrors share one look, and they state nothing the persona record or the
 * story does not. The presenter's words are the notes (derived in
 * ./registry), never painted here.
 */

const Eyebrow: FC<{ children: string }> = ({ children }) => (
  <BodyText className="text-xs font-medium tracking-wide text-fg-muted uppercase mb-3">{children}</BodyText>
);

/** The cover: who this is for, the promise in their words, and the wall they hit. */
export const CoverSlide: FC<{ persona: Persona }> = ({ persona }) => (
  <SlideFrame>
    <Box className="text-center max-w-3xl mx-auto flex flex-col items-center gap-4">
      <Eyebrow>{`The Planton Story · For ${persona.plural}`}</Eyebrow>
      <SectionTitle className="text-3xl md:text-4xl lg:text-5xl">{persona.headline}</SectionTitle>
      <BodyText className="text-sm md:text-base font-medium text-fg-secondary -mt-1">{POSITIONING.umbrella.tagline}</BodyText>
      <SectionSubtitle className="text-base md:text-lg mx-auto mt-1">{persona.wall}</SectionSubtitle>
      <BodyText className="text-fg-secondary">{persona.who}</BodyText>
    </Box>
  </SlideFrame>
);

/**
 * One chapter, told for this person: the chapter's title, the persona's
 * angle, the proof sentences chosen for them, and the record the proving
 * page shows. The chapter's claim is the presenter's sentence (in the notes).
 */
export const ChapterSlide: FC<{ persona: Persona; beat: Beat }> = ({ persona, beat }) => {
  const ch = chapter(beat.chapter);
  const artifact = artifactOf(beat.provenAt);
  const sentences = beat.proof.map((i) => ch.proof[i]);
  return (
    <SlideFrame>
      <Box className="flex flex-col gap-8">
        <Box className="text-center flex flex-col items-center">
          <Eyebrow>{`For ${persona.plural}`}</Eyebrow>
          <SectionTitle className="text-2xl md:text-3xl lg:text-4xl">{ch.title}</SectionTitle>
          <SectionSubtitle className="mx-auto text-base">{beat.angle}</SectionSubtitle>
        </Box>
        {artifact ? (
          <Box className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            <Box className="lg:col-span-5">
              <ProofSentences items={sentences} />
            </Box>
            <Box className="lg:col-span-7">
              <PageArtifact artifact={artifact} />
            </Box>
          </Box>
        ) : (
          sentences.length ? <ProofSentences items={sentences} className="max-w-3xl mx-auto w-full" /> : null
        )}
      </Box>
    </SlideFrame>
  );
};

/** The questions this person asks, answered in the chapter's words. */
export const ObjectionsSlide: FC<{ persona: Persona }> = ({ persona }) => (
  <SlideFrame>
    <Box className="flex flex-col gap-8">
      <Box className="text-center flex flex-col items-center">
        <Eyebrow>{`For ${persona.plural}`}</Eyebrow>
        <SectionTitle className="text-2xl md:text-3xl lg:text-4xl">You Will Ask</SectionTitle>
      </Box>
      <Grid cols={3}>
        {persona.objections.map((o) => (
          <Card key={o.question}>
            <Box className="flex flex-col gap-3 h-full">
              <FeatureTitle className="text-balance">{o.question}</FeatureTitle>
              <BodyText>{o.answer}</BodyText>
            </Box>
          </Card>
        ))}
      </Grid>
    </Box>
  </SlideFrame>
);

/** The numbers and the people, in their own words. */
export const ProofSlide: FC<{ persona: Persona; proof: Chapter }> = ({ persona, proof }) => (
  <SlideFrame>
    <Box className="flex flex-col gap-8">
      <Box className="text-center flex flex-col items-center">
        <SectionTitle className="text-2xl md:text-3xl lg:text-4xl">{proof.title}</SectionTitle>
        <SectionSubtitle className="mx-auto">{persona.testimonials.length ? proof.claim : proof.proof[0]}</SectionSubtitle>
      </Box>
      <PlatformCounts />
      {persona.testimonials.length ? (
        <Grid cols={persona.testimonials.length === 1 ? 1 : 2} className={persona.testimonials.length === 1 ? 'max-w-2xl mx-auto w-full' : ''}>
          {persona.testimonials.map((name) => {
            const t = testimonial(name);
            return <TestimonialCard key={t.name} name={t.name} role={t.role} company={t.company} location={t.location} quote={t.quote} />;
          })}
        </Grid>
      ) : null}
    </Box>
  </SlideFrame>
);

/**
 * The roadmap chapter, with the disclosure the story requires beside it. The
 * only place chapter 13 is ever rendered; it never reaches a page. The claim
 * says "not yet shipped" and the disclosure says it again, and that is the
 * whole of it: no chip, no third sentence, no color.
 */
export const WhatIsNextSlide: FC<{ roadmap: Chapter }> = ({ roadmap }) => (
  <SlideFrame>
    <Box className="text-center max-w-3xl mx-auto flex flex-col items-center gap-5">
      <SectionTitle className="text-2xl md:text-3xl lg:text-4xl">{roadmap.title}</SectionTitle>
      <SectionSubtitle className="mx-auto text-base">{roadmap.claim}</SectionSubtitle>
      <Card hover={false} className="w-full">
        <BodyText className="text-white text-center">{ROADMAP_DISCLOSURE}</BodyText>
      </Card>
    </Box>
  </SlideFrame>
);

/** The close: how to start, and this person's doors. */
export const StartSlide: FC<{ persona: Persona; start: Chapter }> = ({ persona, start }) => (
  <SlideFrame>
    <Box className="text-center max-w-3xl mx-auto flex flex-col items-center gap-5">
      <SectionTitle className="text-2xl md:text-3xl lg:text-4xl">{start.title}</SectionTitle>
      <SectionSubtitle className="mx-auto text-base">{start.claim}</SectionSubtitle>
      <ProofSentences items={start.proof} className="text-left w-full" />
      <Doors {...persona.doors} className="justify-center mt-4" />
    </Box>
  </SlideFrame>
);
