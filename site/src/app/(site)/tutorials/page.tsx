import React from 'react';
import { getAllTutorials, getAllCategories } from '@/lib/tutorials';
import TutorialsPageClient from '@/components/tutorials/TutorialsPageClient';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/tutorials');

export default function TutorialsPage() {
  const tutorials = getAllTutorials();
  const categories = getAllCategories();

  return <TutorialsPageClient tutorials={tutorials} categories={categories} />;
}
