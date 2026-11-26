import Link from 'next/link';
import type { Artifact as ArtifactRecord, Check as CheckRecord } from '@/data/artifact';
import styles from './artifact.module.css';

/** An outside page opens in a new tab; a page of this site stays in this one. */
function linkProps(href: string) {
  return href.startsWith('http') ? { target: '_blank', rel: 'noreferrer' } : {};
}

/**
 * One line of a wrapped code artifact. A continuation hangs where the line's
 * own text starts: past a `# ` comment or `$ ` command marker, past a CLI's
 * `✓ `, `✔ ` or `ℹ ` status mark, past a list's `- `, and past a `key: ` so a
 * wrapped value never reads as a sibling key.
 */
function hangOf(line: string): number {
  const lead = line.match(/^ */)?.[0].length ?? 0;
  const rest = line.slice(lead);
  const marker = rest.match(/^(?:[#$✓✔ℹ] |- )/)?.[0].length ?? 0;
  const key = rest.slice(marker).match(/^[\w.-]+: (?=\S)/)?.[0].length ?? 0;
  return lead + marker + key;
}

/**
 * A CLI's status line (`✓`, `✔`, `ℹ`) wraps only between words: each word
 * holding a hyphen (`keycloak-operator`, an arrow `->`) is kept whole by
 * style, never by an inserted character, so the text a reader copies is
 * exactly what the CLI printed. Commands keep their own `\` continuations,
 * and files wrap as written.
 */
function StatusWords({ line }: { line: string }) {
  return line.split(/(\s+)/).map((part, i) =>
    part.includes('-') ? (
      <span key={i} className={styles.whole}>
        {part}
      </span>
    ) : (
      part
    )
  );
}

function WrappedLine({ line }: { line: string }) {
  const hang = hangOf(line);
  return (
    <span className={styles.line} style={{ paddingLeft: `${hang}ch`, textIndent: `-${hang}ch` }}>
      {/^\s*[✓✔ℹ] /.test(line) && line.includes('-') ? (
        <StatusWords line={line} />
      ) : (
        line || '\u00a0'
      )}
    </span>
  );
}

/**
 * A real artifact, flat and labelled: a file or a command's output in the
 * code panel, or a small table of facts. The label says what it is and where
 * it comes from; the artifact states nothing the record does not hold.
 */
export function Artifact({ artifact, id }: { artifact: ArtifactRecord; id: string }) {
  const labelId = `${id}-artifact`;
  return (
    <figure aria-labelledby={labelId}>
      <figcaption id={labelId} className={styles.label}>
        {artifact.label}
      </figcaption>
      {artifact.kind === 'code' ? (
        <div className={styles.code}>
          {artifact.wrap ? (
            <pre className={styles.wrapped}>
              {artifact.lines.map((line, i) => (
                <WrappedLine key={i} line={line} />
              ))}
            </pre>
          ) : (
            <pre>{artifact.lines.join('\n')}</pre>
          )}
        </div>
      ) : (
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                {artifact.columns.map((c) => (
                  <th key={c} scope="col">
                    {c}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {artifact.rows.map((row) => (
                <tr key={row[0]}>
                  {row.map((cell, i) => (
                    <td key={i}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </figure>
  );
}

/** How a reader proves a sentence themselves: a link, a command, or both. */
export function Check({ check }: { check: CheckRecord }) {
  return (
    <p className={styles.check}>
      {check.href ? (
        <Link href={check.href} {...linkProps(check.href)}>
          {check.text}
        </Link>
      ) : (
        check.text
      )}
      {check.command && <code>{check.command}</code>}
    </p>
  );
}
