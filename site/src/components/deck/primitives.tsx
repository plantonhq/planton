'use client';

import { FC, ReactNode } from 'react';
import { motion } from 'framer-motion';

// ============================================================================
// DESIGN TOKENS - Meets presentation theme
// ============================================================================

export const colors = {
  // Primary gradient (pink/purple)
  gradientStart: '#ec4899',
  gradientEnd: '#8b5cf6',

  // Accent colors
  accentGreen: '#10b981',
  accentRed: '#ef4444',
  accentAmber: '#f59e0b',
  accentBlue: '#3b82f6',
  accentPink: '#ec4899',
  accentPurple: '#8b5cf6',

  // Background layers
  bgPrimary: '#0a0a0a',
  bgSlide: '#12121a',
  bgCard: '#151520',

  // Text
  textPrimary: '#ffffff',
  textSecondary: '#a0a0b8',
  textMuted: '#6b6b80',

  // Borders
  border: '#2a2a3a',
  borderHover: '#3a3a4a',
};

// ============================================================================
// SLIDE WRAPPER - Mobile-first, single screen
// ============================================================================

interface SlideProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'problem' | 'solution' | 'gradient' | 'cover';
}

export const Slide: FC<SlideProps> = ({
  children,
  className = '',
  variant = 'default',
}) => {
  const bgClass = {
    default: 'bg-gradient-to-br from-[#12121a] via-[#12121a] to-[#12121a]',
    problem: 'bg-gradient-to-br from-[#1a0a0a] via-[#2d1515] to-[#1a0a0a]',
    solution: 'bg-gradient-to-br from-[#0a1a0f] via-[#152d1a] to-[#0a1a0f]',
    gradient: 'bg-gradient-to-br from-[#12121a] via-[#1a1a1a] to-[#12121a]',
    cover: 'bg-gradient-to-br from-[#12121a] via-[#1a1a1a] to-[#12121a]',
  }[variant];

  return (
    <div
      className={`
        h-[100dvh] overflow-y-auto
        flex flex-col items-center justify-center
        px-4 py-16 sm:px-6 sm:py-20 md:px-8 md:py-24
        ${bgClass}
        ${className}
      `}
    >
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="w-full max-w-5xl mx-auto"
      >
        {children}
      </motion.div>
    </div>
  );
};

// ============================================================================
// SLIDE HEADER - Section tag + title + subtitle
// ============================================================================

interface SlideHeaderProps {
  sectionTag?: string;
  title: string;
  subtitle?: string;
  className?: string;
}

export const SlideHeader: FC<SlideHeaderProps> = ({
  sectionTag,
  title,
  subtitle,
  className = '',
}) => (
  <div className={`mb-6 sm:mb-8 md:mb-10 text-center ${className}`}>
    {sectionTag && <SectionTag>{sectionTag}</SectionTag>}
    <SlideTitle>{title}</SlideTitle>
    {subtitle && <SlideSubtitle>{subtitle}</SlideSubtitle>}
  </div>
);

// ============================================================================
// TYPOGRAPHY - Responsive scaling
// ============================================================================

interface TypographyProps {
  children: ReactNode;
  className?: string;
}

export const SlideTitle: FC<TypographyProps> = ({ children, className = '' }) => (
  <h1
    className={`
    text-2xl sm:text-3xl md:text-4xl lg:text-[2.5rem]
    font-bold text-white leading-tight
    ${className}
  `}
  >
    {children}
  </h1>
);

export const SlideSubtitle: FC<TypographyProps> = ({ children, className = '' }) => (
  <p
    className={`
    text-sm sm:text-base md:text-lg lg:text-xl
    text-white/60 mt-2 sm:mt-3 md:mt-4
    max-w-3xl mx-auto
    ${className}
  `}
  >
    {children}
  </p>
);

export const SectionTag: FC<TypographyProps> = ({ children, className = '' }) => (
  <span
    className={`
    inline-block
    text-xs font-semibold
    tracking-wider uppercase
    text-white
    bg-white/10
    px-3 py-1.5
    rounded-full
    mb-3 sm:mb-4
    ${className}
  `}
  >
    {children}
  </span>
);

