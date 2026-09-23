'use client';

import React from 'react';
import { ContentCopy } from '@mui/icons-material';
import MermaidDiagram from './MermaidDiagram';
import {
  CODE_BLOCK_CLASSES,
  CODE_BLOCK_COPY_CLASSES,
  CODE_BLOCK_COPY_ACTIVE_CLASSES,
} from '@/theme/docs';

interface CodeBlockProps {
  children: React.ReactNode;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({ children }) => {
  const preRef = React.useRef<HTMLPreElement>(null);
  const [copied, setCopied] = React.useState(false);

  const checkForMermaid = () => {
    if (React.isValidElement(children)) {
      const codeElement = children as React.ReactElement<{ className?: string; children?: React.ReactNode }>;
      if (codeElement.props?.className) {
        const className = codeElement.props.className;
        if (typeof className === 'string' && className.includes('language-mermaid')) {
          const codeContent = codeElement.props.children;
          const content = Array.isArray(codeContent) ? codeContent.join('') : String(codeContent);
          return content;
        }
      }
    }
    return null;
  };

  const mermaidContent = checkForMermaid();

  if (mermaidContent) {
    return <MermaidDiagram chart={mermaidContent} />;
  }

  const handleCopy = async () => {
    try {
      if (preRef.current) {
        const codeElement = preRef.current.querySelector('code');
        if (codeElement && codeElement.textContent) {
          await navigator.clipboard.writeText(codeElement.textContent);
          setCopied(true);
          setTimeout(() => setCopied(false), 2000);
        }
      }
    } catch (err) {
      console.error('Failed to copy code: ', err);
    }
  };

  return (
    <div className="relative group">
      <pre ref={preRef} className={CODE_BLOCK_CLASSES}>
        {children}
      </pre>
      <button
        onClick={handleCopy}
        className={`absolute top-2 right-2 p-2 rounded transition-all duration-200 opacity-0 group-hover:opacity-100 ${
          copied ? CODE_BLOCK_COPY_ACTIVE_CLASSES : CODE_BLOCK_COPY_CLASSES
        }`}
        title={copied ? "Copied!" : "Copy code"}
      >
        {copied ? (
          <div className="h-4 w-4 flex items-center justify-center">
            <span className="text-xs">✓</span>
          </div>
        ) : (
          <ContentCopy className="h-4 w-4" />
        )}
      </button>
    </div>
  );
};
