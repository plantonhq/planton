'use client';

import { motion, useReducedMotion } from 'framer-motion';
import type { FC, ReactNode } from 'react';

/**
 * The frame a slide on the site's palette fills: the canvas, the content
 * column, and the same entrance every slide makes. It paints no gradient and
 * types no color; what a slide says is composed inside it from the marketing
 * primitive library, so a deck and the page it was cut from share one type
 * scale, one card, one quote, one set of doors. The meeting decks' older
 * slide kit (./primitives) keeps its own look until those decks converge.
 */
export const SlideFrame: FC<{ children: ReactNode; className?: string }> = ({ children, className = '' }) => {
  const reduceMotion = useReducedMotion();
  return (
    <div className={`h-full overflow-y-auto bg-canvas flex flex-col items-center justify-center px-4 py-16 sm:px-6 sm:py-20 md:px-8 ${className}`}>
      <motion.div initial={reduceMotion ? false : { opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: reduceMotion ? 0 : 0.4 }} className="w-full max-w-5xl mx-auto">
        {children}
      </motion.div>
    </div>
  );
};
