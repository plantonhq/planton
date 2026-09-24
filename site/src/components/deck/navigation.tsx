'use client';

import { FC } from 'react';
import { ArrowLeft, ArrowRight, Home } from 'lucide-react';
import { motion } from 'framer-motion';

// ============================================================================
// NAVIGATION TYPES
// ============================================================================

export interface SlideInfo {
  id: string;
  name: string;
}

interface NavigationProps {
  slides: SlideInfo[];
  currentIndex: number;
  onNavigate: (index: number) => void;
  onHome: () => void;
  onPrev: () => void;
  onNext: () => void;
  notesVisible: boolean;
  onToggleNotes: () => void;
}

// ============================================================================
// PROGRESS BAR - Top of screen
// ============================================================================

interface ProgressBarProps {
  current: number;
  total: number;
}

export const ProgressBar: FC<ProgressBarProps> = ({ current, total }) => (
  <div className="absolute top-0 left-0 right-0 h-1 bg-white/10 z-50">
    <motion.div
      className="h-full bg-white"
      initial={{ width: 0 }}
      animate={{ width: `${((current + 1) / total) * 100}%` }}
      transition={{ duration: 0.3 }}
    />
  </div>
);

// ============================================================================
// SLIDE DOTS - Navigation dots
// ============================================================================

interface SlideDotsProps {
  slides: SlideInfo[];
  currentIndex: number;
  onNavigate: (index: number) => void;
}

export const SlideDots: FC<SlideDotsProps> = ({ slides, currentIndex, onNavigate }) => (
  <div className="absolute top-3 sm:top-4 md:top-6 left-1/2 -translate-x-1/2 z-50 flex gap-1 sm:gap-1.5">
    {slides.map((slide, index) => (
      <button
        key={slide.id}
        onClick={() => onNavigate(index)}
        className={`h-1.5 sm:h-2 rounded-full transition-all ${
          index === currentIndex
            ? 'bg-white w-4 sm:w-6'
            : 'bg-white/30 hover:bg-white/50 w-1.5 sm:w-2'
        }`}
        title={slide.name}
        aria-label={`Go to ${slide.name} slide`}
      />
    ))}
  </div>
);

// ============================================================================
// HOME BUTTON - Top right
// ============================================================================

interface HomeButtonProps {
  onClick: () => void;
}

export const HomeButton: FC<HomeButtonProps> = ({ onClick }) => (
  <div className="absolute top-3 sm:top-4 md:top-6 right-3 sm:right-4 md:right-6 z-50">
    <button
      onClick={onClick}
      className="p-2 bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 backdrop-blur-sm rounded-lg transition-all"
      aria-label="Go to first slide"
    >
      <Home className="w-4 h-4" />
    </button>
  </div>
);

// ============================================================================
// NAVIGATION FOOTER - Bottom bar with controls
// ============================================================================

interface NavigationFooterProps {
  currentIndex: number;
  totalSlides: number;
  currentSlideName: string;
  isFirstSlide: boolean;
  isLastSlide: boolean;
  onPrev: () => void;
  onNext: () => void;
  notesVisible: boolean;
  onToggleNotes: () => void;
}

