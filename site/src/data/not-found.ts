/**
 * The words of the site's 404, the page every address without a page
 * answers, including the pages the site has retired. A retired page gets no
 * redirect (site README, law 5): the visitor lands here, learns in one line
 * that the page is gone, reads what Planton is, and gets two doors.
 *
 * Chapter: the-wall (the stand, read from story.ts, never retyped), so the
 * 404 says what Home's headline says.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { NOT_FOUND_DOORS } from './doors.ts';
import { STAND } from './story.ts';

export const NOT_FOUND = {
  eyebrow: '404',
  title: 'There’s no page here.',
  line: 'The page you followed may have moved or been retired.',
  sentence: STAND,
  doors: NOT_FOUND_DOORS,
} as const;
