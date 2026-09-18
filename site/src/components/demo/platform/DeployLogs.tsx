'use client';

import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Loader2, Sparkles } from 'lucide-react';
import confetti from 'canvas-confetti';
import StackJobLogger from './StackJobLogger';
import StackJobHeader from './StackJobHeader';
import type { StackJob, StackJobStatus } from '../interfaces/stackjob';
import stackJobData from '../../../../public/demo/stack-job.json';

// Extract the full StackJob data and status from the imported file
// The fixture is a JSON record of a real run; JSON widens its string enums, so it is
// read as the interfaces the components take.
const stackJob = stackJobData as unknown as StackJob;
const stackJobStatus = stackJobData.status as unknown as StackJobStatus;

export default function DeployLogs() {
  const [isDeploying, setIsDeploying] = useState(true);
  const [deploymentComplete, setDeploymentComplete] = useState(false);

  const handleAnimationComplete = () => {
    setIsDeploying(false);
    setDeploymentComplete(true);

    // Trigger confetti
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
    });
  };

  useEffect(() => {
    return () => {
      confetti.reset();
    };
  }, []);

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-8 py-6 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Deployment Console</h2>
          </div>
          <div className="flex items-center gap-3">
            <AnimatePresence>
              {isDeploying && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="flex items-center gap-2 text-white"
                >
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span className="text-sm font-medium">Deploying...</span>
                </motion.div>
              )}
              {deploymentComplete && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="flex items-center gap-2 text-[#10b981]"
                >
                  <Sparkles className="w-4 h-4" />
                  <span className="text-sm font-medium">Deployment Complete!</span>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      {/* StackJob Header */}
      <StackJobHeader stackJob={stackJob} stackJobStatus={stackJobStatus} />

      {/* StackJobLogger */}
      <div className="flex-1 overflow-auto">
        <StackJobLogger
          stackJobStatus={stackJobStatus}
          stackJob={stackJob}
          loading={false}
          onAnimationComplete={handleAnimationComplete}
        />
      </div>
    </div>
  );
}
