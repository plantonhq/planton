'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { AwsAlbForm } from '../forms/AwsAlbForm';
import { Settings, Zap } from 'lucide-react';

interface ComponentConfigurationProps {
  selectedComponent: string | null;
  onDeploy: () => void;
}

export default function ComponentConfiguration({
  selectedComponent: _selectedComponent,
  onDeploy,
}: ComponentConfigurationProps) {
  const handleFormSubmit = (data: unknown) => {
    console.log('Component configuration submitted:', data);
    // Trigger navigation to deploy logs
    onDeploy();
  };

  const handleFormCancel = () => {
    console.log('Component configuration cancelled');
  };

  return (
    <div className="h-full flex flex-col bg-white">
      {/* Educational Header */}
      <div className="bg-gradient-to-r from-white to-[#666] px-8 py-6 text-white">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="max-w-4xl"
        >
          <div className="flex items-center gap-3 mb-3">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center backdrop-blur-sm">
              <Settings className="w-6 h-6 text-white" />
            </div>
            <h2 className="text-3xl font-black">Component Configuration</h2>
          </div>
          <div className="flex items-start gap-2 bg-white/10 rounded-lg p-4 backdrop-blur-sm">
            <Zap className="w-5 h-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-lg font-semibold mb-1">
                Fill out the configuration form
              </p>
              <p className="text-white/90 text-sm">
                Every setting the Terraform provider exposes has a typed home here &mdash; organized 
                so the essentials surface first. Fill in what matters for your deployment; sensible defaults cover the rest.
              </p>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Form Content */}
      <div className="flex-1 overflow-hidden">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3, duration: 0.6 }}
          className="h-full"
        >
          <AwsAlbForm onSubmit={handleFormSubmit} onCancel={handleFormCancel} />
        </motion.div>
      </div>
    </div>
  );
}

