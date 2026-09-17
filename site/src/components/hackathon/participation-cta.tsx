'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  PageSectionBackgroundContainer,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const ParticipationCTA = () => {
  return (
    <PageSection>
      <PageSectionBackgroundContainer>
        <Box
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] filter blur-[120px]"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(236, 72, 153, 0.4) 0%, rgba(168, 85, 247, 0.2) 50%, transparent 70%)',
          }}
        />
      </PageSectionBackgroundContainer>

      <SectionContainer className="z-[1]">
        <Box className="max-w-4xl mx-auto text-center">
          <Typography
            fontSize={48}
            fontWeight={700}
            className="text-white mb-6"
          >
            How to Participate
          </Typography>
          
          <Typography fontSize={18} className="text-gray-300 mb-12 max-w-3xl mx-auto leading-relaxed">
            It&apos;s super simple - no complex registration, no prerequisites, just you and your enthusiasm for building mobile apps!
          </Typography>

          <Box className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-12">
            <StepCard
              number="1"
              title="React on Slack"
              description="Drop a reaction or comment on the Slack announcement to give us a headcount (optional but helpful!)"
              icon="💬"
            />
            <StepCard
              number="2"
              title="Show Up"
              description="Join us Monday, Dec 22nd at 9:00 AM - in person at the office or remotely from home"
              icon="🎯"
            />
            <StepCard
              number="3"
              title="Learn in Phase 1"
              description="This is crucial! Swaroop will lead all technical decisions - platform choice, dev setup, workflow, deployment. Take notes, ask questions."
              icon="📚"
            />
            <StepCard
              number="4"
              title="Build in Phase 2"
              description="Apply what you learned and build real apps using Cursor AI. Focus exclusively on writing code."
              icon="🚀"
            />
          </Box>

          <Box className="bg-gradient-to-r from-white/10 to-[#666]/10 backdrop-blur-sm border border-white/10 rounded-2xl p-8 mb-8">
            <Typography fontSize={20} fontWeight={600} className="text-[#a0a0a0] mb-4">
              What to Bring
            </Typography>
            <Stack direction="row" spacing={4} className="justify-center flex-wrap gap-y-4">
              <Box className="flex items-center gap-2">
                <Typography fontSize={24}>💻</Typography>
                <Typography fontSize={15} className="text-gray-300">Your laptop</Typography>
              </Box>
              <Box className="flex items-center gap-2">
                <Typography fontSize={24}>☕</Typography>
                <Typography fontSize={15} className="text-gray-300">Your favorite drink</Typography>
              </Box>
              <Box className="flex items-center gap-2">
                <Typography fontSize={24}>💡</Typography>
                <Typography fontSize={15} className="text-gray-300">Your app ideas</Typography>
              </Box>
              <Box className="flex items-center gap-2">
                <Typography fontSize={24}>🎉</Typography>
                <Typography fontSize={15} className="text-gray-300">Your enthusiasm</Typography>
              </Box>
            </Stack>
          </Box>

          <Box className="bg-gradient-to-r from-white/5 to-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 mb-8">
            <Typography fontSize={16} fontWeight={600} className="text-[#a0a0a0] mb-2">
              Important:
            </Typography>
            <Typography fontSize={15} className="text-gray-300 leading-relaxed">
              <strong>Phase 1 is essential for everyone.</strong> Even if you can&apos;t stay for the full day, join Phase 1 to learn the foundation. That&apos;s where Swaroop (teacher/decision-maker) will establish the workflow that enables mobile development at Planton.
            </Typography>
          </Box>

          <Typography fontSize={14} className="text-gray-500 mb-8">
            Questions? Reach out to Suresh on Slack or just show up on the day!
          </Typography>
        </Box>
      </SectionContainer>
    </PageSection>
  );
};

const StepCard = ({ 
  number, 
  title, 
  description,
  icon
}: { 
  number: string; 
  title: string; 
  description: string;
  icon: string;
}) => (
  <Box className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all">
    <Box className="flex items-center justify-center gap-3 mb-4">
      <Box className="w-10 h-10 rounded-full bg-gradient-to-br from-white to-[#666] flex items-center justify-center font-bold text-white">
        {number}
      </Box>
      <Typography fontSize={32}>{icon}</Typography>
    </Box>
    <Typography fontSize={20} fontWeight={700} className="text-white mb-3">
      {title}
    </Typography>
    <Typography fontSize={14} className="text-gray-400 leading-relaxed">
      {description}
    </Typography>
  </Box>
);










