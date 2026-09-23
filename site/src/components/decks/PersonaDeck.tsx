'use client';

import { Deck } from '@/components/deck';
import { BodyText } from '@/components/marketing';
import { persona as personaRecord, personaDeckPath, type PersonaSlug } from '@/data/personas';
import { SITE } from '@/data/site-pages';
import { PERSONA_DECKS } from './registry';

/**
 * The story told for one person, as a deck at /decks/<slug>. The slides are
 * the persona's record rendered by the deck engine (see ./registry); the
 * frame names the deck and its address so a screenshot of any slide says
 * where it came from. A client component because the slide list carries
 * bound components, which are built here on the client and never cross the
 * server boundary.
 */
export function PersonaDeck({ slug }: { slug: PersonaSlug }) {
  const persona = personaRecord(slug);
  // Beside the logo, clear of the engine's centered dot strip; hidden where
  // the two would meet.
  const frame = (
    <BodyText className="hidden md:block absolute top-[30px] left-20 text-xs text-fg-muted">
      {`The Planton Story for ${persona.plural} · ${SITE.url.replace(/^https?:\/\//, '')}${personaDeckPath(slug)}`}
    </BodyText>
  );
  return <Deck slides={PERSONA_DECKS[slug]} frame={frame} />;
}
