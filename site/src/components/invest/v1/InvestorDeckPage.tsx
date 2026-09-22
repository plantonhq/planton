'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { ArrowLeft, ArrowRight, Home } from 'lucide-react';
import { Button } from './ui/button';
import { motion, AnimatePresence } from 'framer-motion';

// Import slide components
import SlideCover from './slides/SlideCover';
import SlideProblem from './slides/SlideProblem';
import SlideSolution from './slides/SlideSolution';
import SlideProduct from './slides/SlideProduct';
import SlideTraction from './slides/SlideTraction';
import SlideMarket from './slides/SlideMarket';
import SlideRoadmap from './slides/SlideRoadmap';
import SlideTeam from './slides/SlideTeam';
import SlideAsk from './slides/SlideAsk';
import SlideClose from './slides/SlideClose';

type SlideType = 
  | 'cover'
  | 'problem'
  | 'solution'
  | 'product'
  | 'traction'
  | 'market'
  | 'roadmap'
  | 'team'
  | 'ask'
  | 'close';

const slides: SlideType[] = [
  'cover',
  'problem',
  'solution',
  'product',
  'traction',
  'market',
  'roadmap',
  'team',
  'ask',
  'close',
];

const slideNames: Record<SlideType, string> = {
  cover: 'Cover',
  problem: 'Problem',
  solution: 'Solution',
  product: 'Product',
  traction: 'Traction',
  market: 'Market',
  roadmap: 'Roadmap',
  team: 'Team',
  ask: 'The Ask',
  close: 'Close',
};

export default function InvestorDeckPage() {
  const [currentSlideIndex, setCurrentSlideIndex] = useState(0);

  const currentSlide = slides[currentSlideIndex];
  const isFirstSlide = currentSlideIndex === 0;
  const isLastSlide = currentSlideIndex === slides.length - 1;

  const goToHome = useCallback(() => {
    setCurrentSlideIndex(0);
  }, []);

  const navigateNext = useCallback(() => {
    if (!isLastSlide) {
      setCurrentSlideIndex((prev) => prev + 1);
    }
  }, [isLastSlide]);

  const navigatePrev = useCallback(() => {
    if (!isFirstSlide) {
      setCurrentSlideIndex((prev) => prev - 1);
    }
  }, [isFirstSlide]);

  const handleKeyPress = useCallback(
    (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowRight':
        case ' ':
          e.preventDefault();
          navigateNext();
          break;
        case 'ArrowLeft':
          e.preventDefault();
          navigatePrev();
          break;
        case 'Escape':
          e.preventDefault();
          goToHome();
          break;
      }
    },
    [navigateNext, navigatePrev, goToHome]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [handleKeyPress]);

  const renderCurrentSlide = () => {
    switch (currentSlide) {
      case 'cover':
        return <SlideCover />;
      case 'problem':
        return <SlideProblem />;
      case 'solution':
        return <SlideSolution />;
      case 'product':
        return <SlideProduct />;
      case 'traction':
        return <SlideTraction />;
      case 'market':
        return <SlideMarket />;
      case 'roadmap':
        return <SlideRoadmap />;
      case 'team':
        return <SlideTeam />;
      case 'ask':
        return <SlideAsk />;
      case 'close':
        return <SlideClose />;
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-[#0a0a0a] flex flex-col relative overflow-hidden">
      {/* Home button */}
      <div className="absolute top-6 right-6 z-50">
        <Button
          variant="ghost"
          size="sm"
          onClick={goToHome}
          className="bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 backdrop-blur-sm"
        >
          <Home className="w-4 h-4" />
        </Button>
      </div>

      {/* Slide navigation dots */}
      <div className="absolute top-6 left-1/2 -translate-x-1/2 z-50 flex gap-2">
        {slides.map((slide, index) => (
          <button
            key={slide}
            onClick={() => setCurrentSlideIndex(index)}
            className={`w-2 h-2 rounded-full transition-all ${
              index === currentSlideIndex
                ? 'bg-white w-6'
                : 'bg-white/30 hover:bg-white/50'
            }`}
            title={slideNames[slide]}
          />
        ))}
      </div>

      {/* Main content */}
      <div className="flex-1 flex items-center justify-center">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentSlide}
            initial={{ opacity: 0, x: 50 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -50 }}
            transition={{ duration: 0.3 }}
            className="w-full h-full"
          >
            {renderCurrentSlide()}
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Navigation footer */}
      <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/50 to-transparent pt-12 pb-6">
        <div className="max-w-7xl mx-auto px-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <span className="text-sm font-medium text-white/60">
                {currentSlideIndex + 1} / {slides.length}
              </span>
              <span className="text-sm text-white/40">
                {slideNames[currentSlide]}
              </span>
            </div>
            <div className="flex gap-3">
              <Button
                variant="outline"
                onClick={navigatePrev}
                disabled={isFirstSlide}
                className="bg-white/10 backdrop-blur-sm text-white border-white/20 hover:bg-white/20 hover:border-white/40 disabled:opacity-30"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back
              </Button>
              <Button
                onClick={navigateNext}
                disabled={isLastSlide}
                className="bg-gradient-to-r from-white to-[#666] text-white font-semibold disabled:opacity-30 hover:from-[#111] hover:to-[#222]"
              >
                Next
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </div>
          </div>

          {/* Keyboard shortcuts hint */}
          <div className="text-center mt-4">
            <p className="text-xs text-white/30">
              <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">←</kbd>{' '}
              <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">→</kbd>{' '}
              or{' '}
              <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/50">Space</kbd>{' '}
              to navigate
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
