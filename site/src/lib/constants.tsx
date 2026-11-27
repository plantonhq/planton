import path from 'path';

const CONTENT_DIRECTORIES = {
  CHANGELOG: path.join(process.cwd(), 'public/changelog'),
  DOCS: path.join(process.cwd(), 'public/docs'),
} as const;

export const CHANGELOG_DIRECTORY = CONTENT_DIRECTORIES.CHANGELOG;
export const DOCS_DIRECTORY = CONTENT_DIRECTORIES.DOCS;

