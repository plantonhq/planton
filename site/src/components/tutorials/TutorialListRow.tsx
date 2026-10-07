import React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Tutorial } from '@/lib/tutorials';

interface TutorialListRowProps {
  tutorial: Tutorial;
}

export default function TutorialListRow({ tutorial }: TutorialListRowProps) {
  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMs = now.getTime() - date.getTime();
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));
    
    if (diffInDays === 0) return 'today';
    if (diffInDays === 1) return 'yesterday';
    if (diffInDays < 7) return `${diffInDays} days ago`;
    if (diffInDays < 30) return `${Math.floor(diffInDays / 7)} weeks ago`;
    if (diffInDays < 365) return `${Math.floor(diffInDays / 30)} months ago`;
    return `${Math.floor(diffInDays / 365)} years ago`;
  };

  return (
    <div className="flex items-start gap-4 p-4 hover:bg-black/5 transition-colors duration-200">
      {/* Decorative Arrow Icon */}
      <div className="flex-shrink-0 mt-1">
        <div className="text-[#454545]">
          <svg 
            width="16" 
            height="16" 
            viewBox="0 0 16 16" 
            fill="currentColor"
          >
            <path d="M8 1L14 7L8 13L2 7L8 1Z" />
          </svg>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 min-w-0">
        {/* Title and Timestamp Row */}
        <div className="flex items-center gap-3 mb-1">
          <Link 
            href={`/tutorials/${tutorial.slug}`}
            className="text-[#171717] font-semibold text-base hover:text-[#171717] transition-colors duration-200 line-clamp-1 flex-1"
          >
            {tutorial.title}
          </Link>
          <span className="text-[#454545] text-sm flex-shrink-0">
            {formatTimeAgo(tutorial.date)}
          </span>
        </div>

        {/* Author Row */}
        <div className="flex items-center gap-2">
          <span className="text-[#454545] text-sm">by</span>
          <div className="flex items-center gap-2">
            {tutorial.author[0]?.profilePicture && (
              <Image
                src={tutorial.author[0].profilePicture}
                alt={tutorial.author[0].name}
                width={20}
                height={20}
                className="rounded-full"
              />
            )}
            <Link 
              href={`/tutorials?author=${tutorial.author[0]?.name || 'Planton Team'}`}
              className="text-[#171717] hover:text-[#171717] font-medium text-sm transition-colors duration-200"
            >
              {tutorial.author[0]?.name || 'Planton Team'}
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
