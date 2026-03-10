'use client';

import React, { useEffect, useRef, useId } from 'react';

interface MermaidBlockProps {
  chart: string;
}

export const MermaidBlock: React.FC<MermaidBlockProps> = ({ chart }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const uniqueId = useId().replace(/:/g, '-');

  useEffect(() => {
    let cancelled = false;

    async function render() {
      const mermaid = (await import('mermaid')).default;

      mermaid.initialize({
        startOnLoad: false,
        theme: 'dark',
        themeVariables: {
          darkMode: true,
          background: '#0f172a',
          primaryColor: '#7c3aed',
          primaryTextColor: '#e2e8f0',
          primaryBorderColor: '#6d28d9',
          secondaryColor: '#1e293b',
          secondaryTextColor: '#cbd5e1',
          secondaryBorderColor: '#475569',
          tertiaryColor: '#1e1b4b',
          lineColor: '#94a3b8',
          textColor: '#e2e8f0',
          mainBkg: '#1e293b',
          nodeBorder: '#6d28d9',
          clusterBkg: '#1e1b4b',
          clusterBorder: '#6d28d9',
          titleColor: '#e2e8f0',
          edgeLabelBackground: '#1e293b',
        },
        flowchart: { curve: 'basis', htmlLabels: true },
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      });

      if (cancelled || !containerRef.current) return;

      try {
        const { svg } = await mermaid.render(`mermaid-${uniqueId}`, chart.trim());
        if (!cancelled && containerRef.current) {
          containerRef.current.innerHTML = svg;
        }
      } catch {
        if (!cancelled && containerRef.current) {
          containerRef.current.textContent = chart;
        }
      }
    }

    render();
    return () => { cancelled = true; };
  }, [chart, uniqueId]);

  return (
    <div
      ref={containerRef}
      className="my-6 flex justify-center overflow-x-auto rounded-lg border border-purple-900/30 bg-slate-900 p-4"
    />
  );
};
