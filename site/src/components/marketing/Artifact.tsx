import Link from 'next/link';
import type { CSSProperties } from 'react';
import type {
  Artifact as ArtifactRecord,
  Check as CheckRecord,
  ScreenArtifact,
} from '@/data/artifact';
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
 * A screen of the product, flat and untouched, in a hairline frame. A phone
 * gets the record's narrower crop when it has one, so the console's own text
 * stays readable. Sized from the record, so the page never shifts as it loads;
 * a plain img, because the site is a static export and the image is already
 * compressed at twice its shown width.
 */
function Screen({ screen }: { screen: ScreenArtifact }) {
  return (
    // Never wider than half its crop (the phone's own on a phone): the console is shown at most at its own size, never enlarged.
    <div
      className={styles.screen}
      style={
        {
          '--screen-width': `${screen.width / 2}px`,
          '--screen-phone-width': `${(screen.phone ?? screen).width / 2}px`,
        } as CSSProperties
      }
    >
      <picture>
        {screen.phone && (
          <source
            media="(max-width: 760px)"
            srcSet={screen.phone.src}
            width={screen.phone.width}
            height={screen.phone.height}
          />
        )}
        <img
          src={screen.src}
          alt={screen.alt}
          width={screen.width}
          height={screen.height}
          loading="lazy"
          decoding="async"
        />
      </picture>
    </div>
  );
}

/**
 * A real artifact, flat and labelled: a file or a command's output in the
 * code panel, a small table of facts, or a screen of the product. The label
 * says what it is and where it comes from; the artifact states nothing the
 * record does not hold.
 */
export function Artifact({ artifact, id }: { artifact: ArtifactRecord; id: string }) {
  const labelId = `${id}-artifact`;
  return (
    <figure aria-labelledby={labelId}>
      <figcaption id={labelId} className={styles.label}>
        {artifact.label}
      </figcaption>
      {artifact.kind === 'screen' ? (
        <Screen screen={artifact} />
      ) : artifact.kind === 'code' ? (
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

/**
 * An artifact one click away: the proof a doubting reader opens, kept out of
 * the path of everyone else. A native details element, so it opens with the
 * keyboard and in find-in-page, and the Markdown twin prints it open.
 */
export function Folded({ artifact, id }: { artifact: ArtifactRecord; id: string }) {
  return (
    <details className={styles.folded}>
      <summary>
        <span>{artifact.label}</span>
      </summary>
      <Artifact artifact={artifact} id={id} />
    </details>
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
