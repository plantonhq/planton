'use client';

import React, { useState, useEffect, useDeferredValue, useRef, useCallback } from 'react';
import { Dialog, Typography, Box } from '@mui/material';
import { Search as SearchIcon, InfoOutlined as InfoIcon } from '@mui/icons-material';
import { useRouter } from 'next/navigation';
import { addBasePath } from 'next/dist/client/add-base-path';

// ---------------------------------------------------------------------------
// Pagefind types
// ---------------------------------------------------------------------------

type PagefindOptions = {
  baseUrl?: string;
};

declare global {
  interface Window {
    pagefind?: {
      options: (opts: PagefindOptions) => Promise<void>;
      debouncedSearch: <T>(query: string) => Promise<{
        results: Array<{ data: () => Promise<T> }>;
      } | null>;
    };
  }
}

type PagefindResult = {
  excerpt: string;
  meta: { title: string };
  url: string;
  sub_results: {
    excerpt: string;
    title: string;
    url: string;
  }[];
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const DEV_SEARCH_NOTICE = (
  <Box sx={{ p: 2, textAlign: 'left' }}>
    <Typography variant="body2" sx={{ mb: 1 }}>
      Search isn&apos;t available in development because Pagefind indexes built
      HTML files instead of markdown source files.
    </Typography>
    <Typography variant="body2">
      To test search during development, run <code>yarn build</code> and then{' '}
      <code>yarn start</code>.
    </Typography>
  </Box>
);

async function importPagefind() {
  const pagefindPath = addBasePath('/_pagefind/pagefind.js');
  window.pagefind = (await import(
    /* webpackIgnore: true */ pagefindPath
  )) as typeof window.pagefind;
  await window.pagefind!.options({ baseUrl: '/' });
}

/**
 * Flatten the grouped Pagefind results into a single ordered list so that
 * arrow-key navigation can use a simple numeric index.
 */
function flattenResults(results: PagefindResult[]) {
  const flat: { pageTitle: string; subResult: PagefindResult['sub_results'][0] }[] = [];
  for (const result of results) {
    for (const sub of result.sub_results) {
      flat.push({ pageTitle: result.meta.title, subResult: sub });
    }
  }
  return flat;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

interface SearchModalProps {
  open: boolean;
  onClose: () => void;
}

export const SearchModal: React.FC<SearchModalProps> = ({ open, onClose }) => {
  const [query, setQuery] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | React.ReactElement>('');
  const [results, setResults] = useState<PagefindResult[]>([]);
  const [activeIndex, setActiveIndex] = useState(-1);

  const deferredQuery = useDeferredValue(query);
  const inputRef = useRef<HTMLInputElement>(null);
  const resultsRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Reset state when the modal opens / closes
  useEffect(() => {
    if (open) {
      // Focus is handled by Dialog's autoFocus, but we reset state here
      setQuery('');
      setResults([]);
      setError('');
      setIsLoading(false);
      setActiveIndex(-1);
    }
  }, [open]);

  // -----------------------------------------------------------------------
  // Pagefind search
  // -----------------------------------------------------------------------

  useEffect(() => {
    if (!open) return;

    const search = async (value: string) => {
      if (!value) {
        setResults([]);
        setError('');
        setIsLoading(false);
        return;
      }

      setIsLoading(true);

      if (!window.pagefind) {
        try {
          await importPagefind();
        } catch (err) {
          const message =
            err instanceof Error
              ? process.env.NODE_ENV !== 'production' &&
                err.message.includes('Failed to fetch')
                ? DEV_SEARCH_NOTICE
                : `${err.constructor.name}: ${err.message}`
              : String(err);
          setError(message);
          setIsLoading(false);
          return;
        }
      }

      const response = await window.pagefind!.debouncedSearch<PagefindResult>(value);
      if (!response) return;

      const data = await Promise.all(response.results.map((r) => r.data()));
      setIsLoading(false);
      setError('');
      setResults(
        data.map((d) => ({
          ...d,
          sub_results: d.sub_results.map((r) => {
            const url = r.url.replace(/\.html$/, '').replace(/\.html#/, '#');
            return { ...r, url };
          }),
        })),
      );
      setActiveIndex(-1);
    };

    search(deferredQuery);
  }, [deferredQuery, open]);

  // -----------------------------------------------------------------------
  // Navigation
  // -----------------------------------------------------------------------

  const flat = flattenResults(results);

  const navigateTo = useCallback(
    (sub: PagefindResult['sub_results'][0]) => {
      const [url, hash] = sub.url.split('#');
      const isSamePage = location.pathname === url;

      if (isSamePage && hash) {
        location.href = `#${hash}`;
      } else {
        router.push(sub.url);
      }

      onClose();
    },
    [router, onClose],
  );

  // -----------------------------------------------------------------------
  // Keyboard handling inside the modal
  // -----------------------------------------------------------------------

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (flat.length === 0) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActiveIndex((prev) => (prev < flat.length - 1 ? prev + 1 : 0));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActiveIndex((prev) => (prev > 0 ? prev - 1 : flat.length - 1));
      } else if (e.key === 'Enter' && activeIndex >= 0) {
        e.preventDefault();
        navigateTo(flat[activeIndex].subResult);
      }
    },
    [flat, activeIndex, navigateTo],
  );

  // Scroll the active result into view
  useEffect(() => {
    if (activeIndex < 0 || !resultsRef.current) return;
    const el = resultsRef.current.querySelector(`[data-result-index="${activeIndex}"]`);
    el?.scrollIntoView({ block: 'nearest' });
  }, [activeIndex]);

  // -----------------------------------------------------------------------
  // Derived state
  // -----------------------------------------------------------------------

  const hasQuery = deferredQuery.length > 0;
  const showResults = !error && !isLoading && results.length > 0;
  const showEmpty = !error && !isLoading && hasQuery && results.length === 0;

  // -----------------------------------------------------------------------
  // Render
  // -----------------------------------------------------------------------

  // Track which page group each flat index belongs to, so we can render
  // group headers inline while keeping a single flat index for keyboard nav.
  let flatIdx = 0;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth={false}
      // Position the dialog near the top of the viewport (not dead-center)
      sx={{
        '& .MuiDialog-container': {
          alignItems: 'flex-start',
          pt: { xs: '10vh', md: '15vh' },
        },
        '& .MuiDialog-paper': {
          width: '100%',
          maxWidth: 640,
          mx: 2,
          bgcolor: '#0a0a0a',
          border: '1px solid #242424',
          borderRadius: '12px',
          boxShadow: '0 25px 60px rgba(0, 0, 0, 0.6)',
          overflow: 'hidden',
        },
        '& .MuiBackdrop-root': {
          bgcolor: 'rgba(0, 0, 0, 0.6)',
          backdropFilter: 'blur(4px)',
        },
      }}
    >
      {/* ── Search input ─────────────────────────────────────────────── */}
      <div
        className="flex items-center gap-3 px-4 py-3 border-b border-border"
        onKeyDown={handleKeyDown}
      >
        <SearchIcon sx={{ fontSize: 20, color: '#a1a1a1', flexShrink: 0 }} />
        <input
          ref={inputRef}
          type="text"
          autoFocus
          autoComplete="off"
          spellCheck={false}
          placeholder="Search documentation..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          className="
            flex-1
            bg-transparent
            text-foreground text-base
            placeholder:text-muted-foreground
            outline-none
            border-none
            p-0
          "
        />
        <kbd
          className="
            hidden sm:inline-flex items-center
            px-1.5 py-0.5
            text-[11px] font-mono leading-none
            text-muted-foreground
            border border-border rounded
          "
        >
          ESC
        </kbd>
      </div>

      {/* ── Results area ─────────────────────────────────────────────── */}
      {hasQuery && (
        <div
          ref={resultsRef}
          className="overflow-y-auto overscroll-contain"
          style={{ maxHeight: '60vh' }}
        >
          {error ? (
            /* Error state */
            <div className="flex gap-2 items-start p-4">
              <InfoIcon sx={{ color: 'error.main', fontSize: 20, mt: 0.5 }} />
              <div>
                <p className="text-destructive font-semibold text-sm mb-1">
                  Failed to load search index
                </p>
                {typeof error === 'string' ? (
                  <p className="text-muted-foreground text-sm">{error}</p>
                ) : (
                  error
                )}
              </div>
            </div>
          ) : isLoading ? (
            /* Loading state */
            <div className="flex items-center justify-center gap-2 p-6">
              <div className="w-4 h-4 border-2 border-border border-t-foreground rounded-full animate-spin" />
              <span className="text-muted-foreground text-sm">Searching...</span>
            </div>
          ) : showEmpty ? (
            /* Empty state */
            <div className="p-6 text-center">
              <p className="text-muted-foreground text-sm">
                No results found for &ldquo;{deferredQuery}&rdquo;
              </p>
            </div>
          ) : showResults ? (
            /* Results list */
            <div className="py-2">
              {results.map((result) => {
                const groupHeader = (
                  <div
                    key={`header-${result.url}`}
                    className="px-4 pt-3 pb-1.5"
                  >
                    <span className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider">
                      {result.meta.title}
                    </span>
                  </div>
                );

                const items = result.sub_results.map((sub) => {
                  const idx = flatIdx++;
                  const isActive = idx === activeIndex;

                  return (
                    <button
                      key={sub.url}
                      data-result-index={idx}
                      type="button"
                      onClick={() => navigateTo(sub)}
                      onMouseEnter={() => setActiveIndex(idx)}
                      className={`
                        w-full text-left px-4 py-2.5 mx-0
                        flex flex-col gap-0.5
                        cursor-pointer
                        transition-colors duration-100
                        ${isActive
                          ? 'bg-secondary'
                          : 'bg-transparent hover:bg-secondary'
                        }
                      `}
                    >
                      <span
                        className={`text-sm font-medium ${
                          isActive ? 'text-foreground' : 'text-muted-foreground'
                        }`}
                      >
                        {sub.title}
                      </span>
                      <span
                        className="text-xs text-muted-foreground line-clamp-2 [&_mark]:bg-accent [&_mark]:text-foreground [&_mark]:font-semibold [&_mark]:px-0.5 [&_mark]:rounded-sm"
                        dangerouslySetInnerHTML={{ __html: sub.excerpt }}
                      />
                    </button>
                  );
                });

                return (
                  <div key={result.url}>
                    {groupHeader}
                    {items}
                  </div>
                );
              })}
            </div>
          ) : null}
        </div>
      )}

      {/* ── Footer with keyboard hints ───────────────────────────────── */}
      {hasQuery && (
        <div className="flex items-center gap-4 px-4 py-2 border-t border-border bg-background/50">
          <FooterHint label="to select" glyph="↵" />
          <FooterHint label="to navigate" glyph="↑↓" />
          <FooterHint label="to close" glyph="esc" />
        </div>
      )}
    </Dialog>
  );
};

// ---------------------------------------------------------------------------
// Footer hint (tiny presentational helper — not worth a separate file)
// ---------------------------------------------------------------------------

function FooterHint({ glyph, label }: { glyph: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
      <kbd className="inline-flex items-center px-1 py-0.5 font-mono text-[10px] border border-border rounded leading-none">
        {glyph}
      </kbd>
      {label}
    </span>
  );
}
