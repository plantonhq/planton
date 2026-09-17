'use client';

import { Box, Button, ButtonProps, Typography, TypographyProps } from '@mui/material';
import { ComponentProps, FC, ReactNode } from 'react';

// ============================================================================
// DESIGN TOKENS
// ============================================================================

export const colors = {
  // Background layers
  bgPrimary: '#0a0a0a',
  bgSecondary: '#111111',
  bgTertiary: '#1a1a1a',
  bgCard: '#151515',
  bgCardHover: '#1f1f1f',
  
  // Text
  textPrimary: '#ededed',
  textSecondary: '#a0a0a0',
  textMuted: '#666666',
  
  // Semantic (used only for functional indicators)
  semanticGreen: '#10b981',
  semanticRed: '#ef4444',
  semanticAmber: '#f59e0b',
  
  // Borders
  border: '#2a2a2a',
  borderHover: '#3a3a3a',
};

// ============================================================================
// SECTION WRAPPER
// ============================================================================

interface SectionProps {
  children: ReactNode;
  className?: string;
  id?: string;
  variant?: 'default' | 'dark' | 'gradient';
}

export const Section: FC<SectionProps> = ({ 
  children, 
  className = '', 
  id,
  variant: _variant = 'default' 
}) => {
  return (
    <Box 
      component="section" 
      id={id}
      className={`w-full py-12 md:py-16 px-4 md:px-8 overflow-x-hidden bg-[#0a0a0a] ${className}`}
    >
      <Box className="max-w-7xl mx-auto overflow-hidden">
        {children}
      </Box>
    </Box>
  );
};

// ============================================================================
// TYPOGRAPHY
// ============================================================================

export const SectionTitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    variant="h2"
    className={`text-xl md:text-2xl lg:text-3xl font-semibold text-white leading-snug tracking-tight ${className}`}
    {...props}
  />
);

export const SectionSubtitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    className={`text-sm md:text-base text-[#a0a0a0] font-normal mt-3 max-w-2xl ${className}`}
    {...props}
  />
);

export const FeatureTitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    variant="h3"
    className={`text-base md:text-lg font-semibold text-white ${className}`}
    {...props}
  />
);

export const BodyText: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    className={`text-sm text-[#b0b0b0] leading-relaxed ${className}`}
    {...props}
  />
);

// ============================================================================
// BUTTONS
// ============================================================================

export const PrimaryButton: FC<ButtonProps & ComponentProps<'a'>> = ({ 
  className, 
  children,
  ...props 
}) => (
  <Button
    className={`
      bg-[#fff]
      hover:bg-gray-200
      text-black font-medium text-sm
      px-5 py-2.5 rounded-lg
      transition-all duration-300
      hover:-translate-y-0.5
      ${className}
    `}
    {...props}
  >
    {children}
  </Button>
);

export const SecondaryButton: FC<ButtonProps & ComponentProps<'a'>> = ({ 
  className, 
  children,
  ...props 
}) => (
  <Button
    className={`
      bg-transparent
      border border-[#3a3a3a]
      hover:border-white
      text-white font-medium text-sm
      px-5 py-2.5 rounded-lg
      transition-all duration-300
      hover:bg-white/5
      ${className}
    `}
    {...props}
  >
    {children}
  </Button>
);

// ============================================================================
// CARDS
// ============================================================================

interface CardProps {
  children?: ReactNode;
  className?: string;
  hover?: boolean;
  gradient?: boolean;
}

export const Card: FC<CardProps> = ({ 
  children, 
  className = '', 
  hover = true,
}) => (
  <Box
    className={`
      rounded-xl
      bg-[#151515] border border-[#2a2a2a]
      ${hover ? 'hover:border-[#3a3a3a] hover:bg-[#1a1a1a] transition-all duration-300' : ''}
      p-5 md:p-6
      ${className}
    `}
  >
    {children}
  </Box>
);

export const FeatureCard: FC<Omit<CardProps, 'children'> & { icon?: ReactNode; title: string; description: string }> = ({
  icon,
  title,
  description,
  className = '',
}) => (
  <Card className={`${className}`}>
    {icon && (
      <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center mb-3 text-white">
        {icon}
      </Box>
    )}
    <FeatureTitle className="mb-2">{title}</FeatureTitle>
    <BodyText>{description}</BodyText>
  </Card>
);

