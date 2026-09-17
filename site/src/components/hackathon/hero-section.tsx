'use client';

import { Box, Stack, Button, Typography } from '@mui/material';
import {
  PageSection,
  PageSectionBackgroundContainer,
  SectionContainer,
  Pill,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const HeroSection = () => {
  return (
    <PageSection>
      <PageSectionBackgroundContainer>
        {/* Primary glow */}
        <Box
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] filter blur-[120px]"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(168, 85, 247, 0.4) 0%, rgba(168, 85, 247, 0.1) 40%, transparent 70%)',
          }}
        />
        {/* Secondary accent glows */}
        <Box
          className="absolute top-1/4 left-1/4 w-[400px] h-[400px] filter blur-[100px] opacity-50"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(236, 72, 153, 0.5) 0%, transparent 70%)',
          }}
        />
        <Box
          className="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] filter blur-[100px] opacity-50"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(59, 130, 246, 0.5) 0%, transparent 70%)',
          }}
        />
      </PageSectionBackgroundContainer>
      
      <SectionContainer className="z-[1]">
        <Pill>Hackathon 2025</Pill>
        <Stack direction="column" spacing={4} className="mt-8">
          <h1 className="text-7xl md:text-8xl font-bold text-white leading-tight">
            MobileVibe
          </h1>
          <Typography fontSize={24} className="text-gray-300 max-w-4xl mx-auto leading-relaxed">
            Join us for Planton&apos;s first-ever mobile app hackathon. Build real apps, establish workflows, and normalize mobile development with vibe coding.
          </Typography>
          
          {/* Event Badge */}
          <Box className="flex flex-col items-center gap-2 mt-8">
            <Box className="flex items-center gap-6 bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl px-8 py-4">
              <Box className="text-center">
                <Typography fontSize={12} className="text-gray-400 uppercase tracking-wider">
                  Date
                </Typography>
                <Typography fontSize={20} fontWeight={700} className="text-white">
                  Dec 22, 2025
                </Typography>
              </Box>
              <Box className="w-px h-12 bg-white/20" />
              <Box className="text-center">
                <Typography fontSize={12} className="text-gray-400 uppercase tracking-wider">
                  Time
                </Typography>
                <Typography fontSize={20} fontWeight={700} className="text-white">
                  9:00 AM
                </Typography>
              </Box>
              <Box className="w-px h-12 bg-white/20" />
              <Box className="text-center">
                <Typography fontSize={12} className="text-gray-400 uppercase tracking-wider">
                  Format
                </Typography>
                <Typography fontSize={20} fontWeight={700} className="text-white">
                  Hybrid
                </Typography>
              </Box>
            </Box>
            <Typography fontSize={14} className="text-[#a0a0a0]">
              Monday • Full Day (8-10 hours) • In-person + Remote
            </Typography>
          </Box>

          <div className="flex items-center justify-center gap-4 mt-8">
            <Button
              className="px-8 py-4 bg-gradient-to-r from-[#111] to-[#222] text-white text-lg font-semibold rounded-xl hover:from-white hover:to-[#666] transition-all shadow-lg shadow-black/25"
              onClick={() => window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' })}
            >
              Join the Hackathon
            </Button>
            <Button
              className="bg-white/5 text-white px-8 py-4 text-lg font-medium rounded-xl border border-white/20 hover:bg-white/10 hover:border-white/30 transition-all backdrop-blur-sm"
              onClick={() => {
                const section = document.getElementById('app-ideas');
                section?.scrollIntoView({ behavior: 'smooth' });
              }}
            >
              View App Ideas
            </Button>
          </div>
        </Stack>

        {/* Hero Visual - Mobile Devices */}
        <div className="mx-auto mt-16 relative h-[400px] w-full max-w-5xl">
          {/* Mobile Phone Illustration */}
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
            <div className="relative">
              {/* Phone 1 - Center */}
              <PhoneDevice 
                position="z-20"
                rotation="rotate-0"
                color="from-white to-[#666]"
                delay="0s"
              />
              {/* Phone 2 - Left */}
              <PhoneDevice 
                position="z-10 -left-32 top-8"
                rotation="-rotate-12"
                color="from-white to-[#a0a0a0]"
                delay="0.5s"
              />
              {/* Phone 3 - Right */}
              <PhoneDevice 
                position="z-10 -right-32 top-8"
                rotation="rotate-12"
                color="from-white to-[#666]"
                delay="1s"
              />
            </div>
          </div>
        </div>

        {/* Stats Bar */}
        <div className="mt-24 grid grid-cols-3 gap-8 max-w-3xl mx-auto">
          <div className="text-center">
            <Typography fontSize={48} fontWeight={700} className="text-white">
              2
            </Typography>
            <Typography fontSize={14} className="text-gray-500 mt-1">
              Initial App Ideas
            </Typography>
          </div>
          <div className="text-center">
            <Typography fontSize={48} fontWeight={700} className="text-white">
              Full
            </Typography>
            <Typography fontSize={14} className="text-gray-500 mt-1">
              Day of Building
            </Typography>
          </div>
          <div className="text-center">
            <Typography fontSize={48} fontWeight={700} className="text-white">
              ∞
            </Typography>
            <Typography fontSize={14} className="text-gray-500 mt-1">
              Your Ideas Welcome
            </Typography>
          </div>
        </div>
      </SectionContainer>
    </PageSection>
  );
};

// Phone Device Component
const PhoneDevice = ({ 
  position, 
  rotation, 
  color,
  delay 
}: { 
  position: string; 
  rotation: string;
  color: string; 
  delay: string;
}) => (
  <div 
    className={`absolute ${position} ${rotation} animate-pulse`}
    style={{ animationDelay: delay, animationDuration: '3s' }}
  >
    <div className={`w-32 h-60 rounded-3xl bg-gradient-to-br ${color} flex flex-col p-3 shadow-2xl`}>
      {/* Phone notch */}
      <div className="w-16 h-1 bg-black/30 rounded-full mx-auto mb-2" />
      {/* Phone screen */}
      <div className="flex-1 bg-white/10 rounded-2xl backdrop-blur-sm flex items-center justify-center">
        <Typography fontSize={32} className="text-white/80">
          📱
        </Typography>
      </div>
      {/* Phone bottom bar */}
      <div className="w-12 h-1 bg-white/20 rounded-full mx-auto mt-2" />
    </div>
  </div>
);









