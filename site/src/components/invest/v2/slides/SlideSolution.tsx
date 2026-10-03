'use client';

import React from 'react';
import { Globe, Zap, CheckCircle } from 'lucide-react';
import { Slide, SlideTitle, SlideSubtitle, Card } from '../shared';

export default function SlideSolution() {
  return (
    <Slide>
      <SlideTitle>The Solution</SlideTitle>
      <SlideSubtitle className="mb-6 sm:mb-8">
        One Platform to Deploy Infrastructure and Services Across Any Cloud
      </SlideSubtitle>

      {/* Two product cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6 mb-6 sm:mb-8">
        {/* InfraHub */}
        <Card className="text-left !border-[#10b981]/20">
          <div className="flex items-center gap-2 mb-2">
            <Globe className="w-5 h-5 sm:w-6 sm:h-6 text-[#10b981]" />
            <h3 className="text-lg sm:text-xl font-semibold text-white">Infra Hub</h3>
          </div>
          <p className="text-xs text-[#10b981]/70 mb-2">
            Replaces Terraform Enterprise / Pulumi Cloud
          </p>
          <p className="text-xs sm:text-sm text-[#a0a0a0] mb-2">
            Deploy Any Infra Component with a Single API
          </p>
          <ul className="space-y-1">
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-[#10b981]/70 shrink-0" />
              <span>AWS, GCP, Azure, Kubernetes</span>
            </li>
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-[#10b981]/70 shrink-0" />
              <span>Pre-Built Infrastructure Templates</span>
            </li>
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-[#10b981]/70 shrink-0" />
              <span>Point-and-Click Deployment</span>
            </li>
          </ul>
        </Card>

        {/* ServiceHub */}
        <Card className="text-left">
          <div className="flex items-center gap-2 mb-2">
            <Zap className="w-5 h-5 sm:w-6 sm:h-6 text-white" />
            <h3 className="text-lg sm:text-xl font-semibold text-white">Service Hub</h3>
          </div>
          <p className="text-xs text-[#a0a0a0] mb-2">
            Replaces GitHub Actions / Jenkins / GitLab Pipelines
          </p>
          <p className="text-xs sm:text-sm text-[#a0a0a0] mb-2">
            Service Hub: Vercel for Backend, In Your Own Cloud
          </p>
          <ul className="space-y-1">
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-white/70 shrink-0" />
              <span>Git Push to Production</span>
            </li>
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-white/70 shrink-0" />
              <span>No Dockerfile Required</span>
            </li>
            <li className="flex items-center gap-2 text-xs sm:text-sm text-[#a0a0a0]">
              <CheckCircle className="w-3.5 h-3.5 text-white/70 shrink-0" />
              <span>Built-In CI/CD with Tekton</span>
            </li>
          </ul>
        </Card>
      </div>

      {/* The Promise */}
      <div className="bg-[#151515] border border-[#10b981]/20 rounded-xl p-4 sm:p-5 md:p-6 max-w-2xl mx-auto">
        <p className="text-base sm:text-lg md:text-xl text-white font-medium">
          &ldquo;Vercel for Backend, In Your Own Cloud&rdquo; &mdash; Service Hub
        </p>
        <p className="text-xs sm:text-sm text-[#a0a0a0] mt-1">
          Compose With AI. Deploy In Your Own Cloud.
        </p>
        <p className="text-xs text-[#10b981]/70 mt-2">
          One platform instead of integration chaos
        </p>
      </div>
    </Slide>
  );
}