// ============================================================================
// BADGES & TAGS
// ============================================================================

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'purple';
  className?: string;
}

export const Badge: FC<BadgeProps> = ({ 
  children, 
  variant = 'default',
  className = '' 
}) => {
  const variantClasses = {
    default: 'bg-[#2a2a2a] text-[#a0a0a0] border-[#3a3a3a]',
    success: 'bg-[#10b981]/10 text-[#10b981] border-[#10b981]/30',
    warning: 'bg-[#f59e0b]/10 text-[#f59e0b] border-[#f59e0b]/30',
    purple: 'bg-[#2a2a2a] text-[#a0a0a0] border-[#3a3a3a]',
  };
  
  return (
    <Box
      component="span"
      className={`
        inline-flex items-center
        px-3 py-1.5 rounded-full
        text-xs md:text-sm font-medium
        border
        ${variantClasses[variant]}
        ${className}
      `}
    >
      {children}
    </Box>
  );
};

// ============================================================================
// GRID LAYOUTS
// ============================================================================

interface GridProps {
  children: ReactNode;
  cols?: 1 | 2 | 3 | 4;
  gap?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const Grid: FC<GridProps> = ({ 
  children, 
  cols = 3, 
  gap = 'md',
  className = '' 
}) => {
  const colsClass = {
    1: 'grid-cols-1',
    2: 'grid-cols-1 md:grid-cols-2',
    3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  };
  
  const gapClass = {
    sm: 'gap-4',
    md: 'gap-6',
    lg: 'gap-8',
  };
  
  return (
    <Box className={`grid ${colsClass[cols]} ${gapClass[gap]} ${className}`}>
      {children}
    </Box>
  );
};

// ============================================================================
// ICONS
// ============================================================================

export const CheckIcon = () => (
  <svg className="w-3.5 h-3.5 text-[#10b981]/70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
  </svg>
);

export const XIcon = () => (
  <svg className="w-3.5 h-3.5 text-[#ef4444]/70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M6 18L18 6M6 6l12 12" />
  </svg>
);

export const ArrowRightIcon = () => (
  <svg className="w-4 h-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
  </svg>
);

export const CloudIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
  </svg>
);

export const CodeIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
  </svg>
);

export const RocketIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
  </svg>
);

export const ShieldIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
  </svg>
);

export const GitBranchIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
  </svg>
);

export const CpuIcon = () => (
  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
  </svg>
);

// ============================================================================
// DIVIDER
// ============================================================================

export const Divider: FC<{ className?: string }> = ({ className = '' }) => (
  <Box className={`w-full h-px bg-gradient-to-r from-transparent via-[#2a2a2a] to-transparent ${className}`} />
);

// ============================================================================
// QUOTE / TESTIMONIAL
// ============================================================================

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
        <Box className="w-10 h-10 rounded-full bg-[#1a1a1a] p-2 flex items-center justify-center flex-shrink-0">
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
          <Typography className="text-xs text-[#666]">
            {role}
          </Typography>
        )}
      </Box>
    </Box>
  </Box>
);

// ============================================================================
// METRIC DISPLAY
// ============================================================================

interface MetricProps {
  value: string;
  label: string;
  className?: string;
}

export const Metric: FC<MetricProps> = ({ value, label, className = '' }) => (
  <Box className={`text-center ${className}`}>
    <Typography className="text-2xl md:text-3xl font-bold text-white">
      {value}
    </Typography>
    <Typography className="text-xs md:text-sm text-[#a0a0a0] mt-1">
      {label}
    </Typography>
  </Box>
);

