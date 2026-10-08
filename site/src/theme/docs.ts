/**
 * Centralized design tokens for the documentation system.
 *
 * Every docs component imports class strings from here instead of
 * hardcoding Tailwind color classes. This guarantees the monochrome
 * palette is enforced in one place and makes future palette changes
 * a single-file edit.
 *
 * Docs pages share the public site's light canvas, text, and border palette.
 * Code examples and diagrams may use their own contained visual treatments.
 */

// ---------------------------------------------------------------------------
// Content – Markdown / MDX body
// ---------------------------------------------------------------------------

export const LINK_CLASSES =
  'text-[#2457a6] hover:text-[#171717] underline underline-offset-2 decoration-[#999996] hover:decoration-[#595959] break-words';

export const TAG_CLASSES =
  'px-2 md:px-3 py-1 bg-[#e3e3e0] text-[#454545] text-xs md:text-sm font-medium rounded-full border border-[#c7c7c4]';

export const BLOCKQUOTE_CLASSES =
  'border-l-2 border-[#999996] pl-4 py-3 my-5 bg-[#eeeeeb] rounded-r text-[#454545]';

export const BLOCKQUOTE_WARNING_CLASSES =
  'border-l-2 border-[#ef4444]/40 pl-4 py-3 my-5 bg-[#eeeeeb] rounded-r text-[#454545]';

export const INLINE_CODE_CLASSES =
  'bg-[#e3e3e0] text-[#171717] rounded text-sm break-words';

// Size and weight carry the structural hierarchy throughout dense docs.
export const HEADING_H1_CLASSES =
  'text-xl sm:text-2xl md:text-3xl font-semibold text-[#171717] mt-6 md:mt-8 mb-3 md:mb-4';

export const HEADING_H2_CLASSES =
  'text-lg sm:text-xl md:text-2xl font-semibold text-[#171717] mt-5 md:mt-6 mb-2 md:mb-3';

export const HEADING_H3_CLASSES =
  'text-base sm:text-lg md:text-xl font-semibold text-[#171717] mt-4 md:mt-5 mb-2';

export const HEADING_H4_CLASSES =
  'text-base md:text-lg font-semibold text-[#171717] mt-3 md:mt-4 mb-2';

export const HEADING_H5_CLASSES =
  'text-sm md:text-base font-semibold text-[#171717] mt-3 mb-2';

export const HEADING_H6_CLASSES =
  'text-sm font-semibold text-[#171717] mt-2 mb-1';

export const PARAGRAPH_CLASSES = 'text-[#454545] mb-4 leading-relaxed';

export const LIST_CLASSES = 'text-[#454545] mb-4 space-y-2';

export const NEXT_ARTICLE_BUTTON_CLASSES =
  'inline-flex items-center px-4 py-2 bg-[#171717] text-white hover:bg-[#333333] font-semibold rounded-md transition-colors duration-200 hover:translate-y-[-1px] active:translate-y-[1px]';

export const NEXT_ARTICLE_CARD_CLASSES =
  'mt-8 md:mt-12 p-4 md:p-6 rounded-lg bg-[#eeeeeb] border border-[#c7c7c4]';

// ---------------------------------------------------------------------------
// Fenced code blocks
// ---------------------------------------------------------------------------

export const CODE_BLOCK_CLASSES =
  'bg-[#eeeeeb] border border-[#c7c7c4] p-4 rounded-lg overflow-x-auto mb-4';

export const CODE_BLOCK_COPY_CLASSES =
  'text-[#616161] bg-[#e3e3e0] hover:bg-[#c7c7c4] hover:text-[#171717]';

export const CODE_BLOCK_COPY_ACTIVE_CLASSES =
  'text-[#10b981] bg-[#10b981]/10';

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

export const TABLE_WRAPPER_CLASSES =
  'overflow-x-auto my-4 md:my-6 -mx-4 px-4 sm:mx-0 sm:px-0';

export const TABLE_CLASSES =
  'min-w-full bg-[#eeeeeb] border border-[#c7c7c4] rounded-lg';

export const TABLE_HEAD_CLASSES = 'bg-[#eeeeeb]';

export const TABLE_ROW_CLASSES = 'border-b border-[#c7c7c4]';

export const TABLE_HEADER_CLASSES =
  'px-3 md:px-4 py-2 md:py-3 text-left text-[#171717] font-semibold text-sm md:text-base';

export const TABLE_CELL_CLASSES =
  'px-3 md:px-4 py-2 md:py-3 text-[#454545] text-sm md:text-base';

// ---------------------------------------------------------------------------
// Mermaid diagrams
// ---------------------------------------------------------------------------

export const MERMAID_CONTAINER_CLASSES =
  'my-6 p-4 bg-[#eeeeeb] rounded-lg border border-[#c7c7c4] overflow-x-auto';

// ---------------------------------------------------------------------------
// Horizontal rules
// ---------------------------------------------------------------------------

export const HR_CLASSES = 'my-6 md:my-8 border-[#c7c7c4]';

// ---------------------------------------------------------------------------
// Search modal
// ---------------------------------------------------------------------------

export const SEARCH_DIALOG_BORDER = '1px solid rgba(23, 23, 23, 0.12)';

// ---------------------------------------------------------------------------
// Sidebar badges – semantic colors are preserved intentionally
// ---------------------------------------------------------------------------

export const SIDEBAR_BADGE_COLORS: Record<string, string> = {
  Popular: 'bg-[#10b981]/10 text-[#10b981]',
  Beta: 'bg-[#e3e3e0] text-[#171717]/20',
  New: 'bg-[#e3e3e0] text-[#171717]',
  Deprecated: 'bg-[#ef4444]/10 text-[#ef4444]',
  Experimental: 'bg-[#e3e3e0] text-[#454545]',
};

export const SIDEBAR_BADGE_DEFAULT = 'bg-[#e3e3e0] text-[#454545]';

export const SIDEBAR_ACTIVE_CLASSES = 'bg-[#e3e3e0] text-[#171717]';

export const SIDEBAR_ITEM_CLASSES = 'text-[#454545]';
