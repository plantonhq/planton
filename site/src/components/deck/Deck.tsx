'use client';

import React, { useState, useCallback, useEffect, useRef, type ComponentType, type ReactNode } from 'react';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { Navigation } from './navigation';
import { PresenterNotes } from './presenter-notes';

/** One slide: its hash id, the name the dots show, the component, and the presenter's notes. */
export interface SlideConfig {
  id: string;
  name: string;
  component: ComponentType<SlideComponentProps>;
  presenterNotes?: string[];
}

export interface SlideComponentProps {
  notesVisible?: boolean;
}

export interface DeckProps {
  slides: SlideConfig[];
  /**
   * Optional chrome a caller lays over the deck: a meeting's guest line, a
   * persona deck's name and address. Rendered above the slide, below the
   * navigation, and never part of any slide's own layout. What qualifies a
   * slide's content (a roadmap disclosure) belongs on that slide, not here.
   */
  frame?: ReactNode;
}

/**
 * The deck engine every presentation on the site runs on: meeting decks,
 * persona decks, and anything else that is a sequence of full-screen slides.
 * It owns the mechanics and nothing else: the slide index lives in the URL
 * hash (#slide-id) so a slide is shareable and back/forward work; arrows,
 * space, PageUp/PageDown, Home/End move; N toggles the presenter notes and F
 * fullscreen; a horizontal swipe moves on touch; slides cross-fade through
 * AnimatePresence. What a slide says and how it looks belongs to the slide
 * components and the deck's slide kit (./primitives), never here.
 */
export function Deck({ slides, frame }: DeckProps) {
  const [currentSlideIndex, setCurrentSlideIndex] = useState(0);
  // A person who asked their system for less motion gets an instant slide
  // change; so does the capture harness, which emulates that preference so a
  // deep-linked slide reaches its resting frame instead of being caught
  // mid cross-fade.
  const reduceMotion = useReducedMotion();
  const [notesVisible, setNotesVisible] = useState(false);
  const touchStartX = useRef<number | null>(null);
  const touchEndX = useRef<number | null>(null);

  const currentSlide = slides[currentSlideIndex];
  const isFirstSlide = currentSlideIndex === 0;
  const isLastSlide = currentSlideIndex === slides.length - 1;

  // Sync slide index from URL hash (initial load + browser back/forward).
  useEffect(() => {
    const syncFromHash = () => {
      const hash = window.location.hash.slice(1);
      if (hash) {
        const index = slides.findIndex((s) => s.id === hash);
        if (index !== -1) setCurrentSlideIndex(index);
      }
    };

    const raf = requestAnimationFrame(syncFromHash);
    window.addEventListener('hashchange', syncFromHash);
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener('hashchange', syncFromHash);
    };
  }, [slides]);

  // Update hash when slide changes
  const updateHash = useCallback((slideId: string) => {
    window.history.replaceState(null, '', `#${slideId}`);
  }, []);

  const goToSlide = useCallback(
    (index: number) => {
      if (index >= 0 && index < slides.length) {
        setCurrentSlideIndex(index);
        updateHash(slides[index].id);
      }
    },
    [slides, updateHash]
  );

  const goToHome = useCallback(() => {
    goToSlide(0);
  }, [goToSlide]);

  const navigateNext = useCallback(() => {
    if (!isLastSlide) {
      goToSlide(currentSlideIndex + 1);
    }
  }, [isLastSlide, currentSlideIndex, goToSlide]);

  const navigatePrev = useCallback(() => {
    if (!isFirstSlide) {
      goToSlide(currentSlideIndex - 1);
    }
  }, [isFirstSlide, currentSlideIndex, goToSlide]);

  const toggleNotes = useCallback(() => {
    setNotesVisible((prev) => !prev);
  }, []);

  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  }, []);

  const handleKeyPress = useCallback(
    (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowRight':
        case ' ':
        case 'PageDown':
          e.preventDefault();
          navigateNext();
          break;
        case 'ArrowLeft':
        case 'PageUp':
          e.preventDefault();
          navigatePrev();
          break;
        case 'Home':
        case 'h':
        case 'H':
          e.preventDefault();
          goToHome();
          break;
        case 'End':
          e.preventDefault();
          goToSlide(slides.length - 1);
          break;
        case 'n':
        case 'N':
          e.preventDefault();
          toggleNotes();
          break;
        case 'f':
        case 'F':
          e.preventDefault();
          toggleFullscreen();
          break;
      }
    },
    [navigateNext, navigatePrev, goToHome, goToSlide, slides.length, toggleNotes, toggleFullscreen]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [handleKeyPress]);

  const handleTouchStart = (e: React.TouchEvent) => {
    touchStartX.current = e.touches[0].clientX;
    touchEndX.current = e.touches[0].clientX;
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    touchEndX.current = e.touches[0].clientX;
  };

  const handleTouchEnd = () => {
    if (touchStartX.current === null || touchEndX.current === null) return;

    const diff = touchStartX.current - touchEndX.current;
    const threshold = 50; // Minimum swipe distance in pixels

    if (Math.abs(diff) > threshold) {
      if (diff > 0) {
        // Swipe left -> next slide
        navigateNext();
      } else {
        // Swipe right -> previous slide
        navigatePrev();
      }
    }

    touchStartX.current = null;
    touchEndX.current = null;
  };

  const CurrentSlideComponent = currentSlide.component;
  const slideInfos = slides.map((s) => ({ id: s.id, name: s.name }));

  return (
    <div
      className="h-[100dvh] bg-canvas flex flex-col relative overflow-hidden touch-pan-x"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      style={{ touchAction: 'pan-x' }}
    >
      <Navigation
        slides={slideInfos}
        currentIndex={currentSlideIndex}
        onNavigate={goToSlide}
        onHome={goToHome}
        onPrev={navigatePrev}
        onNext={navigateNext}
        notesVisible={notesVisible}
        onToggleNotes={toggleNotes}
      />

      {frame}

      <div className="flex-1 flex items-center justify-center">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentSlide.id}
            initial={reduceMotion ? false : { opacity: 0, x: 50 }}
            animate={{ opacity: 1, x: 0 }}
            exit={reduceMotion ? undefined : { opacity: 0, x: -50 }}
            transition={{ duration: reduceMotion ? 0 : 0.3 }}
            className="w-full h-full"
          >
            <CurrentSlideComponent notesVisible={notesVisible} />
          </motion.div>
        </AnimatePresence>
      </div>

      <PresenterNotes
        notes={currentSlide.presenterNotes || []}
        visible={notesVisible}
      />
    </div>
  );
}
