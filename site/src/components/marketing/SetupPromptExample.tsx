import { Fragment, type ReactNode } from 'react';
import { SETUP_PROMPT_EXAMPLE as P } from '@/data/homepage';
import styles from './homepage.module.css';

/**
 * Placeholders (the visitor's own code, and `<step>` the agent fills in) are
 * marked alike, and a path or URL never breaks across lines.
 */
function renderText(text: string): ReactNode {
  return text.split(/(\{code\}|<step>|\S*\/\S*)/).map((part, i) => {
    if (part === '{code}' || part === '<step>')
      return (
        <span key={i} className={styles.promptCode}>
          {part === '{code}' ? P.code : part}
        </span>
      );
    if (part.includes('/'))
      return (
        <span key={i} className={styles.promptToken}>
          {part}
        </span>
      );
    return <Fragment key={i}>{part}</Fragment>;
  });
}

/** One line of the example prompt, with the ask-first sentence in full ink. */
function renderLine(line: string): ReactNode {
  const at = line.indexOf(P.emphasis);
  if (at < 0) return renderText(line);
  return (
    <>
      {renderText(line.slice(0, at))}
      <strong>{P.emphasis}</strong>
      {renderText(line.slice(at + P.emphasis.length))}
    </>
  );
}

/**
 * The setup prompt as words a visitor reads, never a prompt they copy: the
 * real one carries a personal single-use code, so there is no copy button.
 */
export function SetupPromptExample({ labelledBy }: { labelledBy: string }) {
  return (
    <div className={styles.prompt} role="group" aria-labelledby={labelledBy}>
      {P.lines.map((line, i) => {
        // Wrapped lines hang under their own text: a command keeps its indent, and a
        // numbered step's continuation lines up after its number.
        const text = line.trimStart();
        const lead = line.length - text.length;
        const marker = /^\d+\. /.test(text) ? 3 : 0;
        return (
          <p key={i} style={{ paddingLeft: `${lead + marker}ch`, textIndent: `${-marker}ch` }}>
            {text ? renderLine(text) : ' '}
          </p>
        );
      })}
    </div>
  );
}
