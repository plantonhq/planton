import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Utility function to merge Tailwind CSS classes with proper precedence handling
 */
export function cn(...inputs: (string | undefined | null | false)[]): string {
  // Use the ES6 imports that are already imported at the top
  try {
    return twMerge(clsx(inputs));
  } catch (_e) {
    return inputs.filter(Boolean).join(' ');
  }
}

/**
 * Formats a date string into a human-readable format.
 * 
 * @param dateString - The date string to format (ISO format: YYYY-MM-DD)
 * @returns A formatted date string (e.g., "July 21, 2025")
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

/**
 * Formats a date string into a short, compact format.
 * 
 * @param dateString - The date string to format (ISO format: YYYY-MM-DD)
 * @returns A formatted date string (e.g., "Jul 21, 2025")
 */
export function formatShortDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
} 

/**
 * Removes markdown file extensions (.md) from a slug.
 * 
 * @param slug - The slug string to clean
 * @returns The slug without markdown file extensions
 */
export function cleanSlug(slug: string): string {
  return slug.replace(/\.md$/i, '');
}

/**
 * Generates a clean text excerpt from markdown content by removing markdown symbols and formatting.
 * This function is safe to use on both server and client side.
 * 
 * @param content - The markdown content to generate excerpt from
 * @param maxLength - Maximum length of the excerpt (default: 500)
 * @returns Clean text excerpt without markdown symbols
 */
export function generateExcerptFromContent(content: string, maxLength: number = 500): string {
  // Remove frontmatter
  const contentWithoutFrontmatter = content.replace(/^---[\s\S]*?---/, '');

  // Remove markdown symbols and formatting
  const cleanText = contentWithoutFrontmatter
    // Remove code blocks
    .replace(/```[\s\S]*?```/g, '')
    // Remove inline code
    .replace(/`([^`]+)`/g, '$1')
    // Remove entire heading lines (headings are structural, not prose content for excerpts)
    .replace(/^#{1,6}\s+.*$/gm, '')
    // Remove bold/italic
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    // Remove links but keep text
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    // Remove images
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    // Remove HTML tags
    .replace(/<[^>]*>/g, '')
    // Remove horizontal rules
    .replace(/^[-*_]{3,}$/gm, '')
    // Remove blockquotes
    .replace(/^>\s+/gm, '')
    // Remove list markers
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    // Remove emphasis
    .replace(/_{1,2}([^_]+)_{1,2}/g, '$1')
    // Remove strikethrough
    .replace(/~~([^~]+)~~/g, '$1')
    // Remove tables
    .replace(/\|.*\|/g, '')
    // Remove multiple newlines and spaces
    .replace(/\n\s*\n/g, '\n')
    .replace(/\s+/g, ' ')
    .trim();

  // If content is shorter than maxLength, return as is
  if (cleanText.length <= maxLength) {
    return cleanText;
  }

  // Truncate to maxLength and add ellipsis if needed
  const truncated = cleanText.substring(0, maxLength);
  const lastSpace = truncated.lastIndexOf(' ');

  if (lastSpace > maxLength * 0.8) {
    // If we can break at a word boundary
    return truncated.substring(0, lastSpace) + '...';
  }

  return truncated + '...';
}
