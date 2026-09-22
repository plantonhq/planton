import { notFound } from 'next/navigation';
import { PersonaPage } from '@/components/solutions/PersonaPage';
import { PERSONAS, personaPagePath, type PersonaSlug } from '@/data/personas';
import { pageMetadata } from '@/lib/page-metadata';

interface Props {
  params: Promise<{ persona: string }>;
}

/** One page per persona; the static export needs the list up front. */
export function generateStaticParams() {
  return PERSONAS.map((p) => ({ persona: p.slug }));
}

const isPersona = (slug: string): slug is PersonaSlug => PERSONAS.some((p) => p.slug === slug);

export async function generateMetadata({ params }: Props) {
  const { persona } = await params;
  if (!isPersona(persona)) return {};
  return pageMetadata(personaPagePath(persona));
}

export default async function Page({ params }: Props) {
  const { persona } = await params;
  if (!isPersona(persona)) return notFound();
  return <PersonaPage slug={persona} />;
}
