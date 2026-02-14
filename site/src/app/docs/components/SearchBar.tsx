'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { SearchTrigger } from '@/app/docs/components/SearchTrigger';
import { SearchModal } from '@/app/docs/components/SearchModal';

/**
 * Orchestrator that renders the header search trigger and controls the
 * search modal. Owns the global ⌘K / `/` keyboard shortcut.
 *
 * This is the only component imported by DocsHeader — the public API is
 * unchanged.
 */
export const SearchBar: React.FC = () => {
  const [open, setOpen] = useState(false);

  const handleOpen = useCallback(() => setOpen(true), []);
  const handleClose = useCallback(() => setOpen(false), []);

  // Global keyboard shortcut: ⌘K (Mac) / Ctrl+K (Windows/Linux) or `/`
  useEffect(() => {
    const INPUTS = new Set(['INPUT', 'SELECT', 'BUTTON', 'TEXTAREA']);

    function handleKeyDown(event: KeyboardEvent) {
      // Don't intercept when the user is already typing in an input
      const el = document.activeElement;
      if (
        el &&
        (INPUTS.has(el.tagName) || (el as HTMLElement).isContentEditable)
      ) {
        // Exception: allow ⌘K even inside an input (standard pattern)
        const isCmdK =
          event.key === 'k' &&
          !event.shiftKey &&
          (navigator.userAgent.includes('Mac') ? event.metaKey : event.ctrlKey);
        if (!isCmdK) return;
      }

      const isCmdK =
        event.key === 'k' &&
        !event.shiftKey &&
        (navigator.userAgent.includes('Mac') ? event.metaKey : event.ctrlKey);
      const isSlash = event.key === '/';

      if (isCmdK || isSlash) {
        event.preventDefault();
        setOpen(true);
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <>
      <SearchTrigger onClick={handleOpen} />
      <SearchModal open={open} onClose={handleClose} />
    </>
  );
};
