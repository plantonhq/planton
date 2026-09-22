'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { SearchTrigger } from '@/app/(site)/docs/components/SearchTrigger';
import { SearchModal } from '@/app/(site)/docs/components/SearchModal';

/**
 * Orchestrator that renders the header search trigger and controls the
 * search modal. Owns the global Cmd+K / Ctrl+K keyboard shortcut.
 *
 * This is the only component imported by DocsLayout — the public API is
 * unchanged from the previous stub.
 */

interface SearchBarProps {
  /** When provided, the parent can open the modal imperatively. */
  onOpenRef?: React.MutableRefObject<(() => void) | null>;
}

export const SearchBar: React.FC<SearchBarProps> = ({ onOpenRef }) => {
  const [open, setOpen] = useState(false);

  const handleOpen = useCallback(() => setOpen(true), []);
  const handleClose = useCallback(() => setOpen(false), []);

  // Expose handleOpen to the parent via ref (for mobile trigger)
  useEffect(() => {
    if (onOpenRef) {
      onOpenRef.current = handleOpen;
    }
    return () => {
      if (onOpenRef) {
        onOpenRef.current = null;
      }
    };
  }, [onOpenRef, handleOpen]);

  // Global keyboard shortcut: Cmd+K (Mac) / Ctrl+K (Windows/Linux) or "/"
  useEffect(() => {
    const INPUTS = new Set(['INPUT', 'SELECT', 'BUTTON', 'TEXTAREA']);

    function handleKeyDown(event: KeyboardEvent) {
      const el = document.activeElement;

      // Don't intercept when the user is typing in an input — except for Cmd+K
      if (
        el &&
        (INPUTS.has(el.tagName) || (el as HTMLElement).isContentEditable)
      ) {
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
