'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  PageSectionBackgroundContainer,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const GoalsSection = () => {
  return (
    <PageSection>
      <PageSectionBackgroundContainer>
        <Box
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] filter blur-[100px]"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(168, 85, 247, 0.3) 0%, transparent 70%)',
          }}
        />
      </PageSectionBackgroundContainer>

      <SectionContainer className="z-[1]">
        <Stack spacing={3} className="max-w-4xl mx-auto text-center mb-12">
          <Typography
            fontSize={48}
            fontWeight={700}
            className="text-white"
          >
            Why This Hackathon?
          </Typography>
          <Typography fontSize={18} className="text-gray-400 max-w-3xl mx-auto mb-4">
            We&apos;ve never built mobile apps at Planton. We have ideas - productivity tools, personal utilities, hobby projects - but they stay ideas because we don&apos;t have the workflow established.
          </Typography>
          <Typography fontSize={18} className="text-gray-400 max-w-3xl mx-auto">
            This hackathon will establish that foundation, just like a foundational 30-minute call in 2020 enabled 5 years of API development at Planton.
          </Typography>
        </Stack>

        <Box className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
          <GoalCard
            icon="🚀"
            title="Establish Foundation"
            description="Make all the technical decisions upfront - platform, tools, workflow, deployment. Document everything so it's repeatable."
          />
          <GoalCard
            icon="⚡"
            title="Normalize Building"
            description="Make mobile app building as natural as vibe coding APIs. After this hackathon, you'll only worry about writing code."
          />
          <GoalCard
            icon="💡"
            title="Enable Quick Ideas"
            description="Turn those small app ideas into reality. No more 'I wish I could build this' - just build it using Cursor AI."
          />
          <GoalCard
            icon="🤝"
            title="Learn Together"
            description="Share knowledge, solve problems as a team. Swaroop leads the technical decisions, we all learn and grow."
          />
        </Box>

        {/* The Pitch */}
        <Box className="mt-16 max-w-4xl mx-auto">
          <Box className="bg-gradient-to-r from-white/10 to-[#666]/10 backdrop-blur-sm border border-white/10 rounded-2xl p-8">
            <Typography fontSize={20} fontWeight={600} className="text-[#a0a0a0] mb-4">
              The Vision
            </Typography>
            <Typography fontSize={16} className="text-gray-300 leading-relaxed">
              Just like we&apos;ve made infrastructure deployment simple with vibe coding, we want to make mobile app development equally accessible. There are so many small ideas that could benefit the team and our community - a mileage tracker, an expense splitter, productivity tools, hobby projects. This hackathon is about removing the friction and making mobile development a normal part of our toolkit.
            </Typography>
          </Box>
        </Box>
      </SectionContainer>
    </PageSection>
  );
};

const GoalCard = ({ 
  icon, 
  title, 
  description 
}: { 
  icon: string; 
  title: string; 
  description: string;
}) => (
  <Box className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all hover:border-white/20">
    <Typography fontSize={40} className="mb-3">
      {icon}
    </Typography>
    <Typography fontSize={22} fontWeight={700} className="text-white mb-3">
      {title}
    </Typography>
    <Typography fontSize={15} className="text-gray-400 leading-relaxed">
      {description}
    </Typography>
  </Box>
);










