'use client';

import { FC } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

// ============================================================================
// PRESENTER NOTES PANEL
// ============================================================================

interface PresenterNotesProps {
  notes: string[];
  visible: boolean;
  className?: string;
}

/**
 * PresenterNotes - Yellow-highlighted panel for presenter talking points
 *
 * Features:
 * - Toggle visibility with N key (handled by the Deck engine)
 * - Animated show/hide
 * - Bullet points for talking points
 * - Positioned at bottom of slide
 */
export const PresenterNotes: FC<PresenterNotesProps> = ({
  notes,
  visible,
  className = '',
}) => {
  if (!notes || notes.length === 0) return null;

  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 20 }}
          transition={{ duration: 0.2 }}
          className={`
            fixed bottom-20 left-4 right-4 sm:left-8 sm:right-8 md:left-12 md:right-12
            max-w-4xl mx-auto
            bg-white/5 border border-white/20
            rounded-xl sm:rounded-2xl
            p-4 sm:p-5 md:p-6
            backdrop-blur-sm
            z-30
            max-h-[40vh] overflow-y-auto
            ${className}
          `}
        >
          <div className="flex items-center gap-2 mb-3">
            <span className="text-fg-secondary text-sm">📋</span>
            <h4 className="text-xs sm:text-sm font-semibold text-fg-secondary uppercase tracking-wider">
              Presenter Notes
            </h4>
          </div>
          <ul className="space-y-2 sm:space-y-3">
            {notes.map((note, index) => (
              <li
                key={index}
                className="flex items-start gap-2 sm:gap-3 text-xs sm:text-sm text-white/70 leading-relaxed"
              >
                <span className="text-fg-secondary/60 mt-0.5 shrink-0">•</span>
                <span dangerouslySetInnerHTML={{ __html: note }} />
              </li>
            ))}
          </ul>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

// ============================================================================
// PRESENTER NOTES INDICATOR - Shows when notes are available
// ============================================================================

interface NotesIndicatorProps {
  hasNotes: boolean;
  notesVisible: boolean;
}

export const NotesIndicator: FC<NotesIndicatorProps> = ({ hasNotes, notesVisible }) => {
  if (!hasNotes) return null;

  return (
    <div
      className={`
        fixed top-3 sm:top-4 md:top-6 right-14 sm:right-16 md:right-20
        z-50
        px-2 py-1
        rounded-lg
        text-xs
        transition-all
        ${notesVisible
          ? 'bg-white/10 text-fg-secondary border border-white/20'
          : 'bg-white/10 text-white/40'}
      `}
    >
      📋
    </div>
  );
};

// ============================================================================
// INLINE NOTES SLOT - For embedding notes in slide content
// ============================================================================

interface PresenterNotesSlotProps {
  notes?: string[];
  visible: boolean;
  className?: string;
}

/**
 * PresenterNotesSlot - Inline version for embedding within slide content
 *
 * Use this when you want notes to appear as part of the slide
 * rather than in a fixed overlay.
 */
export const PresenterNotesSlot: FC<PresenterNotesSlotProps> = ({
  notes,
  visible,
  className = '',
}) => {
  if (!notes || notes.length === 0 || !visible) return null;

  return (
    <div
      className={`
        mt-6 sm:mt-8
        bg-white/5 border border-white/20
        rounded-xl
        p-4 sm:p-5
        text-left
        ${className}
      `}
    >
      <h4 className="text-xs font-semibold text-fg-secondary uppercase tracking-wider mb-3">
        📋 Presenter Notes
      </h4>
      <ul className="space-y-2">
        {notes.map((note, index) => (
          <li
            key={index}
            className="flex items-start gap-2 text-xs sm:text-sm text-white/70 leading-relaxed"
          >
            <span className="text-fg-secondary/60 mt-0.5 shrink-0">•</span>
            <span dangerouslySetInnerHTML={{ __html: note }} />
          </li>
        ))}
      </ul>
    </div>
  );
};
