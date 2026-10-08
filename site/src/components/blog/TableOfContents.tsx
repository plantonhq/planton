'use client';

import React, { useMemo } from 'react';

interface HeadingItem {
  id: string;
  text: string;
  level: number;
}

interface TableOfContentsProps {
  content: string;
}

const CODE_KEYWORDS = ['Build', 'Production', 'Install', 'Copy', 'Create', 'User', 'Volume', 'Expose', 'Arg', 'Label', 'Cmd', 'Entrypoint', 'Workdir'];
const CODE_KEYWORDS_SET = new Set(CODE_KEYWORDS);
const CODE_PREFIX_RE = new RegExp(`^(${CODE_KEYWORDS.join('|')})\\s+`);

function extractHeadings(content: string): HeadingItem[] {
  if (!content) return [];

  const lines = content.split('\n');
  const result: HeadingItem[] = [];
  let inCodeBlock = false;
  let inIndentedCodeBlock = false;
  let codeBlockDepth = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmedLine = line.trim();

    if (trimmedLine === '') continue;

    if (trimmedLine.startsWith('```')) {
      if (!inCodeBlock) {
        inCodeBlock = true;
        codeBlockDepth++;
      } else {
        codeBlockDepth--;
        if (codeBlockDepth === 0) inCodeBlock = false;
      }
      continue;
    }

    if (line.startsWith('    ') || line.startsWith('\t')) {
      inIndentedCodeBlock = true;
      continue;
    } else if (inIndentedCodeBlock && trimmedLine !== '') {
      inIndentedCodeBlock = false;
    }

    if (inCodeBlock || inIndentedCodeBlock) continue;

    const headingMatch = trimmedLine.match(/^(#{1,6})\s+(.+)$/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      const text = headingMatch[2].trim();
      if (text.length < 3) continue;
      if (text.split(' ').length === 1 && CODE_KEYWORDS_SET.has(text)) continue;
      if (CODE_PREFIX_RE.test(text)) continue;

      const id = text
        .toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-');

      result.push({ id, text, level });
    }
  }

  return result;
}

const TableOfContents: React.FC<TableOfContentsProps> = ({ content }) => {
  const headings = useMemo(() => extractHeadings(content), [content]);

  if (headings.length === 0) {
    return null;
  }

  // Handle smooth scrolling for navigation links
  const handleNavigationClick = (e: React.MouseEvent<HTMLAnchorElement>, targetId: string) => {
    e.preventDefault();

    // Update the URL hash
    window.history.pushState(null, '', `#${targetId}`);

    // Smooth scroll to the target element with offset for sticky header
    const targetElement = document.getElementById(targetId);
    if (targetElement) {
      const targetPosition = targetElement.offsetTop - 70;
      window.scrollTo({
        top: targetPosition,
        behavior: 'smooth',
      });
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-4">
      <h3 className="text-lg font-semibold text-[#171717] mb-4">On this page</h3>
      <nav className="space-y-2">
        {headings.map((heading) => (
          <a
            key={heading.id}
            href={`#${heading.id}`}
            onClick={(e) => handleNavigationClick(e, heading.id)}
            className={`block text-sm text-[#454545] hover:text-[#171717] transition-colors cursor-pointer ${
              heading.level === 1
                ? 'font-semibold'
                : heading.level === 2
                ? 'ml-0'
                : heading.level === 3
                ? 'ml-4'
                : 'ml-6'
            }`}
          >
            {heading.text}
          </a>
        ))}
      </nav>
    </div>
  );
};

export default TableOfContents;