export const NavigationFooter: FC<NavigationFooterProps> = ({
  currentIndex,
  totalSlides,
  currentSlideName,
  isFirstSlide,
  isLastSlide,
  onPrev,
  onNext,
  notesVisible,
  onToggleNotes,
}) => (
  <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent pt-8 sm:pt-10 md:pt-12 pb-3 sm:pb-4 md:pb-6 z-40">
    <div className="max-w-7xl mx-auto px-3 sm:px-4 md:px-8">
      <div className="flex items-center justify-between">
        {/* Slide counter */}
        <div className="flex items-center gap-2 sm:gap-4">
          <span className="text-xs sm:text-sm font-medium text-white/60">
            {currentIndex + 1}/{totalSlides}
          </span>
          <span className="hidden sm:inline text-xs sm:text-sm text-white/40">
            {currentSlideName}
          </span>
        </div>

        {/* Navigation buttons */}
        <div className="flex gap-2 sm:gap-3">
          {/* Notes toggle button - subtle, doesn't compete with nav */}
          <button
            onClick={onToggleNotes}
            className={`hidden sm:flex items-center gap-1.5 px-2.5 py-1.5 sm:py-2 text-xs sm:text-sm border rounded-lg transition-all ${
              notesVisible
                ? 'bg-white/15 text-white/80 border-white/30'
                : 'bg-transparent text-white/40 border-white/10 hover:text-white/60 hover:border-white/20'
            }`}
          >
            <kbd className="px-1 py-0.5 bg-white/10 rounded text-[10px] text-white/50">N</kbd>
          </button>

          {/* Previous button */}
          <button
            onClick={onPrev}
            disabled={isFirstSlide}
            className="flex items-center gap-1 sm:gap-2 px-2 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-transparent text-white text-xs sm:text-sm border border-white/20 hover:bg-white/5 hover:border-white/40 disabled:opacity-30 disabled:cursor-not-allowed rounded-lg transition-all"
          >
            <ArrowLeft className="w-3 h-3 sm:w-4 sm:h-4" />
            <span className="hidden sm:inline">Back</span>
          </button>

          {/* Next button */}
          <button
            onClick={onNext}
            disabled={isLastSlide}
            className="flex items-center gap-1 sm:gap-2 px-2 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-white text-black text-xs sm:text-sm font-medium disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/90 rounded-lg transition-all"
          >
            <span className="hidden sm:inline">Next</span>
            <ArrowRight className="w-3 h-3 sm:w-4 sm:h-4" />
          </button>
        </div>
      </div>

      {/* Keyboard shortcuts hint */}
      <div className="hidden md:block text-center mt-3">
        <p className="text-xs text-white/30">
          <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">←</kbd>{' '}
          <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">→</kbd> to navigate •{' '}
          <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">N</kbd> notes •{' '}
          <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">F</kbd> fullscreen •{' '}
          <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">Home</kbd> start
        </p>
      </div>
    </div>
  </div>
);

// ============================================================================
// MOBILE SWIPE HINT - Shows on first slide
// ============================================================================

interface SwipeHintProps {
  visible: boolean;
}

export const SwipeHint: FC<SwipeHintProps> = ({ visible }) => {
  if (!visible) return null;

  return (
    <div className="sm:hidden absolute bottom-20 left-0 right-0 text-center z-50">
      <div className="inline-flex items-center gap-2 px-4 py-2 bg-white/10 border border-white/20 rounded-full animate-pulse backdrop-blur-sm">
        <span className="text-xs font-medium text-white/60">Swipe to Navigate</span>
        <ArrowRight className="w-3.5 h-3.5 text-white/60" />
      </div>
    </div>
  );
};

// ============================================================================
// FULL NAVIGATION COMPONENT - Combines all navigation elements
// ============================================================================

export const Navigation: FC<NavigationProps> = ({
  slides,
  currentIndex,
  onNavigate,
  onHome,
  onPrev,
  onNext,
  notesVisible,
  onToggleNotes,
}) => {
  const isFirstSlide = currentIndex === 0;
  const isLastSlide = currentIndex === slides.length - 1;
  const currentSlide = slides[currentIndex];

  return (
    <>
      <ProgressBar current={currentIndex} total={slides.length} />
      <SlideDots slides={slides} currentIndex={currentIndex} onNavigate={onNavigate} />
      <HomeButton onClick={onHome} />
      <NavigationFooter
        currentIndex={currentIndex}
        totalSlides={slides.length}
        currentSlideName={currentSlide?.name || ''}
        isFirstSlide={isFirstSlide}
        isLastSlide={isLastSlide}
        onPrev={onPrev}
        onNext={onNext}
        notesVisible={notesVisible}
        onToggleNotes={onToggleNotes}
      />
      <SwipeHint visible={isFirstSlide} />
    </>
  );
};