export const CardTitle: FC<TypographyProps> = ({ children, className = '' }) => (
  <h3
    className={`
    text-base sm:text-lg md:text-xl
    font-semibold text-white
    ${className}
  `}
  >
    {children}
  </h3>
);

export const CardText: FC<TypographyProps> = ({ children, className = '' }) => (
  <p
    className={`
    text-xs sm:text-sm
    text-white/60
    leading-relaxed
    ${className}
  `}
  >
    {children}
  </p>
);

// ============================================================================
// CARDS - Content containers
// ============================================================================

interface CardProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'highlight' | 'success' | 'danger' | 'purple' | 'pink';
}

export const Card: FC<CardProps> = ({
  children,
  className = '',
  variant = 'default',
}) => {
  const variantClass = {
    default: 'bg-white/[0.03] border-white/10',
    highlight: 'bg-gradient-to-br from-white/20 to-[#666]/20 border-white/20',
    success: 'bg-[#10b981]/10 border-[#10b981]/30',
    danger: 'bg-red-500/10 border-red-500/30',
    purple: 'bg-white/5 border-white/20',
    pink: 'bg-white/5 border-white/20',
  }[variant];

  return (
    <div
      className={`
      rounded-xl sm:rounded-2xl
      border p-4 sm:p-5 md:p-6
      ${variantClass}
      ${className}
    `}
    >
      {children}
    </div>
  );
};

// ============================================================================
// QUOTE BOX - Case study quotes with attribution
// ============================================================================

interface QuoteBoxProps {
  children: ReactNode;
  attribution?: string;
  className?: string;
}

export const QuoteBox: FC<QuoteBoxProps> = ({
  children,
  attribution,
  className = '',
}) => (
  <div
    className={`
    bg-gradient-to-r from-white/10 to-[#666]/10
    border-l-4 border-white
    px-5 py-4 sm:px-6 sm:py-5 md:px-7 md:py-6
    rounded-r-xl
    ${className}
  `}
  >
    <blockquote className="text-sm sm:text-base md:text-lg text-white leading-relaxed italic">
      {children}
    </blockquote>
    {attribution && (
      <p className="mt-3 text-xs sm:text-sm text-white/50 not-italic">
        {attribution}
      </p>
    )}
  </div>
);

// ============================================================================
// COMPARISON - Before/After layout
// ============================================================================

interface ComparisonItemProps {
  label: string;
  value: string;
  subtext?: string;
  variant: 'before' | 'after';
}

interface ComparisonProps {
  before: Omit<ComparisonItemProps, 'variant'>;
  after: Omit<ComparisonItemProps, 'variant'>;
  className?: string;
}

export const Comparison: FC<ComparisonProps> = ({ before, after, className = '' }) => (
  <div className={`flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-8 md:gap-12 ${className}`}>
    <ComparisonItem {...before} variant="before" />
    <div className="text-2xl sm:text-3xl text-[#10b981] rotate-90 sm:rotate-0">→</div>
    <ComparisonItem {...after} variant="after" />
  </div>
);

const ComparisonItem: FC<ComparisonItemProps> = ({ label, value, subtext, variant }) => {
  const borderClass = variant === 'before'
    ? 'border-red-500/30'
    : 'border-[#10b981]/30';

  return (
    <div
      className={`
      text-center p-5 sm:p-6 md:p-8
      bg-white/[0.03] rounded-2xl
      border ${borderClass}
      min-w-[140px] sm:min-w-[180px]
    `}
    >
      <div className="text-xs uppercase tracking-wider text-white/50 mb-2">{label}</div>
      <div
        className={`
        text-3xl sm:text-4xl md:text-5xl font-bold
        text-white
      `}
      >
        {value}
      </div>
      {subtext && <div className="text-xs sm:text-sm text-white/50 mt-1">{subtext}</div>}
    </div>
  );
};

// ============================================================================
// FLOW DIAGRAM - Step-by-step process
// ============================================================================

