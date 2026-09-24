/**
 * The ambient module types Next provides for images and its own globals.
 * `next build` writes the same references into next-env.d.ts, but that file
 * is gitignored and does not exist on a clean checkout, and the type check
 * runs before the first build. Without this file, an `import icon from
 * '...svg'` has no module declaration and `tsc --noEmit` fails in CI while
 * passing on any machine that has built the site once.
 */
/// <reference types="next" />
/// <reference types="next/image-types/global" />
