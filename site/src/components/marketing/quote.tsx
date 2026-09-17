/**
 * Testimonials. Two shapes: an inline quote with a left rule, and the card
 * the proof section stacks. The words are always verbatim and attributed to
 * a named person; these components render them and add nothing.
 */
import { Box, Typography } from '@mui/material';
import type { FC } from 'react';

interface QuoteProps {
  text: string;
  author: string;
  role?: string;
  avatar?: string | null;
  companyLogo?: string | null;
  className?: string;
}

export const Quote: FC<QuoteProps> = ({ text, author, role, avatar, companyLogo, className = '' }) => (
  <Box className={`border-l-4 border-white/30 pl-5 py-1.5 ${className}`}>
    <Typography className="text-sm md:text-base text-white italic mb-3">
      &quot;{text}&quot;
    </Typography>
    <Box className="flex items-center gap-3">
      {avatar ? (
        <Box className="w-10 h-10 rounded-full overflow-hidden flex-shrink-0">
          <img src={avatar} alt={author} className="w-full h-full object-cover" />
        </Box>
      ) : companyLogo ? (
        <Box className="w-10 h-10 rounded-full bg-raised p-2 flex items-center justify-center flex-shrink-0">
          <img src={companyLogo} alt={role || ''} className="w-full h-full object-contain brightness-0 invert" />
        </Box>
      ) : (
        <Box className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center text-white font-semibold text-sm flex-shrink-0">
          {author.split(' ').map(n => n[0]).join('')}
        </Box>
      )}
      <Box>
        <Typography className="text-sm text-white font-medium">
          {author}
        </Typography>
        {role && (
          <Typography className="text-xs text-fg-muted">
            {role}
          </Typography>
        )}
      </Box>
    </Box>
  </Box>
);
interface TestimonialCardProps {
  name: string;
  role: string;
  company: string;
  quote: string;
  location?: string;
  avatar?: string | null;
  className?: string;
}

export const TestimonialCard: FC<TestimonialCardProps> = ({
  name,
  role,
  company,
  quote,
  location,
  avatar,
  className = '',
}) => (
  <Box
    className={`
      rounded-xl bg-raised border border-edge
      p-5 transition-all duration-300
      hover:border-edge-hover hover:translate-y-[-2px]
      hover:shadow-lg hover:shadow-black/30
      ${className}
    `}
  >
    <Box className="flex items-center gap-2.5 mb-3">
      {avatar ? (
        <Box className="w-8 h-8 rounded-full overflow-hidden flex-shrink-0">
          <img 
            src={avatar} 
            alt={name}
            className="w-full h-full object-cover"
          />
        </Box>
      ) : (
        <Box className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center text-white font-semibold text-xs flex-shrink-0">
          {name.split(' ').map(n => n[0]).join('')}
        </Box>
      )}
      <Box>
        <Typography className="text-white font-medium text-xs">
          {name}
        </Typography>
        <Typography className="text-fg-muted text-[11px]">
          {role}, {company}
        </Typography>
      </Box>
    </Box>
    
    <Typography className="text-fg-body text-xs leading-relaxed mb-3">
      &quot;{quote}&quot;
    </Typography>
    
    <Box className="flex items-center gap-2 text-[11px] text-fg-muted">
      <span>🏢 {company}</span>
      {location && (
        <>
          <span>•</span>
          <span>🌍 {location}</span>
        </>
      )}
    </Box>
  </Box>
);
