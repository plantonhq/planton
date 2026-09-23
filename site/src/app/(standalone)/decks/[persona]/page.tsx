import { notFound } from 'next/navigation';
import { PersonaDeck } from '@/components/decks/PersonaDeck';
import { PERSONAS, personaDeckPath, type PersonaSlug } from '@/data/personas';
import { pageMetadata } from '@/lib/page-metadata';

interface Props {
  params: Promise<{ persona: string }>;
}

/** One deck per persona; the static export needs the list up front. */
export function generateStaticParams() {
  return PERSONAS.map((p) => ({ persona: p.slug }));
}

const isPersona = (slug: string): slug is PersonaSlug => PERSONAS.some((p) => p.slug === slug);

export async function generateMetadata({ params }: Props) {
  const { persona } = await params;
  if (!isPersona(persona)) return {};
  return pageMetadata(personaDeckPath(persona));
}

export default async function Page({ params }: Props) {
  const { persona } = await params;
  if (!isPersona(persona)) return notFound();
  return <PersonaDeck slug={persona} />;
}
