'use client';

import Image from 'next/image';
import { Slide, SlideHeader, Callout, Checklist } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

const CollegeBadge = () => (
  <div
    className="absolute -top-1 -right-1 bg-white/10 border border-white/20 rounded-full px-1.5 py-0.5"
    title="College Batchmates (2007-2011)"
  >
    <span className="text-xs">🎓</span>
  </div>
);

interface TeamMemberCardProps {
  name: string;
  role: string;
  avatar: string;
  lines: string[];
  collegeMate?: boolean;
  highlight?: boolean;
}

function TeamMemberCard({ name, role, avatar, lines, collegeMate, highlight }: TeamMemberCardProps) {
  return (
    <div className={`relative rounded-xl border p-3 sm:p-4 text-left ${
      highlight
        ? 'bg-white/[0.06] border-white/20'
        : 'bg-white/[0.03] border-white/10'
    }`}>
      {collegeMate && <CollegeBadge />}
      <div className="flex items-start gap-3">
        <Image
          src={avatar}
          alt={name}
          width={64}
          height={64}
          className="w-10 h-10 sm:w-14 sm:h-14 rounded-full object-cover object-[center_25%] shrink-0"
        />
        <div className="min-w-0">
          <h3 className="text-sm sm:text-base font-semibold text-white truncate">{name}</h3>
          <p className="text-xs sm:text-sm text-white/50">{role}</p>
          <div className="mt-1 space-y-0.5">
            {lines.map((line) => (
              <p key={line} className="text-xs text-white/40">{line}</p>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function S13Team(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Team"
        title="Small, Focused, Committed for 3+ Years"
      />

      {/* Founders row */}
      <div className="grid grid-cols-2 gap-3 sm:gap-4 max-w-3xl mx-auto mb-3 sm:mb-4">
        <TeamMemberCard
          name="Swarup Donepudi"
          role="Founder"
          avatar="/_site/images/team/swarup-donepudi.jpg"
          lines={['Silicon Valley experience', '10+ yrs DevOps & Cloud', 'Platform Architect']}
          collegeMate
          highlight
        />
        <TeamMemberCard
          name="Suresh Attaluri"
          role="Co-Founder"
          avatar="/_site/images/team/suresh-attaluri.jpg"
          lines={['Leading AI R&D', 'Backend & Data Systems']}
          collegeMate
          highlight
        />
      </div>

      {/* Team row */}
      <div className="grid grid-cols-3 gap-3 sm:gap-4 max-w-4xl mx-auto mb-5 sm:mb-6">
        <TeamMemberCard
          name="Irshad Ahmed"
          role="Lead UX Designer"
          avatar="/_site/images/team/irshad-ahmed.jpg"
          lines={['5 years with team', 'All product design']}
        />
        <TeamMemberCard
          name="Avinash Sana"
          role="Operations & BD"
          avatar="/_site/images/team/avinash-sana.jpg"
          lines={['All non-technical operations']}
          collegeMate
        />
        <TeamMemberCard
          name="Satish Lakhani"
          role="Full-Stack Engineer"
          avatar="/_site/images/team/satish-lakhani.jpg"
          lines={['Frontend Expert', 'Built Planton Console']}
        />
      </div>

      <Callout className="max-w-2xl mx-auto">
        <h3 className="text-xs sm:text-sm font-semibold text-white mb-2 sm:mb-3 text-center">Why This Team Wins</h3>
        <Checklist
          items={[
            'Deep domain expertise — 10+ years in DevOps & cloud',
            '3+ years building together — $500K+ skin in the game',
            'Founding team: college batchmates (2007–2011)',
            'Production platform shipped and running',
          ]}
        />
      </Callout>
    </Slide>
  );
}
