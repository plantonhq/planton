'use client';

import { Slide, SlideHeader, StatsGrid, Checklist, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S12Traction(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Traction"
        title="Conviction, Not Just Code"
      />

      <StatsGrid
        stats={[
          { icon: '📅', value: '3+', label: 'Years Building' },
          { icon: '💰', value: '$500K+', label: 'Self-Funded' },
          { icon: '📦', value: '370+', label: 'Cloud Resources' },
          { icon: '👥', value: '3', label: 'Active Customers' },
          { icon: '🖥️', value: '4', label: 'App Surfaces' },
        ]}
        className="mb-6 sm:mb-8 max-w-3xl mx-auto"
      />

      <div className="max-w-2xl mx-auto mb-6 sm:mb-8">
        <Checklist
          items={[
            '3 customers actively using the platform — Pro, Plus, and Free tiers',
            '100% retention — zero churn since first paying customer 8 months ago',
            'Both Planton and Stigmer run on Planton — 100% dogfooding',
            '370+ cloud resource kinds across 14 cloud providers',
            'Complete platform: Web + Desktop + Mobile + CLI',
            'Open-source IaC foundation (Planton) — zero lock-in',
          ]}
        />
      </div>

      <Callout variant="success" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base text-white font-medium text-center">
          $500K of our own money. 3+ years of engineering. Zero churn.
          <br />
          <span className="text-white/60">We eat our own cooking — both companies run entirely on Planton.</span>
        </p>
      </Callout>
    </Slide>
  );
}