// ============================================================================
// TESTIMONIAL CARD (Twitter-style)
// ============================================================================

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
      rounded-xl bg-[#1a1a1a] border border-[#2a2a2a]
      p-5 transition-all duration-300
      hover:border-[#3a3a3a] hover:translate-y-[-2px]
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
        <Typography className="text-[#666] text-[11px]">
          {role}, {company}
        </Typography>
      </Box>
    </Box>
    
    <Typography className="text-[#b0b0b0] text-xs leading-relaxed mb-3">
      &quot;{quote}&quot;
    </Typography>
    
    <Box className="flex items-center gap-2 text-[11px] text-[#666]">
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

// ============================================================================
// TERMINAL WINDOW
// ============================================================================

interface TerminalWindowProps {
  children: ReactNode;
  className?: string;
  title?: string;
}

export const TerminalWindow: FC<TerminalWindowProps> = ({
  children,
  className = '',
  title = 'Terminal',
}) => (
  <Box
    className={`
      rounded-xl bg-[#1a1a1a] border border-[#2a2a2a]
      overflow-hidden
      ${className}
    `}
  >
    <Box className="flex items-center gap-2 px-4 py-3 bg-[#2a2a2a] border-b border-[#3a3a3a]">
      <Box className="w-3 h-3 rounded-full bg-[#ef4444]" />
      <Box className="w-3 h-3 rounded-full bg-[#f59e0b]" />
      <Box className="w-3 h-3 rounded-full bg-[#10b981]" />
      <Typography className="ml-3 text-xs text-[#666]">{title}</Typography>
    </Box>
    
    <Box className="p-4 font-mono text-sm">
      {children}
    </Box>
  </Box>
);

// ============================================================================
// COMPARISON ICONS
// ============================================================================

export const WarningIcon = () => (
  <svg className="w-3.5 h-3.5 text-[#f59e0b]/70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
  </svg>
);

// ============================================================================
// COMPARISON CELL
// ============================================================================

type ComparisonStatus = 'yes' | 'no' | 'partial' | 'na';

interface ComparisonCellProps {
  status: ComparisonStatus;
  text?: string;
  className?: string;
}

export const ComparisonCell: FC<ComparisonCellProps> = ({
  status,
  text,
  className = '',
}) => {
  const getStatusDisplay = () => {
    switch (status) {
      case 'yes':
        return <CheckIcon />;
      case 'no':
        return <XIcon />;
      case 'partial':
        return <WarningIcon />;
      case 'na':
        return <span className="text-[#666]">N/A</span>;
    }
  };

  return (
    <Box className={`flex items-center gap-2 ${className}`}>
      {getStatusDisplay()}
      {text && <span className="text-sm text-[#b0b0b0]">{text}</span>}
    </Box>
  );
};

// ============================================================================
// STEP INDICATOR
// ============================================================================

interface StepProps {
  number: number;
  title: string;
  description: string;
  icon: ReactNode;
  isLast?: boolean;
}

export const Step: FC<StepProps> = ({
  number,
  title,
  description,
  icon,
  isLast = false,
}) => (
  <Box className="flex-1 relative">
    {!isLast && (
      <Box className="hidden lg:block absolute top-12 left-[calc(50%+40px)] w-[calc(100%-80px)] h-0.5 bg-white/10" />
    )}
    
    <Box className="flex flex-col items-center text-center">
      <Box className="w-14 h-14 rounded-xl bg-white/10 border border-white/10 flex items-center justify-center mb-3 text-white">
        {icon}
      </Box>
      
      <Badge className="mb-2">Step {number}</Badge>
      
      <Typography className="text-sm font-semibold text-white mb-1.5">
        {title}
      </Typography>
      
      <Typography className="text-xs text-[#a0a0a0] max-w-xs">
        {description}
      </Typography>
    </Box>
  </Box>
);

// ============================================================================
// METRIC CARD (Enhanced)
// ============================================================================

interface MetricCardProps {
  value: string;
  label: string;
  sublabel?: string;
  className?: string;
}

export const MetricCard: FC<MetricCardProps> = ({
  value,
  label,
  sublabel,
  className = '',
}) => (
  <Card className={`text-center p-4 ${className}`}>
    <Typography className="text-xl md:text-2xl font-bold text-white">
      {value}
    </Typography>
    <Typography className="text-xs text-[#a0a0a0] mt-1">
      {label}
    </Typography>
    {sublabel && (
      <Typography className="text-xs text-[#666] mt-1">
        {sublabel}
      </Typography>
    )}
  </Card>
);
