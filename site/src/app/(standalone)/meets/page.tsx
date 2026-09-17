import { redirect } from 'next/navigation';

/**
 * Meets index page - redirects to the latest presentation
 * For now, redirect to the SEP presentation
 */
export default function MeetsIndexPage() {
  redirect('/meets/sep');
}
