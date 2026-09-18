'use client';

import React, { useEffect, useRef, useState } from 'react';
import mermaid from 'mermaid';
import { MERMAID_CONTAINER_CLASSES } from '@/theme/docs';

interface MermaidDiagramProps {
  chart: string;
  className?: string;
}

const MermaidDiagram: React.FC<MermaidDiagramProps> = ({ chart, className = '' }) => {
  const [svg, setSvg] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current && svg) {
      const svgElement = containerRef.current.querySelector('svg');
      if (svgElement) {
        svgElement.style.backgroundColor = '#1a1a1a';
        svgElement.style.maxWidth = '100%';
        svgElement.style.height = 'auto';

        const originalWidth = svgElement.getAttribute('width');
        if (originalWidth && parseInt(originalWidth) > 800) {
          svgElement.style.transform = 'scale(0.8)';
          svgElement.style.transformOrigin = 'top left';
        }

        const selectors = [
          { selector: '.node rect, .node circle, .node ellipse, .node polygon', styles: { fill: '#2a2a2a', stroke: '#ededed', strokeWidth: '1px' } },
          { selector: '.node .label, text', styles: { fill: '#ededed', fontSize: '12px' } },
          { selector: '.edgeLabel rect', styles: { fill: '#2a2a2a' } },
          { selector: '.edgeLabel text', styles: { fill: '#ededed', fontSize: '11px' } },
          { selector: '.cluster rect', styles: { fill: '#1a1a1a', stroke: '#3a3a3a' } },
          { selector: 'path.path', styles: { stroke: '#666666', strokeWidth: '1px' } },
          { selector: '.flowchart-link', styles: { stroke: '#666666', strokeWidth: '1px' } },
          { selector: '.actor', styles: { fill: '#2a2a2a', stroke: '#ededed' } },
          { selector: '.actor-line', styles: { stroke: '#666666' } },
          { selector: '.messageLine0, .messageLine1', styles: { stroke: '#666666' } },
        ];

        selectors.forEach(({ selector, styles }) => {
          svgElement.querySelectorAll<SVGElement>(selector).forEach((el) => {
            Object.assign(el.style, styles);
          });
        });
      }
    }
  }, [svg]);

  useEffect(() => {
    let mounted = true;

    const renderDiagram = async () => {
      try {
        setIsLoading(true);
        setError('');

        mermaid.initialize({
          startOnLoad: false,
          theme: 'dark',
          themeVariables: {
            darkMode: true,
            background: '#1a1a1a',
            primaryColor: '#ededed',
            primaryTextColor: '#ededed',
            primaryBorderColor: '#3a3a3a',
            lineColor: '#666666',
            secondaryColor: '#2a2a2a',
            tertiaryColor: '#111111',
            fontSize: '12px',
            mainBkg: '#2a2a2a',
            secondBkg: '#3a3a3a',
            secondaryTextColor: '#a0a0a0',
            tertiaryTextColor: '#666666',
            edgeLabelBackground: '#2a2a2a',
            classText: '#ededed',
            actorBkg: '#2a2a2a',
            actorBorder: '#ededed',
            actorTextColor: '#ededed',
            actorLineColor: '#666666',
            signalColor: '#ededed',
            signalTextColor: '#ededed',
            messageLine0: '#666666',
            messageLine1: '#666666',
            // Git graph — functional colors for branch differentiation
            git0: '#ef4444',
            git1: '#22c55e',
            git2: '#3b82f6',
            git3: '#f59e0b',
            git4: '#8b5cf6',
            git5: '#ec4899',
            git6: '#06b6d4',
            git7: '#84cc16',
            gitBranchLabel0: '#ededed',
            gitBranchLabel1: '#ededed',
            gitBranchLabel2: '#ededed',
            gitBranchLabel3: '#ededed',
            gitBranchLabel4: '#ededed',
            gitBranchLabel5: '#ededed',
            gitBranchLabel6: '#ededed',
            gitBranchLabel7: '#ededed',
            // Pie chart — functional colors for segment differentiation
            pie1: '#ef4444',
            pie2: '#f59e0b',
            pie3: '#22c55e',
            pie4: '#3b82f6',
            pie5: '#8b5cf6',
            pie6: '#ec4899',
            pie7: '#06b6d4',
            pie8: '#84cc16',
            pieTitleTextSize: '14px',
            pieTitleTextColor: '#ededed',
            pieSectionTextSize: '11px',
            pieSectionTextColor: '#ededed',
            pieLegendTextSize: '11px',
            pieLegendTextColor: '#ededed',
          },
          fontFamily: "'Inter', system-ui, -apple-system, sans-serif",
          flowchart: {
            useMaxWidth: false,
            htmlLabels: true,
            curve: 'basis',
            padding: 10,
            nodeSpacing: 50,
            rankSpacing: 50,
          },
          sequence: {
            useMaxWidth: false,
            showSequenceNumbers: true,
            diagramMarginX: 50,
            diagramMarginY: 10,
            actorMargin: 50,
            width: 150,
            height: 65,
            boxMargin: 10,
            boxTextMargin: 5,
            noteMargin: 10,
            messageMargin: 35,
          },
          gantt: {
            useMaxWidth: false,
            leftPadding: 75,
            gridLineStartPadding: 35,
            fontSize: 11,
          },
          journey: {
            useMaxWidth: false,
            diagramMarginX: 50,
            diagramMarginY: 10,
          },
          timeline: {
            useMaxWidth: false,
            padding: 5,
          },
          gitGraph: {
            useMaxWidth: false,
            mainBranchName: 'main',
            showBranches: true,
            showCommitLabel: true,
            rotateCommitLabel: true,
          },
          er: {
            useMaxWidth: false,
            diagramPadding: 20,
            layoutDirection: 'TB',
            minEntityWidth: 100,
            minEntityHeight: 75,
            entityPadding: 15,
            stroke: '#3a3a3a',
            fill: '#2a2a2a',
            fontSize: 12,
          },
          pie: {
            useMaxWidth: false,
            textPosition: 0.75,
          },
          quadrantChart: {
            useMaxWidth: false,
            chartWidth: 500,
            chartHeight: 400,
          },
          xyChart: {
            useMaxWidth: false,
            width: 700,
            height: 500,
          },
          requirement: {
            useMaxWidth: false,
            rect_fill: '#2a2a2a',
            text_color: '#ededed',
            rect_border_size: '0.5px',
            rect_border_color: '#3a3a3a',
            rect_min_width: 200,
            rect_min_height: 200,
            fontSize: 14,
          },
        });

        const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`;
        const { svg: renderedSvg } = await mermaid.render(id, chart.trim());

        if (mounted) {
          setSvg(renderedSvg);
          setError('');
        }
      } catch (err) {
        console.error('Mermaid rendering error:', err);
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to render diagram');
          setSvg('');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    renderDiagram();

    return () => {
      mounted = false;
    };
  }, [chart]);

  if (isLoading) {
    return (
      <div className={`${MERMAID_CONTAINER_CLASSES} ${className}`}>
        <div className="flex items-center justify-center h-32">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#3a3a3a]"></div>
          <span className="ml-3 text-[#666]">Rendering diagram...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`my-6 p-4 bg-[#ef4444]/5 rounded-lg border border-[#ef4444]/30 ${className}`}>
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-[#ef4444]" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-[#ef4444]">Mermaid Diagram Error</h3>
            <div className="mt-2 text-sm text-[#ef4444]/80">
              <p>{error}</p>
            </div>
            <details className="mt-3">
              <summary className="text-xs text-[#ef4444]/70 cursor-pointer hover:text-[#ef4444]">
                Show diagram source
              </summary>
              <pre className="mt-2 text-xs text-[#a0a0a0] bg-[#1a1a1a] p-2 rounded border border-[#2a2a2a] overflow-x-auto">
                {chart}
              </pre>
            </details>
          </div>
        </div>
      </div>
    );
  }

  if (!svg) {
    return (
      <div className={`${MERMAID_CONTAINER_CLASSES} ${className}`}>
        <div className="text-[#666] text-center">No diagram to display</div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={`${MERMAID_CONTAINER_CLASSES} ${className}`}
    >
      <div
        className="mermaid-diagram flex justify-center"
        dangerouslySetInnerHTML={{ __html: svg }}
      />
    </div>
  );
};

export default MermaidDiagram;
