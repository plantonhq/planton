'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { CheckCircle, Zap, Globe } from 'lucide-react';

export default function SlideSolution() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-[#0a1a0f] via-[#152d1a] to-[#0a1a0f] p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center max-w-5xl w-full"
      >
        {/* Title */}
        <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
          The Solution
        </h2>
        <p className="text-xl text-white/60 mb-12 max-w-3xl mx-auto">
          One platform to deploy infrastructure and services across any cloud.
        </p>

        {/* Main Value Prop */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-12">
          {/* InfraHub */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
            className="bg-white/5 backdrop-blur-sm border border-emerald-500/30 rounded-2xl p-8 text-left"
          >
            <div className="flex items-center gap-3 mb-2">
              <Globe className="w-8 h-8 text-emerald-400" />
              <h3 className="text-2xl font-bold text-white">Infra Hub</h3>
            </div>
            <p className="text-xs text-emerald-400/80 mb-4">
              Replaces Terraform Enterprise / Pulumi Cloud
            </p>
            <p className="text-white/70 mb-4">
              Deploy any infra component with a single API
            </p>
            <ul className="space-y-2">
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-emerald-400" />
                AWS, GCP, Azure, Kubernetes
              </li>
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-emerald-400" />
                Pre-built infrastructure templates
              </li>
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-emerald-400" />
                Point-and-click deployment
              </li>
            </ul>
          </motion.div>

          {/* ServiceHub */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.3 }}
            className="bg-white/5 backdrop-blur-sm border border-white/20 rounded-2xl p-8 text-left"
          >
            <div className="flex items-center gap-3 mb-2">
              <Zap className="w-8 h-8 text-white" />
              <h3 className="text-2xl font-bold text-white">Service Hub</h3>
            </div>
            <p className="text-xs text-white/80 mb-4">
              Replaces GitHub Actions / Jenkins / GitLab Pipelines
            </p>
            <p className="text-white/70 mb-4">
              Service Hub: Vercel for Backend, In Your Own Cloud
            </p>
            <ul className="space-y-2">
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-white" />
                Git push to production
              </li>
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-white" />
                No Dockerfile required
              </li>
              <li className="flex items-center gap-2 text-white/60">
                <CheckCircle className="w-4 h-4 text-white" />
                Built-in CI/CD with Tekton
              </li>
            </ul>
          </motion.div>
        </div>

        {/* The Promise */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="bg-gradient-to-r from-emerald-500/10 to-[#666]/10 border border-white/10 rounded-xl p-6 max-w-2xl mx-auto"
        >
          <p className="text-2xl text-white font-medium">
            &ldquo;Vercel for Backend, In Your Own Cloud&rdquo; &mdash; Service Hub
          </p>
          <p className="text-white/60 mt-2">
            Compose With AI. Deploy In Your Own Cloud.
          </p>
          <p className="text-sm text-white/40 mt-3">
            One platform instead of integration chaos
          </p>
        </motion.div>
      </motion.div>
    </div>
  );
}
