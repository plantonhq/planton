'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { Layers, GitBranch, Cloud, Terminal } from 'lucide-react';

const features = [
  {
    icon: Layers,
    title: 'Catalog Kinds',
    description: 'Pre-built, validated infra component definitions',
  },
  {
    icon: Cloud,
    title: 'Multi-Cloud',
    description: 'Same experience across AWS, GCP, Azure, Kubernetes',
  },
  {
    icon: GitBranch,
    title: 'CI/CD Pipelines',
    description: 'Git-based workflows with zero configuration',
  },
  {
    icon: Terminal,
    title: 'Open Source Core',
    description: 'Planton open source - CLI and IaC modules on GitHub',
  },
];

export default function SlideProduct() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-[#0a0a0a] via-[#1a1a1a] to-[#0a0a0a] p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center max-w-5xl w-full"
      >
        {/* Title */}
        <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
          How It Works
        </h2>
        <p className="text-xl text-white/60 mb-12">
          Three steps from code to production
        </p>

        {/* Steps */}
        <div className="flex flex-col md:flex-row items-center justify-center gap-4 mb-12">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.2 }}
            className="bg-white/5 border border-white/10 rounded-2xl p-6 flex-1 max-w-xs"
          >
            <div className="text-4xl font-bold text-white mb-2">1</div>
            <h3 className="text-lg font-bold text-white mb-2">Connect</h3>
            <p className="text-white/60 text-sm">
              Link your cloud accounts and Git repositories
            </p>
          </motion.div>

          <div className="hidden md:block text-white/30 text-3xl">→</div>

          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.3 }}
            className="bg-white/5 border border-white/10 rounded-2xl p-6 flex-1 max-w-xs"
          >
            <div className="text-4xl font-bold text-white mb-2">2</div>
            <h3 className="text-lg font-bold text-white mb-2">Configure</h3>
            <p className="text-white/60 text-sm">
              Pick components from the catalog or use templates
            </p>
          </motion.div>

          <div className="hidden md:block text-white/30 text-3xl">→</div>

          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.4 }}
            className="bg-white/5 border border-white/10 rounded-2xl p-6 flex-1 max-w-xs"
          >
            <div className="text-4xl font-bold text-emerald-500 mb-2">3</div>
            <h3 className="text-lg font-bold text-white mb-2">Deploy</h3>
            <p className="text-white/60 text-sm">
              Click deploy or push code—done in minutes
            </p>
          </motion.div>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5 + index * 0.1 }}
              className="bg-white/5 border border-white/10 rounded-xl p-4 text-center"
            >
              <feature.icon className="w-8 h-8 text-white/70 mx-auto mb-2" />
              <h4 className="text-sm font-medium text-white">{feature.title}</h4>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </div>
  );
}
