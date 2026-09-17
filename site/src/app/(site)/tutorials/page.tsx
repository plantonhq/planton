import React from 'react';
import { getAllTutorials, getAllCategories } from '@/lib/tutorials';
import TutorialsPageClient from '@/components/tutorials/TutorialsPageClient';

export default function TutorialsPage() {
  const tutorials = getAllTutorials();
  const categories = getAllCategories();

  return <TutorialsPageClient tutorials={tutorials} categories={categories} />;
}