interface FlowStep {
  icon?: string;
  label: string;
}

interface FlowDiagramProps {
  steps: FlowStep[];
  className?: string;
}

export const FlowDiagram: FC<FlowDiagramProps> = ({ steps, className = '' }) => (
  <div className={`flex flex-wrap items-center justify-center gap-3 sm:gap-4 py-4 ${className}`}>
    {steps.map((step, index) => (
      <div key={index} className="flex items-center gap-3 sm:gap-4">
        <div className="bg-white/5 border border-white/10 px-3 py-2 sm:px-4 sm:py-3 rounded-lg text-xs sm:text-sm text-white/70">
          {step.icon && <span className="mr-2">{step.icon}</span>}
          {step.label}
        </div>
        {index < steps.length - 1 && (
          <span className="text-white text-lg">→</span>
        )}
      </div>
    ))}
  </div>
);

// ============================================================================
// CHECKLIST - Checkmark list items
// ============================================================================

interface ChecklistProps {
  items: string[];
  className?: string;
}

export const Checklist: FC<ChecklistProps> = ({ items, className = '' }) => (
  <ul className={`space-y-2 sm:space-y-3 ${className}`}>
    {items.map((item, index) => (
      <li key={index} className="flex items-start gap-3 text-left">
        <span className="text-[#10b981] font-bold shrink-0 mt-0.5">✓</span>
        <span className="text-sm sm:text-base text-white/70">{item}</span>
      </li>
    ))}
  </ul>
);

// ============================================================================
// ICON LIST - Icon + title + description list
// ============================================================================

interface IconListItem {
  icon: string;
  title: string;
  description: string;
}

interface IconListProps {
  items: IconListItem[];
  className?: string;
}

export const IconList: FC<IconListProps> = ({ items, className = '' }) => (
  <ul className={`space-y-4 sm:space-y-5 ${className}`}>
    {items.map((item, index) => (
      <li key={index} className="flex items-start gap-3 sm:gap-4 text-left">
        <div className="w-10 h-10 sm:w-12 sm:h-12 bg-white/10 rounded-xl flex items-center justify-center text-lg sm:text-xl shrink-0">
          {item.icon}
        </div>
        <div>
          <h4 className="text-sm sm:text-base font-semibold text-white">{item.title}</h4>
          <p className="text-xs sm:text-sm text-white/60 mt-0.5">{item.description}</p>
        </div>
      </li>
    ))}
  </ul>
);

// ============================================================================
// METRICS - Large numbers
// ============================================================================

interface MetricProps {
  value: string;
  label: string;
  sublabel?: string;
  highlight?: boolean;
  className?: string;
}

export const Metric: FC<MetricProps> = ({
  value,
  label,
  sublabel,
  highlight = false,
  className = '',
}) => (
  <div
    className={`
    text-center p-3 sm:p-4 md:p-6
    ${highlight ? 'bg-gradient-to-br from-white/20 to-[#666]/20 border border-white/20 rounded-xl sm:rounded-2xl' : ''}
    ${className}
  `}
  >
    <div
      className={`
      text-2xl sm:text-3xl md:text-4xl lg:text-5xl
      font-bold mb-1 sm:mb-2
      text-white
    `}
    >
      {value}
    </div>
    <div className="text-xs sm:text-sm md:text-base text-white/60">{label}</div>
    {sublabel && <div className="text-xs text-white/40 mt-0.5">{sublabel}</div>}
  </div>
);

// ============================================================================
// STATS GRID - Multiple metrics in grid
// ============================================================================

interface StatItem {
  icon?: string;
  value: string;
  label: string;
}

interface StatsGridProps {
  stats: StatItem[];
  className?: string;
}

