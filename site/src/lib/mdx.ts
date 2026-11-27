import matter from 'gray-matter';

/**
 * The docs' Markdown helpers: a page's authors, and the frontmatter rebuilt
 * for the page's Markdown twin. (The site's other content readers live with
 * their content: the changelog's in ./changelog.ts.)
 */
export interface Author {
  name: string;
  title: string;
  bio?: string;
  profilePicture?: string;
  twitter?: string;
  github?: string;
  linkedin?: string;
  website?: string;
}

export class MDXParser {
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

    return `---\n${frontmatterFields.join('\n')}\n---\n\n${content}`;
  }
}
