import { bindSlide, type SlideConfig } from '@/components/deck';
import { PERSONAS, type Persona, type PersonaSlug } from '@/data/personas';
import { chapter } from '@/data/story';
import { ChapterSlide, CoverSlide, ObjectionsSlide, ProofSlide, StartSlide, WhatIsNextSlide } from './slides';

/**
 * Every persona's slides, built once from its record when this module loads:
 * the cover, one slide per beat in the persona's order, the objections, the
 * proof, the roadmap with its disclosure, and the close. Binding here at
 * module scope gives each slide a stable component identity, so the engine's
 * cross-fade never remounts a slide because the presenter toggled the notes.
 *
 * The presenter notes are derived, not written: the persona's opening line
 * on the cover; for each chapter the claim (the sentence to say), the proof
 * sentences on the slide, and the chapter's never-say list as what not to
 * say; the disclosure line on the roadmap. A chapter edit in the story
 * changes what the presenter is told to say in the same commit.
 */
const notes = {
  neverSay: (items: readonly string[]) => (items.length ? [`<strong>Do not say:</strong> ${items.join('; ')}.`] : []),
};

function slidesFor(persona: Persona): SlideConfig[] {
  const wall = chapter('the-wall');
  const proof = chapter('proof-it-works');
  const roadmap = chapter('what-is-next');
  const start = chapter('start');
  return [
    bindSlide(CoverSlide, { persona }, {
      id: 'cover',
      name: 'Cover',
      // The cover is chapter 1 told for this person: the wall, then the story's own words for it.
      presenterNotes: [persona.opening, wall.claim, `<strong>Who is in the room:</strong> ${persona.who}`, ...notes.neverSay(wall.neverSay)],
    }),
    ...persona.beats.map((beat) => {
      const ch = chapter(beat.chapter);
      return bindSlide(ChapterSlide, { persona, beat }, {
        id: beat.chapter,
        name: ch.title,
        presenterNotes: [ch.claim, `<strong>For ${persona.plural.toLowerCase()}:</strong> ${beat.angle}`, ...beat.proof.map((i) => ch.proof[i]), ...notes.neverSay(ch.neverSay)],
      });
    }),
    bindSlide(ObjectionsSlide, { persona }, {
      id: 'you-will-ask',
      name: 'You Will Ask',
      presenterNotes: persona.objections.map((o) => `<strong>${o.question}</strong> ${o.answer}`),
    }),
    bindSlide(ProofSlide, { persona, proof }, {
      id: proof.id,
      name: proof.title,
      presenterNotes: [proof.claim, ...proof.proof, ...notes.neverSay(proof.neverSay)],
    }),
    bindSlide(WhatIsNextSlide, { roadmap }, {
      id: roadmap.id,
      name: roadmap.title,
      presenterNotes: [roadmap.claim, ...roadmap.proof, ...notes.neverSay(roadmap.neverSay)],
    }),
    bindSlide(StartSlide, { persona, start }, {
      id: start.id,
      name: start.title,
      presenterNotes: [start.claim, ...start.proof, ...notes.neverSay(start.neverSay)],
    }),
  ];
}

/** Slide ids become URL hashes; two slides with one id would share an address. */
function uniqueIds(persona: Persona, slides: SlideConfig[]): SlideConfig[] {
  const seen = new Set<string>();
  for (const s of slides) {
    if (seen.has(s.id)) throw new Error(`the ${persona.slug} deck has two slides with the id "${s.id}"`);
    seen.add(s.id);
  }
  return slides;
}

export const PERSONA_DECKS: Readonly<Record<PersonaSlug, SlideConfig[]>> = Object.fromEntries(
  PERSONAS.map((p) => [p.slug, uniqueIds(p, slidesFor(p))]),
) as Record<PersonaSlug, SlideConfig[]>;
