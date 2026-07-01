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
          background: '#0a0a0a',
          primaryColor: '#1a1a1a',
          primaryTextColor: '#ededed',
          primaryBorderColor: '#333333',
          secondaryColor: '#111111',
          secondaryTextColor: '#a1a1a1',
          secondaryBorderColor: '#333333',
          tertiaryColor: '#111111',
          lineColor: '#a1a1a1',
          textColor: '#ededed',
          mainBkg: '#1a1a1a',
          nodeBorder: '#333333',
          clusterBkg: '#111111',
          clusterBorder: '#333333',
          titleColor: '#ededed',
          edgeLabelBackground: '#111111',
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
      className="my-6 flex justify-center overflow-x-auto rounded-lg border border-border bg-card p-4"
    />
  );
};
