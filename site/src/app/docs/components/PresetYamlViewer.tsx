'use client';

import React, { useState, useCallback, useRef } from 'react';

interface PresetYamlViewerProps {
  yamlContent: string;
  /** Optional label shown above the code block. Defaults to "Manifest". */
  label?: string;
}

/**
 * YAML code viewer with syntax highlighting (via highlight.js classes) and a
 * copy-to-clipboard button.  Built on the same design language as CodeBlock
 * but purpose-built for preset YAML manifests.
 */
export const PresetYamlViewer: React.FC<PresetYamlViewerProps> = ({
  yamlContent,
  label = 'Manifest',
}) => {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(yamlContent).then(() => {
      setCopied(true);
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    });
  }, [yamlContent]);

  return (
    <div className="relative group mb-6">
      {/* Header bar */}
      <div className="flex items-center justify-between px-4 py-2 bg-secondary rounded-t-lg border border-b-0 border-border">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent transition-all"
          aria-label={copied ? 'Copied' : 'Copy YAML'}
        >
          {copied ? (
            <>
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-success">
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span className="text-success">Copied</span>
            </>
          ) : (
            <>
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              <span>Copy</span>
            </>
          )}
        </button>
      </div>

      {/* Code block */}
      <pre className="bg-card rounded-b-lg p-4 overflow-x-auto border border-t-0 border-border text-sm leading-relaxed">
        <code className="language-yaml text-muted-foreground">{yamlContent}</code>
      </pre>
    </div>
  );
};
