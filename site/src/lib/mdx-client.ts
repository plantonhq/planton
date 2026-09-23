// Client-safe MDX utilities (no Node.js imports)
import matter from 'gray-matter';
import type { Author } from '@/lib/types-client';

export class MDXParserClient {
  static reconstructMDX(input: string): string {
    const { data, content } = matter(input);
    
    const authorYaml = ((data.author ?? []) as (Author | string)[]).map((author) => {
      if (typeof author === 'string') {
        return `  - ${author}`;
      } else {
        const authorFields = [`    title: ${author.title}`];
        if (author.bio) authorFields.push(`    bio: ${author.bio}`);
        if (author.profilePicture) authorFields.push(`    profilePicture: ${author.profilePicture}`);
        if (author.twitter) authorFields.push(`    twitter: ${author.twitter}`);
        if (author.github) authorFields.push(`    github: ${author.github}`);
        if (author.linkedin) authorFields.push(`    linkedin: ${author.linkedin}`);
        if (author.website) authorFields.push(`    website: ${author.website}`);
        
        return `  - name: ${author.name}\n${authorFields.join('\n')}`;
      }
    }).join('\n');

    const frontmatterFields = [
      `title: "${data.title || ''}"`,
      ...(data.date ? [`date: ${data.date}`] : []),
      `author:\n${authorYaml}`
    ];

    if (data.featuredImage) {
      frontmatterFields.push(`featuredImage: >-\n  ${data.featuredImage}`);
    }
    
    if (data.featuredImageType) {
      frontmatterFields.push(`featuredImageType: ${data.featuredImageType}`);
    }
    
    if (data.featuredVideo) {
      frontmatterFields.push(`featuredVideo: ${data.featuredVideo}`);
    }
    
    if (data.tags && data.tags.length > 0) {
      frontmatterFields.push(`tags:\n${(data.tags as string[]).map((tag) => `  - ${tag}`).join('\n')}`);
    }
    
    if (data.excerpt) {
      frontmatterFields.push(`excerpt: "${data.excerpt}"`);
    }
    
    if (data.hideCopyMarkdown) {
      frontmatterFields.push(`hideCopyMarkdown: true`);
    }
    
    if (data.hideViewMarkdown) {
      frontmatterFields.push(`hideViewMarkdown: true`);
    }

    return `---\n${frontmatterFields.join('\n')}\n---\n\n${content}`;
  }
}
