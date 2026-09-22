import { Box } from '@mui/material';
import type { FC } from 'react';
import {
  CenteredCards,
  Doors,
  Grid,
  PageArtifact,
  PageCard,
  PageHero,
  ProofList,
  ProviderStrip,
  Section,
  SectionSubtitle,
  SectionTitle,
  Step,
} from '@/components/marketing';
import { START_DOORS } from '@/data/doors';
import { productPage } from '@/data/product';
import { sitePage } from '@/data/site-pages';

/**
 * One template renders every Product page from its record in
 * src/data/product.ts. The order is the user's order: what this is and who
 * it is for, how it works as a person experiences it, what you get (the
 * proof points beside the record the product shows, or the commands a person
 * types), where the same claims are proven for the person who signs (the
 * Trust pages), then the doors: two sibling pages and the way to start. The
 * template adds no sentence of its own.
 */
export const ProductPage: FC<{ path: string }> = ({ path }) => {
  const page = sitePage(path);
  const record = productPage(path);
  const doors = record.doors ?? START_DOORS;

  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Product', href: '/product' }} title={page.title} kicker={record.analogy} lede={record.lede} forWhom={record.forWhom}>
        <Doors {...doors} className="justify-center mt-2" />
      </PageHero>
      {record.strip === 'providers' ? (
        <Box className="max-w-5xl mx-auto -mt-10 mb-4 px-4">
          <ProviderStrip />
        </Box>
      ) : null}

      {record.steps ? (
        <Section>
          <Box className="max-w-6xl mx-auto flex flex-col items-center gap-10">
            <SectionTitle>How It Works</SectionTitle>
            <Box className="flex flex-col lg:flex-row gap-8 lg:gap-4 w-full">
              {record.steps.map((step, i, all) => (
                <Step key={step.title} number={i + 1} title={step.title} description={step.text} isLast={i === all.length - 1} />
              ))}
            </Box>
          </Box>
        </Section>
      ) : null}

      <Section>
        <Box className="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start max-w-6xl mx-auto">
          <Box className="lg:col-span-5 flex flex-col gap-4">
            <SectionTitle>What You Get</SectionTitle>
            <SectionSubtitle className="mt-0">Shipped behavior, stated as it works today.</SectionSubtitle>
            <ProofList items={record.points} />
          </Box>
          <Box className="lg:col-span-7 lg:sticky lg:top-24">
            <PageArtifact artifact={record.artifact} />
          </Box>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-5xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Where This Is Proven</SectionTitle>
          <SectionSubtitle className="mx-auto mt-0">The same claims, written for the person who signs, with the record beside each one.</SectionSubtitle>
          <CenteredCards className="w-full">
            {record.provenAt.map((trustPath) => (
              <PageCard key={trustPath} path={trustPath} linkLabel="What It Proves" />
            ))}
          </CenteredCards>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Read Next</SectionTitle>
          <Grid cols={2} className="w-full">
            {record.next.map((sibling) => (
              <PageCard key={sibling} path={sibling} linkLabel="Read More" />
            ))}
          </Grid>
          <Doors {...doors} className="mt-4" />
        </Box>
      </Section>
    </main>
  );
};
