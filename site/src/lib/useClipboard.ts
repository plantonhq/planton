"use client";

import * as React from "react";

/** Copy-to-clipboard with a transient "copied" flag. Shared across the site. */
export function useClipboard(resetMs = 2000) {
  const [copied, setCopied] = React.useState(false);
  const timer = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  const copy = React.useCallback(
    (text: string) => {
      navigator.clipboard.writeText(text).then(() => {
        setCopied(true);
        if (timer.current) clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), resetMs);
      });
    },
    [resetMs],
  );

  React.useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  return { copied, copy } as const;
}

export default useClipboard;
