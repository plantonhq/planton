'use client';

import { Slide, SlideHeader, IconList, Callout, TwoCol, Card, CardTitle } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S08AITeammates(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="AI Teammates"
        title="Your DevOps Team, Instantly"
        subtitle="Not a chatbot. Not a copilot. A teammate."
      />

      <TwoCol className="mb-6 sm:mb-8 max-w-4xl mx-auto">
        <Card className="text-left">
          <CardTitle className="mb-4 !text-base sm:!text-lg">How It Works</CardTitle>
          <IconList
            items={[
              {
                icon: '👤',
                title: 'Specialized Personas',
                description: 'Each teammate has a name, avatar, and deep expertise — AWS, GCP, Kubernetes, security.',
              },
              {
                icon: '⚡',
                title: 'Instant Availability',
                description: 'Add to your team with zero cost and zero onboarding. Available immediately.',
              },
              {
                icon: '🧠',
                title: 'Full Context',
                description: 'Teammates know your org\'s connections, environments, variables, and secrets.',
              },
            ]}
          />
        </Card>

        <Card className="text-left">
          <CardTitle className="mb-4 !text-base sm:!text-lg">Why It&apos;s Different</CardTitle>
          <IconList
            items={[
              {
                icon: '🔒',
                title: 'Deterministic Execution',
                description: 'Every action is backed by the cloud catalog. No hallucinated Terraform.',
              },
              {
                icon: '📱',
                title: 'Always Available',
                description: 'Desktop, mobile, web — interact with your teammates wherever you are, like Slack.',
              },
              {
                icon: '🎯',
                title: 'Skills-Based',
                description: 'Each teammate comes with inspectable skills built on real platform documentation.',
              },
            ]}
          />
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base text-white/70 text-center">
          We solved the talent hiring problem by <strong className="text-white">replacing the need to hire</strong>.
          <br />
          Teams add AI teammates that are instantly available, at a fraction of the cost.
        </p>
      </Callout>
    </Slide>
  );
}
