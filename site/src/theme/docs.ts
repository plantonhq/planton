/**
 * Centralized design tokens for the documentation system.
 *
 * Every docs component imports class strings from here instead of
 * hardcoding Tailwind color classes. This guarantees the monochrome
 * palette is enforced in one place and makes future palette changes
 * a single-file edit.
 *
 * Palette: the hexes in these class strings are the website palette defined
 * once in packages/website-shell/src/theme/tokens.ts (canvas #0a0a0a, panel
 * #111111, raised #1a1a1a, text #ededed / #a0a0a0 / #666666, edges #2a2a2a /
 * #3a3a3a, semantic green #10b981 and red #ef4444). This file predates the
 * role-named Tailwind classes (bg-canvas, text-fg-secondary, border-edge) and
 * moves onto them when the docs pages are rebuilt.
 */

// ---------------------------------------------------------------------------
// Content – Markdown / MDX body
// ---------------------------------------------------------------------------

export const LINK_CLASSES =
  'text-white/80 hover:text-white underline underline-offset-2 decoration-white/30 hover:decoration-white/60 break-words';

export const TAG_CLASSES =
  'px-2 md:px-3 py-1 bg-white/10 text-white/70 text-xs md:text-sm font-medium rounded-full border border-white/10';

export const BLOCKQUOTE_CLASSES =
  'border-l-2 border-[#3a3a3a] pl-4 py-3 my-5 bg-[#111] rounded-r text-[#a0a0a0]';

export const BLOCKQUOTE_WARNING_CLASSES =
  'border-l-2 border-[#ef4444]/40 pl-4 py-3 my-5 bg-[#111] rounded-r text-[#a0a0a0]';

export const INLINE_CODE_CLASSES =
  'bg-[#2a2a2a] text-white rounded text-sm break-words';

// In-body headings use #b0b0b0 (secondary.80) with semibold weight —
// ~10% brighter than body text (#a0a0a0). Size and weight carry the
// structural hierarchy; the subtle brightness lift aids scanning without
// creating visual fatigue in dense documentation. The frontmatter page
// title stays at text-white font-bold (set in MDXRenderer header).
export const HEADING_H1_CLASSES =
  'text-xl sm:text-2xl md:text-3xl font-semibold text-[#b0b0b0] mt-6 md:mt-8 mb-3 md:mb-4';

export const HEADING_H2_CLASSES =
  'text-lg sm:text-xl md:text-2xl font-semibold text-[#b0b0b0] mt-5 md:mt-6 mb-2 md:mb-3';

export const HEADING_H3_CLASSES =
  'text-base sm:text-lg md:text-xl font-semibold text-[#b0b0b0] mt-4 md:mt-5 mb-2';

export const HEADING_H4_CLASSES =
  'text-base md:text-lg font-semibold text-[#b0b0b0] mt-3 md:mt-4 mb-2';

export const HEADING_H5_CLASSES =
  'text-sm md:text-base font-semibold text-[#b0b0b0] mt-3 mb-2';

export const HEADING_H6_CLASSES =
  'text-sm font-semibold text-[#b0b0b0] mt-2 mb-1';

export const PARAGRAPH_CLASSES = 'text-[#a0a0a0] mb-4 leading-relaxed';

export const LIST_CLASSES = 'text-[#a0a0a0] mb-4 space-y-2';

export const NEXT_ARTICLE_BUTTON_CLASSES =
  'inline-flex items-center px-4 py-2 bg-[#fff] text-black hover:bg-white/80 font-semibold rounded-md transition-colors duration-200 hover:translate-y-[-1px] active:translate-y-[1px]';

export const NEXT_ARTICLE_CARD_CLASSES =
  'mt-8 md:mt-12 p-4 md:p-6 rounded-lg bg-[#1a1a1a] border border-[#2a2a2a]';

// ---------------------------------------------------------------------------
// Fenced code blocks
// ---------------------------------------------------------------------------

export const CODE_BLOCK_CLASSES =
  'bg-[#1a1a1a] border border-[#2a2a2a] p-4 rounded-lg overflow-x-auto mb-4';

export const CODE_BLOCK_COPY_CLASSES =
  'text-[#666] bg-[#2a2a2a] hover:bg-[#3a3a3a] hover:text-white';

export const CODE_BLOCK_COPY_ACTIVE_CLASSES =
  'text-[#10b981] bg-[#10b981]/10';

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

export const TABLE_WRAPPER_CLASSES =
  'overflow-x-auto my-4 md:my-6 -mx-4 px-4 sm:mx-0 sm:px-0';

export const TABLE_CLASSES =
  'min-w-full bg-[#1a1a1a] border border-[#2a2a2a] rounded-lg';

export const TABLE_HEAD_CLASSES = 'bg-[#111]';

export const TABLE_ROW_CLASSES = 'border-b border-[#2a2a2a]';

export const TABLE_HEADER_CLASSES =
  'px-3 md:px-4 py-2 md:py-3 text-left text-[#b0b0b0] font-semibold text-sm md:text-base';

export const TABLE_CELL_CLASSES =
  'px-3 md:px-4 py-2 md:py-3 text-[#a0a0a0] text-sm md:text-base';

// ---------------------------------------------------------------------------
// Mermaid diagrams
// ---------------------------------------------------------------------------

export const MERMAID_CONTAINER_CLASSES =
  'my-6 p-4 bg-[#1a1a1a] rounded-lg border border-[#2a2a2a] overflow-x-auto';

// ---------------------------------------------------------------------------
// Horizontal rules
// ---------------------------------------------------------------------------

export const HR_CLASSES = 'my-6 md:my-8 border-[#2a2a2a]';

// ---------------------------------------------------------------------------
// Search modal
// ---------------------------------------------------------------------------

export const SEARCH_DIALOG_BORDER = '1px solid rgba(255, 255, 255, 0.1)';

// ---------------------------------------------------------------------------
// Sidebar badges – semantic colors are preserved intentionally
// ---------------------------------------------------------------------------

export const SIDEBAR_BADGE_COLORS: Record<string, string> = {
  Popular: 'bg-[#10b981]/10 text-[#10b981]',
  Beta: 'bg-white/10 text-white/20',
  New: 'bg-white/10 text-white',
  Deprecated: 'bg-[#ef4444]/10 text-[#ef4444]',
  Experimental: 'bg-white/10 text-white/60',
};

export const SIDEBAR_BADGE_DEFAULT = 'bg-white/10 text-white/60';

export const SIDEBAR_ACTIVE_CLASSES = 'bg-white/10 text-white';

export const SIDEBAR_ITEM_CLASSES = 'text-[#a0a0a0]';
