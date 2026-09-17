'use client';

import { Box } from '@mui/material';
import { useRouter } from 'next/navigation';
import { useEffect, useCallback, PropsWithChildren } from 'react';

/**
 * A full-screen frame for one task (booking a demo, handing off to the
 * desktop app): no header, no footer, a close button, and Escape returns to
 * the home page. The surfaces that want it mount it in their own layout.
 */
export function FocusFrame({ children }: PropsWithChildren) {
  const router = useRouter();

  const handleClose = useCallback(() => {
    router.push('/');
  }, [router]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        handleClose();
      }
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [handleClose]);

  return (
    <Box className="relative min-h-screen bg-canvas">
      <button
        onClick={handleClose}
        aria-label="Close"
        className="fixed top-5 right-5 z-50 w-10 h-10 rounded-full border border-edge bg-panel flex items-center justify-center text-fg-secondary hover:text-white hover:border-edge-hover transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
      </button>
      {children}
    </Box>
  );
}
