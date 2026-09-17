'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const EventDetails = () => {
  return (
    <PageSection>
      <SectionContainer>
        <Stack spacing={3} className="max-w-4xl mx-auto text-center mb-12">
          <Typography
            fontSize={48}
            fontWeight={700}
            className="text-white"
          >
            Event Details
          </Typography>
          <Typography fontSize={18} className="text-gray-400">
            Everything you need to know about the hackathon
          </Typography>
        </Stack>

        <Box className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
          {/* Date & Time Card */}
          <DetailCard
            icon="📅"
            title="When"
            items={[
              'Monday, December 22, 2025',
              '9:00 AM Start Time',
              'Full Day (8-10 hours)',
              'Flexible breaks throughout'
            ]}
          />

          {/* Format Card */}
          <DetailCard
            icon="🌐"
            title="Format"
            items={[
              'Hybrid Event',
              'Join in-person at office',
              'Or participate remotely',
              'Real-time collaboration'
            ]}
          />

          {/* RSVP Card */}
          <DetailCard
            icon="✋"
            title="RSVP"
            items={[
              'Super informal - just show up!',
              'React on Slack for headcount',
              'No formal registration needed',
              'Bring your enthusiasm'
            ]}
          />

          {/* Prerequisites Card */}
          <DetailCard
            icon="🛠️"
            title="Prerequisites"
            items={[
              'None! We\'ll set up together',
              'Bring your laptop',
              'We\'ll configure dev environments',
              'Beginners welcome'
            ]}
          />
        </Box>
      </SectionContainer>
    </PageSection>
  );
};

const DetailCard = ({ 
  icon, 
  title, 
  items 
}: { 
  icon: string; 
  title: string; 
  items: string[];
}) => (
  <Box className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all">
    <Box className="flex items-center gap-3 mb-4">
      <Typography fontSize={32}>{icon}</Typography>
      <Typography fontSize={24} fontWeight={700} className="text-white">
        {title}
      </Typography>
    </Box>
    <Stack spacing={2}>
      {items.map((item, index) => (
        <Box key={index} className="flex items-start gap-2">
          <Box className="w-1.5 h-1.5 rounded-full bg-white mt-2 flex-shrink-0" />
          <Typography fontSize={15} className="text-gray-300">
            {item}
          </Typography>
        </Box>
      ))}
    </Stack>
  </Box>
);









