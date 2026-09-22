'use client';

import { Slide, SlideHeader, Comparison } from '@/components/deck';

export default function S03SEPChallenge() {
  return (
    <Slide>
      <SlideHeader
        title="You've Solved This Once. Now let's make it repeatable."
      />

      <p className="text-center text-white/50 text-xs sm:text-sm mb-6">
        Source:{' '}
        <a
          href="https://sep.com/our-work/case-study/implementing-enterprise-devops-solutions/"
          target="_blank"
          rel="noopener noreferrer"
          className="text-white/70 hover:text-white underline"
        >
          SEP Enterprise DevOps Solutions Case Study
        </a>
      </p>

      <Comparison
        before={{ label: 'Before', value: '30 Days', subtext: 'Manual setup' }}
        after={{
          label: 'After Standardization',
          value: '30 Min',
          subtext: 'Automated deployment',
        }}
        className="mb-8"
      />

      <p className="text-center text-white/70 text-sm sm:text-base max-w-2xl mx-auto">
        That was custom-built for one client.
        <br />
        <span className="text-white font-medium">
          What if every SEP project could start from that same 30-minute
          baseline?
        </span>
      </p>
    </Slide>
  );
}