export const StatsGrid: FC<StatsGridProps> = ({ stats, className = '' }) => (
  <div className={`grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 sm:gap-4 ${className}`}>
    {stats.map((stat, index) => (
      <div
        key={index}
        className="text-center p-3 sm:p-4 bg-white/[0.03] rounded-xl border border-white/10 flex flex-col justify-center min-h-[100px] sm:min-h-[120px]"
      >
        {stat.icon && (
          <div className="text-2xl sm:text-3xl mb-2">{stat.icon}</div>
        )}
        <div className="text-lg sm:text-xl md:text-2xl font-bold text-white">
          {stat.value}
        </div>
        <div className="text-xs text-white/50 mt-1">{stat.label}</div>
      </div>
    ))}
  </div>
);

// ============================================================================
// BADGES
// ============================================================================

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'purple' | 'pink' | 'demo';
  className?: string;
}

export const Badge: FC<BadgeProps> = ({
  children,
  variant = 'default',
  className = '',
}) => {
  const variantClass = {
    default: 'bg-white/10 text-white/70 border-white/20',
    success: 'bg-[#10b981]/20 text-[#10b981] border-[#10b981]/30',
    warning: 'bg-white/10 text-[#a0a0a0] border-white/20',
    purple: 'bg-white/10 text-white border-white/20',
    pink: 'bg-white/10 text-white border-white/20',
    demo: 'bg-gradient-to-r from-[#10b981] to-[#666] text-white border-transparent',
  }[variant];

  return (
    <span
      className={`
      inline-flex items-center
      px-3 py-1 sm:px-4 sm:py-1.5
      rounded-full border
      text-xs sm:text-sm font-semibold
      ${variantClass}
      ${className}
    `}
    >
      {children}
    </span>
  );
};

// ============================================================================
// DEMO BADGE - Live demo section marker
// ============================================================================

export const DemoBadge: FC<{ className?: string }> = ({ className = '' }) => (
  <Badge variant="demo" className={className}>
    🎬 LIVE DEMO
  </Badge>
);

// ============================================================================
// GRID LAYOUTS
// ============================================================================

interface GridProps {
  children: ReactNode;
  cols?: 2 | 3 | 4;
  gap?: 'sm' | 'md';
  className?: string;
}

export const Grid: FC<GridProps> = ({
  children,
  cols = 3,
  gap = 'md',
  className = '',
}) => {
  const colsClass = {
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
  }[cols];

  const gapClass = {
    sm: 'gap-3 sm:gap-4',
    md: 'gap-4 sm:gap-5 md:gap-6',
  }[gap];

  return <div className={`grid ${colsClass} ${gapClass} ${className}`}>{children}</div>;
};

// ============================================================================
// TWO COLUMN LAYOUT
// ============================================================================

interface TwoColProps {
  children: ReactNode;
  className?: string;
}

export const TwoCol: FC<TwoColProps> = ({ children, className = '' }) => (
  <div className={`grid grid-cols-1 md:grid-cols-2 gap-5 sm:gap-6 md:gap-8 ${className}`}>
    {children}
  </div>
);

// ============================================================================
// CALLOUT BOX
// ============================================================================

interface CalloutProps {
  children: ReactNode;
  variant?: 'default' | 'success' | 'highlight' | 'warning';
  className?: string;
}

export const Callout: FC<CalloutProps> = ({
  children,
  variant = 'default',
  className = '',
}) => {
  const variantClass = {
    default: 'bg-white/5 border-white/10',
    success: 'bg-[#10b981]/10 border-[#10b981]/30',
    highlight: 'bg-gradient-to-r from-white/10 to-[#666]/10 border-white/10',
    warning: 'bg-white/5 border-white/20',
  }[variant];

  return (
    <div
      className={`
      border rounded-xl sm:rounded-2xl
      p-4 sm:p-5 md:p-6
      ${variantClass}
      ${className}
    `}
    >
      {children}
    </div>
  );
};

// ============================================================================
// ICONS
// ============================================================================

export const CheckIcon: FC<{ className?: string }> = ({ className = '' }) => (
  <svg
    className={`w-4 h-4 sm:w-5 sm:h-5 text-[#10b981] ${className}`}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
  </svg>
);

export const XIcon: FC<{ className?: string }> = ({ className = '' }) => (
  <svg
    className={`w-4 h-4 sm:w-5 sm:h-5 text-red-400 ${className}`}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
  </svg>
);
