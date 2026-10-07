'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  ChevronDown,
  ExternalLink,
  Link as LinkIcon,
  Check,
  Search,
  X,
} from 'lucide-react';
import ChangelogCategoryBadge from './ChangelogCategoryBadge';
import ChangelogMarkdownBody from './ChangelogMarkdownBody';
import type { ChangelogCategory } from '@/lib/types-client';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface TimelineEntry {
  slug: string;
  title: string;
  date: string;
  category: ChangelogCategory;
  tags: string[];
  excerpt: string;
  content: string;
}

interface ChangelogTimelineProps {
  entries: TimelineEntry[];
}

// ---------------------------------------------------------------------------
// Filter tab configuration
// ---------------------------------------------------------------------------

interface FilterTab {
  id: ChangelogCategory | 'all';
  label: string;
}

const FILTER_TABS: FilterTab[] = [
  { id: 'all', label: 'All' },
  { id: 'feature', label: 'Features' },
  { id: 'improvement', label: 'Improvements' },
  { id: 'fix', label: 'Fixes' },
  { id: 'breaking', label: 'Breaking' },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return dateString;
  }
}

function formatMonthYear(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
    });
  } catch {
    return dateString;
  }
}

/** Group entries by "Month YYYY" while preserving the within-group order. */
function groupByMonth(
  entries: TimelineEntry[],
): { month: string; entries: TimelineEntry[] }[] {
  const groups = new Map<string, TimelineEntry[]>();

  for (const entry of entries) {
    const key = formatMonthYear(entry.date);
    const group = groups.get(key);
    if (group) {
      group.push(entry);
    } else {
      groups.set(key, [entry]);
    }
  }

  return Array.from(groups.entries()).map(([month, items]) => ({
    month,
    entries: items,
  }));
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export default function ChangelogTimeline({
  entries,
}: ChangelogTimelineProps) {
  // -- State ----------------------------------------------------------------
  const [expandedSlug, setExpandedSlug] = useState<string | null>(null);
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);
  const [activeCategory, setActiveCategory] = useState<
    ChangelogCategory | 'all'
  >('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');

  // Debounce search input (300ms)
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // -- Derived data ---------------------------------------------------------

  const filteredEntries = useMemo(() => {
    let result = entries;

    // Category filter
    if (activeCategory !== 'all') {
      result = result.filter((e) => e.category === activeCategory);
    }

    // Text search
    if (debouncedQuery) {
      const q = debouncedQuery.toLowerCase();
      result = result.filter(
        (e) =>
          e.title.toLowerCase().includes(q) ||
          e.excerpt.toLowerCase().includes(q) ||
          e.tags.some((t) => t.toLowerCase().includes(q)),
      );
    }

    return result;
  }, [entries, activeCategory, debouncedQuery]);

  const monthGroups = useMemo(
    () => groupByMonth(filteredEntries),
    [filteredEntries],
  );

  // -- Handlers -------------------------------------------------------------

  const toggleEntry = (slug: string) => {
    setExpandedSlug(expandedSlug === slug ? null : slug);
  };

  const copyLink = async (slug: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const url = `${window.location.origin}/changelog/${slug}`;
    await navigator.clipboard.writeText(url);
    setCopiedSlug(slug);
    setTimeout(() => setCopiedSlug(null), 2000);
  };

  const openInNewPage = (slug: string, e: React.MouseEvent) => {
    e.stopPropagation();
    window.open(`/changelog/${slug}`, '_blank');
  };

  const resetFilters = () => {
    setActiveCategory('all');
    setSearchQuery('');
    setDebouncedQuery('');
  };

  // -- Render ---------------------------------------------------------------

  return (
    <div>
      {/* Search input */}
      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#616161]" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search changelog..."
          className="w-full pl-10 pr-10 py-2.5 bg-[#eeeeeb] border border-[#c7c7c4] rounded-lg text-[#171717] placeholder-[#777774] text-sm focus:outline-none focus:border-[#777774] transition-colors"
        />
        {searchQuery && (
          <button
            onClick={() => {
              setSearchQuery('');
              setDebouncedQuery('');
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 rounded hover:bg-[#e3e3e0] text-[#595959] hover:text-[#171717] transition-colors"
            aria-label="Clear search"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Filter tabs */}
      <div className="flex flex-wrap items-center gap-2 mb-8">
        {FILTER_TABS.map((tab) => {
          const isActive = activeCategory === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveCategory(tab.id)}
              className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-white text-black'
                  : 'bg-[#eeeeeb] text-[#454545] hover:bg-[#e3e3e0] hover:text-[#171717]'
              }`}
            >
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Timeline entries */}
      {monthGroups.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-[#595959] text-lg mb-2">
            No changelog entries match your search.
          </p>
          <button
            onClick={resetFilters}
            className="text-[#171717] hover:text-[#454545] text-sm font-medium transition-colors"
          >
            Clear filters
          </button>
        </div>
      ) : (
        <div className="space-y-10">
          {monthGroups.map((group) => (
            <section key={group.month}>
              {/* Month header */}
              <h2 className="text-xl font-bold text-[#171717] mb-4">
                {group.month}
              </h2>

              {/* Entries */}
              <div className="space-y-1">
                {group.entries.map((entry) => {
                  const isExpanded = expandedSlug === entry.slug;
                  const isCopied = copiedSlug === entry.slug;

                  return (
                    <article key={entry.slug} className="relative">
                      {/* Clickable header */}
                      <div
                        onClick={() => toggleEntry(entry.slug)}
                        className="relative py-6 px-6 -mx-6 rounded-lg hover:bg-black/5 transition-all duration-200 cursor-pointer group"
                      >
                        {/* Row 1: date + badge + action icons */}
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-3">
                            <time
                              dateTime={entry.date}
                              className="text-sm font-medium text-[#595959]"
                            >
                              {formatDate(entry.date)}
                            </time>
                            <ChangelogCategoryBadge
                              category={entry.category}
                            />
                          </div>

                          {/* Action icons (visible on hover) */}
                          <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                              onClick={(e) => copyLink(entry.slug, e)}
                              className="p-1.5 rounded-md hover:bg-black/5 text-[#595959] hover:text-[#171717] transition-colors"
                              title="Copy link"
                            >
                              {isCopied ? (
                                <Check className="w-4 h-4 text-[#10b981]" />
                              ) : (
                                <LinkIcon className="w-4 h-4" />
                              )}
                            </button>
                            <button
                              onClick={(e) =>
                                openInNewPage(entry.slug, e)
                              }
                              className="p-1.5 rounded-md hover:bg-black/5 text-[#595959] hover:text-[#171717] transition-colors"
                              title="Open in new page"
                            >
                              <ExternalLink className="w-4 h-4" />
                            </button>
                            <div
                              className={`p-1.5 text-[#595959] transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`}
                            >
                              <ChevronDown className="w-4 h-4" />
                            </div>
                          </div>
                        </div>

                        {/* Row 2: title + tags */}
                        <div className="flex items-start justify-between gap-4">
                          <h3 className="text-lg font-semibold text-[#171717] group-hover:text-[#171717] transition-colors">
                            {entry.title}
                          </h3>
                          {entry.tags.length > 0 && (
                            <div className="hidden sm:flex flex-shrink-0 flex-wrap gap-1.5">
                              {entry.tags.map((tag) => (
                                <span
                                  key={tag}
                                  className="px-2 py-0.5 bg-[#eeeeeb] text-[#595959] border border-[#c7c7c4] rounded text-xs"
                                >
                                  {tag}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>

                        {/* Row 3: excerpt (only when collapsed) */}
                        {!isExpanded && (
                          <p className="text-[#595959] text-sm leading-relaxed mt-2">
                            {entry.excerpt}
                          </p>
                        )}
                      </div>

                      {/* Expanded content */}
                      {isExpanded && (
                        <div className="px-6 -mx-6 pb-6 pt-2">
                          <ChangelogMarkdownBody
                            content={entry.content}
                          />
                        </div>
                      )}
                    </article>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
