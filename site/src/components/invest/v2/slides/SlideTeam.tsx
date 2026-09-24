'use client';

import React from 'react';
import Image from 'next/image';
import { Slide, SlideTitle, SlideSubtitle, TeamMember, Callout } from '../shared';

// College badge for founders who went to same college (2007-2011)
const CollegeBadge = () => (
  <div 
    className="bg-[#2a2a2a] border border-[#3a3a3a] rounded-full px-1.5 py-0.5 flex items-center gap-1"
    title="College Batchmates (2007-2011)"
  >
    <span className="text-sm">🎓</span>
  </div>
);

// Row 1: Founders (2 members)
const foundersRow = [
  {
    name: 'Swarup Donepudi',
    role: 'Founder',
    avatar: '/_site/images/team/swarup-donepudi.jpg',
    description: (
      <>
        <div><span className="text-[#10b981]">Silicon Valley</span> experience.</div>
        <div>10+ yrs DevOps & Cloud.</div>
        <div>Platform Architect.</div>
      </>
    ),
    isCollegeMate: true,
  },
  {
    name: 'Suresh Attaluri',
    role: 'Co-Founder',
    avatar: '/_site/images/team/suresh-attaluri.jpg',
    description: (
      <>
        <div>Leading <span className="text-[#10b981]">AI</span> R&D.</div>
        <div>Backend & Data Systems.</div>
      </>
    ),
    isCollegeMate: true,
  },
];

// Row 2: Team (3 members)
const teamRow = [
  {
    name: 'Irshad Ahmed',
    role: 'Lead UX Designer',
    avatar: '/_site/images/team/irshad-ahmed.jpg',
    description: (
      <>
        <div>5 years with team.</div>
        <div>All product design.</div>
      </>
    ),
    isCollegeMate: false,
  },
  {
    name: 'Avinash Sana',
    role: 'Operations & BD',
    avatar: '/_site/images/team/avinash-sana.jpg',
    description: (
      <>
        <div>All non-technical operations.</div>
      </>
    ),
    isCollegeMate: true,
  },
  {
    name: 'Satish Lakhani',
    role: 'Full-Stack Engineer',
    avatar: '/_site/images/team/satish-lakhani.jpg',
    description: (
      <>
        <div>Frontend Expert.</div>
        <div>Built Planton WebConsole.</div>
      </>
    ),
    isCollegeMate: false,
  },
];

const strengths = [
  'Deep Domain Expertise',
  '3+ Years Building Together',
  '$500K+ Skin in the Game',
  'Production Platform Shipped',
  'Founding Team: College Batchmates 🎓',
];

export default function SlideTeam() {
  // Combine all team members for mobile grid
  const allMembers = [...foundersRow, ...teamRow];
  
  return (
    <Slide className="!justify-start !pt-24 sm:!pt-28 md:!pt-32">
      <SlideTitle>Team</SlideTitle>
      <SlideSubtitle className="mb-2 sm:mb-6 text-xs sm:text-sm">
        Small, Focused, Committed for 3+ Years
      </SlideSubtitle>

      {/* Mobile: 2-column compact grid for all 5 members */}
      <div className="sm:hidden grid grid-cols-2 gap-1.5 mb-2 mx-auto w-full">
        {allMembers.map((member) => (
          <div key={member.name} className="bg-[#151515] border border-[#2a2a2a] rounded-lg p-1.5 text-left relative">
            {member.isCollegeMate && (
              <div className="absolute -top-1 -right-1">
                <CollegeBadge />
              </div>
            )}
            <div className="flex items-center gap-1.5">
              <Image
                src={member.avatar}
                alt={member.name}
                width={24}
                height={24}
                className="w-6 h-6 rounded-full object-cover object-[center_25%] shrink-0"
              />
              <div className="min-w-0">
                <h3 className="text-[10px] font-semibold text-white truncate">{member.name}</h3>
                <p className="text-[8px] text-[#666] truncate">{member.role}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Desktop: 2 rows - 2 founders + 3 team members */}
      <div className="hidden sm:flex flex-col gap-4 md:gap-5 lg:gap-6 mb-4 md:mb-6 mx-auto">
        {/* Row 1: Founders (2 members, centered) */}
        <div className="grid grid-cols-2 gap-4 md:gap-5 lg:gap-6 max-w-xl md:max-w-3xl lg:max-w-4xl mx-auto w-full">
          {foundersRow.map((member) => (
            <TeamMember
              key={member.name}
              name={member.name}
              role={member.role}
              description={member.description}
              avatar={member.avatar}
              badge={member.isCollegeMate ? <CollegeBadge /> : undefined}
            />
          ))}
        </div>
        {/* Row 2: Team (3 members) */}
        <div className="grid grid-cols-3 gap-4 md:gap-5 lg:gap-6 max-w-2xl md:max-w-4xl lg:max-w-5xl mx-auto w-full">
          {teamRow.map((member) => (
            <TeamMember
              key={member.name}
              name={member.name}
              role={member.role}
              description={member.description}
              avatar={member.avatar}
              badge={member.isCollegeMate ? <CollegeBadge /> : undefined}
            />
          ))}
        </div>
      </div>

      {/* Team Strengths - compact on mobile */}
      <Callout className="max-w-2xl p-2 sm:p-4">
        <h3 className="text-[10px] sm:text-base font-semibold text-white mb-1 sm:mb-3 text-center">Why This Team Wins</h3>
        <div className="grid grid-cols-2 sm:flex sm:flex-col gap-x-2 gap-y-0.5 sm:gap-1.5">
          {strengths.map((strength, index) => (
            <div key={index} className="flex items-center gap-1 sm:gap-1.5 text-[8px] sm:text-sm text-[#a0a0a0]">
              <span className="text-[#10b981]/70 shrink-0">✓</span>
              <span className="leading-tight">{strength}</span>
            </div>
          ))}
        </div>
      </Callout>
    </Slide>
  );
}

