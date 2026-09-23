import type { FC } from 'react';
import { BodyText, Card, ChapterSection, FeatureTitle, Grid } from '@/components/marketing';
import { chapter } from '@/data/story';
/**
 * Chapter 1. The wall a developer with a coding agent hits today, said as
 * three facts about what the agent leaves behind. This section names the
 * problem and nothing else: no product, no price, no blame on the agent (it
 * is good at this; that is the problem). The hero already carries the
 * chapter's first sentence, so the subtitle here is its second half.
 */

const ch = chapter('the-wall');

const FACETS = [
  { title: 'Unverified', text: 'Nobody priced it before it existed, and nobody checked what permissions it was given.' },
  { title: 'Unrecorded', text: 'Nothing remembers exactly what was made, by whom, or why. The transcript is not a record.' },
  { title: 'Unrepeatable', text: 'The next environment starts from a blank prompt, and staging will not match production.' },
];

export const TheWall: FC = () => (
  <ChapterSection chapter={ch} subtitle={ch.proof[0]}>
    <Grid cols={3} className="max-w-6xl mx-auto">
      {FACETS.map((facet) => (
        <Card key={facet.title} hover={false}>
          <FeatureTitle className="mb-2">{facet.title}</FeatureTitle>
          <BodyText>{facet.text}</BodyText>
        </Card>
      ))}
    </Grid>
  </ChapterSection>
);
