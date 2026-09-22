import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';
import { BLOG_DIRECTORY, TUTORIALS_DIRECTORY } from '@/lib/constants';
import { Tutorial } from '@/lib/tutorials';
import { generateExcerptFromContent } from '@/lib/utils';

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

export interface BlogPost {
  slug: string;
  title: string;
  date: string;
  author: Author[];
  featuredImage?: string;
  featuredImageType?: string;
  featuredVideo?: string;
  tags: string[];
  excerpt?: string;
  content: string;
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

export function getBlogPostContentBySlug(slug: string): string {
  try {
    const fullPath = path.join(BLOG_DIRECTORY, `${slug}.md`);
    
    if (!fs.existsSync(fullPath)) {
      return '';
    }

    return fs.readFileSync(fullPath, 'utf8');
  } catch (error) {
    console.error('Error reading blog post content for slug: ', slug, error);
    return '';
  }
}

export function getAllBlogPosts(): BlogPost[] {
  try {
    if (!fs.existsSync(BLOG_DIRECTORY)) {
      return [];
    }

    const fileNames = fs.readdirSync(BLOG_DIRECTORY);
    const allPostsData = fileNames
      .filter((fileName) => fileName.endsWith('.md'))
      .map((fileName) => {
        const slug = fileName.replace(/\.md$/, '');
        const fullPath = path.join(BLOG_DIRECTORY, fileName);
        const fileContents = fs.readFileSync(fullPath, 'utf8');
        const { data, content } = matter(fileContents);
        return {
          slug,
          title: data.title,
          date: data.date,
          author: data.author || [],
          featuredImage: data.featuredImage,
          featuredImageType: data.featuredImageType,
          featuredVideo: data.featuredVideo,
          tags: data.tags || [],
          excerpt: data.excerpt,
          content,
        };
      })
      .filter((post) => post !== null)
      .sort((post1, post2) => (post1.date > post2.date ? -1 : 1));

    return allPostsData;
  } catch (error) {
    console.error('Error reading blog posts:', error);
    return [];
  }
}

export function getBlogPostsByTag(tag: string): BlogPost[] {
  const allPosts = getAllBlogPosts();
  return allPosts.filter((post) => post.tags.includes(tag));
}

export function getBlogPostsByAuthor(author: string): BlogPost[] {
  const allPosts = getAllBlogPosts();
  return allPosts.filter((post) => post.author.some(a => a.name === author));
}

export function getTutorialContentBySlug(slug: string): string {
  try {
    const fullPath = path.join(TUTORIALS_DIRECTORY, `${slug}.md`);
    
    if (!fs.existsSync(fullPath)) {
      return '';
    }

    return fs.readFileSync(fullPath, 'utf8');
  } catch (error) {
    console.error('Error reading tutorial content for slug: ', slug, error);
    return '';
  }
}

export function getAllTutorials(): Tutorial[] {
  try {
    if (!fs.existsSync(TUTORIALS_DIRECTORY)) {
      return [];
    }

    const fileNames = fs.readdirSync(TUTORIALS_DIRECTORY);
    const allTutorialsData = fileNames
      .filter((fileName) => fileName.endsWith('.md'))
      .map((fileName) => {
        const slug = fileName.replace(/\.md$/, '');
        const fullPath = path.join(TUTORIALS_DIRECTORY, fileName);
        const fileContents = fs.readFileSync(fullPath, 'utf8');
        const { data, content } = matter(fileContents);
        return {
          slug,
          title: data.title,
          date: data.date,
          author: data.author || [],
          featuredImage: data.featuredImage,
          featuredImageType: data.featuredImageType,
          featuredVideo: data.featuredVideo,
          tags: data.tags || [],
          category: data.category || 'general',
          excerpt: data.excerpt,
          content,
        };
      })
      .filter((tutorial) => tutorial !== null)
      .sort((tutorial1, tutorial2) => (tutorial1.date > tutorial2.date ? -1 : 1));

    return allTutorialsData;
  } catch (error) {
    console.error('Error reading tutorials:', error);
    return [];
  }
}

export function getTutorialsByTag(tag: string): Tutorial[] {
  const allTutorials = getAllTutorials();
  return allTutorials.filter((tutorial) => tutorial.tags.includes(tag));
}

export function getTutorialsByAuthor(author: string): Tutorial[] {
  const allTutorials = getAllTutorials();
  return allTutorials.filter((tutorial) => tutorial.author.some((a) => a.name === author));
}

// Utility functions to get next article
export function getNextBlogPost(currentSlug: string, allPosts: BlogPost[], sortOption: 'date-desc' | 'date-asc' | 'title-asc' | 'title-desc' = 'date-desc'): { title: string; excerpt: string; slug: string } | null {
  // Apply the same sorting logic as the sidebar
  const sortedPosts = [...allPosts];
  switch (sortOption) {
    case 'date-desc':
      sortedPosts.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
      break;
    case 'date-asc':
      sortedPosts.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
      break;
    case 'title-asc':
      sortedPosts.sort((a, b) => a.title.localeCompare(b.title));
      break;
    case 'title-desc':
      sortedPosts.sort((a, b) => b.title.localeCompare(a.title));
      break;
  }
  
  const currentIndex = sortedPosts.findIndex(post => post.slug === currentSlug);
  
  if (currentIndex === -1 || currentIndex === sortedPosts.length - 1) {
    return null;
  }
  
  const nextPost = sortedPosts[currentIndex + 1];
  return {
    title: nextPost.title,
    excerpt: generateExcerptFromContent(nextPost.content || ''),
    slug: nextPost.slug
  };
}

export function getNextTutorial(currentSlug: string, sortOption: 'date-desc' | 'date-asc' | 'title-asc' | 'title-desc' = 'date-desc', allTutorials: Tutorial[]): { title: string; excerpt: string; slug: string } | null {  
  // Apply the same sorting logic as the sidebar
  const sortedTutorials = [...allTutorials];
  switch (sortOption) {
    case 'date-desc':
      sortedTutorials.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
      break;
    case 'date-asc':
      sortedTutorials.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
      break;
    case 'title-asc':
      sortedTutorials.sort((a, b) => a.title.localeCompare(b.title));
      break;
    case 'title-desc':
      sortedTutorials.sort((a, b) => b.title.localeCompare(a.title));
      break;
  }
  
  const currentIndex = sortedTutorials.findIndex(tutorial => tutorial.slug === currentSlug);
  
  if (currentIndex === -1 || currentIndex === sortedTutorials.length - 1) {
    return null;
  }
  
  const nextTutorial = sortedTutorials[currentIndex + 1];
  return {
    title: nextTutorial.title,
    excerpt: generateExcerptFromContent(nextTutorial.content || ''),
    slug: nextTutorial.slug
  };
}


