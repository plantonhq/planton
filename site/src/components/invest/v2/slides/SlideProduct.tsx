'use client';

import React from 'react';
import Image from 'next/image';
import { Layers } from 'lucide-react';
import { Slide, SlideTitle, SlideSubtitle } from '../shared';

// GitHub icon SVG (same as landing page) - kept for potential future use
const _GitHubIcon = ({ className }: { className?: string }) => (
  <svg className={className} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

// Multi-cloud icon component
const MultiCloudIcon = () => (
  <div className="flex items-center justify-center gap-0.5 sm:gap-1">
    <Image src="/_site/images/providers/aws.svg" alt="AWS" width={24} height={24} className="w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7" />
    <Image src="/_site/images/providers/gcp.svg" alt="GCP" width={24} height={24} className="w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7" />
    <Image src="/_site/images/providers/azure.svg" alt="Azure" width={24} height={24} className="w-5 h-5 sm:w-6 sm:h-6 md:w-7 md:h-7" />
  </div>
);

// Terraform + Pulumi icon component
const IaCIcon = () => (
  <div className="flex items-center justify-center gap-1 sm:gap-2">
    <Image src="/_site/images/providers/terraform.svg" alt="Terraform" width={48} height={48} className="w-6 h-6 sm:w-8 sm:h-8 md:w-9 md:h-9" />
    <Image src="/_site/images/providers/pulumi.svg" alt="Pulumi" width={48} height={48} className="w-7 h-7 sm:w-9 sm:h-9 md:w-10 md:h-10" />
  </div>
);

// GitHub + GitLab icon component
const GitProvidersIcon = () => (
  <div className="flex items-center justify-center gap-2 sm:gap-3">
    <Image src="/_site/images/providers/github-dark.svg" alt="GitHub" width={48} height={48} className="w-7 h-7 sm:w-9 sm:h-9 md:w-11 md:h-11" />
    <Image src="/_site/images/providers/gitlab.svg" alt="GitLab" width={48} height={48} className="w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12" />
  </div>
);

// Infrastructure steps
const infraSteps = [
  {
    iconType: 'multiCloud' as const,
    title: 'Connect Cloud',
    description: 'Link AWS, GCP, or Azure with Secure OAuth',
  },
  {
    iconType: 'lucide' as const,
    icon: Layers,
    title: 'Choose Infra',
    description: 'Pick from 120+ Catalog Kinds',
  },
  {
    iconType: 'iac' as const,
    title: 'Deploy',
    description: 'Live Terraform / Pulumi Visualization',
  },
];

// Service steps
const serviceSteps = [
  {
    iconType: 'gitProviders' as const,
    title: 'Connect Git',
    description: 'Link GitHub or GitLab',
  },
  {
    iconType: 'image' as const,
    iconSrc: '/_site/images/tekton.svg',
    title: 'Deploy Services',
    description: 'Git Push to Production, No CI/CD Setup',
  },
];

export default function SlideProduct() {
  return (
    <Slide className="!justify-start !pt-24 sm:!pt-28 md:!pt-32">
      <SlideTitle className="mb-1 sm:mb-2">How It Works</SlideTitle>
      <SlideSubtitle className="mb-4 sm:mb-6">
        From Zero to Production in Minutes
      </SlideSubtitle>

      {/* InfraHub - 3-Step Process */}
      <div className="mb-4 sm:mb-6 lg:mb-8 mt-4 sm:mt-6">
        <span className="inline-block text-xs text-[#a0a0a0] font-medium mb-3 sm:mb-4 px-3 py-1 rounded-full bg-[#2a2a2a] border border-[#3a3a3a]">Infrastructure</span>
        <div className="flex flex-row items-center justify-center gap-4 sm:gap-8 md:gap-12 lg:gap-16">
          {infraSteps.map((step, index) => (
            <React.Fragment key={step.title}>
              <div className="flex flex-col items-center text-center flex-1 max-w-[140px] sm:max-w-[180px] lg:max-w-[200px]">
                <div className="w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center mb-2 sm:mb-3">
                  {step.iconType === 'multiCloud' ? (
                    <MultiCloudIcon />
                  ) : step.iconType === 'iac' ? (
                    <IaCIcon />
                  ) : (
                    <step.icon className="w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 text-[#10b981]" />
                  )}
                </div>
                <h3 className="text-xs sm:text-sm font-semibold text-white mb-0.5">{step.title}</h3>
                <p className="text-xs text-[#666] hidden sm:block">{step.description}</p>
              </div>
              {index < infraSteps.length - 1 && (
                <div className="text-[#3a3a3a] text-lg">—</div>
              )}
            </React.Fragment>
          ))}
        </div>
      </div>

      {/* ServiceHub - 2-Step Process */}
      <div className="mb-4 sm:mb-6 lg:mb-8 mt-2 sm:mt-4">
        <span className="inline-block text-xs text-[#a0a0a0] font-medium mb-3 sm:mb-4 px-3 py-1 rounded-full bg-[#2a2a2a] border border-[#3a3a3a]">Services</span>
        <div className="flex flex-row items-center justify-center gap-6 sm:gap-10 md:gap-16 lg:gap-20">
          {serviceSteps.map((step, index) => (
            <React.Fragment key={step.title}>
              <div className="flex flex-col items-center text-center flex-1 max-w-[160px] sm:max-w-[200px] lg:max-w-[220px]">
                <div className="w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center mb-2 sm:mb-3">
                  {step.iconType === 'gitProviders' ? (
                    <GitProvidersIcon />
                  ) : (
                    <Image 
                      src={step.iconSrc!} 
                      alt={step.title} 
                      width={48} 
                      height={48} 
                      className="w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12"
                    />
                  )}
                </div>
                <h3 className="text-xs sm:text-sm font-semibold text-white mb-0.5">{step.title}</h3>
                <p className="text-xs text-[#666] hidden sm:block">{step.description}</p>
              </div>
              {index < serviceSteps.length - 1 && (
                <div className="text-[#3a3a3a] text-lg">—</div>
              )}
            </React.Fragment>
          ))}
        </div>
      </div>

      {/* Result callout */}
      <div className="bg-[#151515] border border-[#2a2a2a] rounded-xl p-4 sm:p-5 max-w-lg mx-auto mt-4 sm:mt-6">
        <p className="text-sm sm:text-base text-white font-medium">
          Complete Cloud Infra + Service Deployments in <span className="text-[#10b981]">&lt;1 Hour</span>
        </p>
        <p className="text-xs text-[#666] mt-1">vs. Weeks with Manual Setup</p>
      </div>
    </Slide>
  );
}

