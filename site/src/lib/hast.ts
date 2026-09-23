/**
 * Typed helpers over the HTML syntax tree react-markdown hands a custom
 * component as its `node`. The markdown renderers (the docs, blog, and
 * tutorials renderer; the changelog body) inspect that tree to decide how
 * to render a paragraph, a blockquote, or a fenced block; this file is the
 * one place they read it, so the shape is named once and no renderer types
 * a node as `any`.
 */
import type { ExtraProps } from 'react-markdown';

export type HastElement = NonNullable<ExtraProps['node']>;
export type HastNode = HastElement['children'][number];

export const isElement = (node: HastNode | HastElement | undefined): node is HastElement => node?.type === 'element';

/** The element children of a node, in order. */
export const elementChildren = (node: HastElement | undefined): HastElement[] => (node?.children ?? []).filter(isElement);

/** The children that carry content: elements, and text that is not only whitespace. */
export const contentChildren = (node: HastElement | undefined): HastNode[] =>
  (node?.children ?? []).filter((c) => (c.type === 'text' ? c.value.trim().length > 0 : true));

/** The text a node contains, concatenated in order, ignoring anything that is not text. */
export const textOf = (node: HastNode | HastElement | undefined): string => {
  if (!node) return '';
  if (node.type === 'text') return node.value;
  if ('children' in node) return node.children.map(textOf).join('');
  return '';
};

/** A string attribute of an element, or undefined when absent or not a string. */
export const attribute = (node: HastElement | undefined, name: string): string | undefined => {
  const value = node?.properties?.[name];
  return typeof value === 'string' ? value : undefined;
};

/** The class names of an element, as react-markdown provides them (an array, or nothing). */
export const classNames = (node: HastElement | undefined): string[] => {
  const value = node?.properties?.className;
  return Array.isArray(value) ? value.map(String) : [];
};

/**
 * A fenced ```mermaid block arrives as <pre><code class="language-mermaid">;
 * this returns its source, without the trailing newline, or undefined for
 * any other <pre>.
 */
export const mermaidSource = (pre: HastElement | undefined): string | undefined => {
  const code = elementChildren(pre)[0];
  if (!code || code.tagName !== 'code' || !classNames(code).includes('language-mermaid')) return undefined;
  return textOf(code).replace(/\n$/, '');
};
