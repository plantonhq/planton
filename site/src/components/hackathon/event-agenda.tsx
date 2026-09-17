'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  PageSectionBackgroundContainer,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const EventAgenda = () => {
  return (
    <PageSection>
      <PageSectionBackgroundContainer>
        <Box
          className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] filter blur-[120px]"
          sx={{
            backgroundImage:
              'radial-gradient(circle, rgba(168, 85, 247, 0.2) 0%, rgba(236, 72, 153, 0.2) 50%, transparent 70%)',
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
            Event Agenda
          </Typography>
          <Typography fontSize={18} className="text-gray-400 max-w-3xl mx-auto">
            This hackathon is structured in two phases - first we establish the foundation together, then we build individually.
          </Typography>
        </Stack>

        {/* Two Phase Grid */}
        <Box className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-6xl mx-auto mb-12">
          {/* Phase 1 */}
          <Box className="bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-sm border border-white/20 rounded-2xl p-8 hover:border-white/30 transition-all">
            <Box className="flex items-center justify-between mb-4">
              <Typography fontSize={14} fontWeight={700} className="text-[#a0a0a0] uppercase tracking-wider">
                Phase 1
              </Typography>
              <Typography fontSize={14} className="text-[#a0a0a0]">
                ~4 hours
              </Typography>
            </Box>
            
            <Typography fontSize={28} fontWeight={700} className="text-white mb-3">
              Foundation Setup
            </Typography>
            
            <Typography fontSize={15} className="text-gray-300 mb-6 leading-relaxed">
              Swaroop leads as teacher and technical decision-maker. We learn together and establish the mobile development workflow.
            </Typography>

            <Box className="mb-6">
              <Typography fontSize={16} fontWeight={600} className="text-[#a0a0a0] mb-3">
                Decisions We&apos;ll Make:
              </Typography>
              <Stack spacing={2}>
                {[
                  'Mobile Platform (React Native vs Flutter vs Native)',
                  'Development Environment (IDE, simulators, tools)',
                  'Project Structure (monorepo vs separate repos)',
                  'Git Workflow (branching, commits, PRs)',
                  'Build System (local builds, hot reload, debugging)',
                  'Deployment (getting apps on devices)',
                  'State Management (libraries and patterns)',
                  'Testing (unit, integration, manual)',
                  'Backend Strategy (local storage vs cloud backends)',
                  'Common Libraries (navigation, UI, utilities)',
                ].map((item, index) => (
                  <Box key={index} className="flex items-start gap-2">
                    <Box className="w-1.5 h-1.5 rounded-full bg-white mt-2 flex-shrink-0" />
                    <Typography fontSize={14} className="text-gray-300">
                      {item}
                    </Typography>
                  </Box>
                ))}
              </Stack>
            </Box>

            <Box className="bg-white/10 border border-white/10 rounded-xl p-4">
              <Typography fontSize={13} fontWeight={600} className="text-[#a0a0a0] mb-1">
                Outcome:
              </Typography>
              <Typography fontSize={14} className="text-gray-300 leading-relaxed">
                Everyone has a working dev environment and understands the full SDLC - ready to just write code.
              </Typography>
            </Box>
          </Box>

          {/* Phase 2 */}
          <Box className="bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-sm border border-white/20 rounded-2xl p-8 hover:border-white/30 transition-all">
            <Box className="flex items-center justify-between mb-4">
              <Typography fontSize={14} fontWeight={700} className="text-[#a0a0a0] uppercase tracking-wider">
                Phase 2
              </Typography>
              <Typography fontSize={14} className="text-[#a0a0a0]">
                ~4-6 hours
              </Typography>
            </Box>
            
            <Typography fontSize={28} fontWeight={700} className="text-white mb-3">
              Execution
            </Typography>
            
            <Typography fontSize={15} className="text-gray-300 mb-6 leading-relaxed">
              Everyone builds using Cursor AI (vibe coding). Apply Phase 1 decisions to build actual apps.
            </Typography>

            <Box className="mb-6">
              <Typography fontSize={16} fontWeight={600} className="text-[#a0a0a0] mb-3">
                What We&apos;ll Build:
              </Typography>
              <Stack spacing={3}>
                <Box>
                  <Typography fontSize={15} fontWeight={600} className="text-white mb-1">
                    Vehicle Mileage Tracker
                  </Typography>
                  <Typography fontSize={14} className="text-gray-400">
                    Track odometer readings and visualize trends
                  </Typography>
                </Box>
                <Box>
                  <Typography fontSize={15} fontWeight={600} className="text-white mb-1">
                    Construction Expense Tracker
                  </Typography>
                  <Typography fontSize={14} className="text-gray-400">
                    Track expenses by partner, calculate profit distribution
                  </Typography>
                </Box>
                <Box>
                  <Typography fontSize={15} fontWeight={600} className="text-white mb-1">
                    Your App Ideas
                  </Typography>
                  <Typography fontSize={14} className="text-gray-400">
                    Build whatever you&apos;ve been wanting to create!
                  </Typography>
                </Box>
              </Stack>
            </Box>

            <Box className="bg-white/10 border border-white/10 rounded-xl p-4">
              <Typography fontSize={13} fontWeight={600} className="text-[#a0a0a0] mb-1">
                Outcome:
              </Typography>
              <Typography fontSize={14} className="text-gray-300 leading-relaxed">
                Real working apps built with AI assistance. Focus exclusively on writing code.
              </Typography>
            </Box>
          </Box>
        </Box>

        {/* The 2020 Parallel Story */}
        <Box className="max-w-4xl mx-auto">
          <Box className="bg-gradient-to-r from-white/5 to-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-8">
            <Typography fontSize={20} fontWeight={600} className="text-[#a0a0a0] mb-4">
              The 2020 Parallel
            </Typography>
            <Stack spacing={3}>
              <Typography fontSize={15} className="text-gray-300 leading-relaxed">
                In 2020, Suresh knew zero about building APIs. Swaroop had a 30-minute call with him explaining all the foundational choices: Java, PostgreSQL, gRPC, git workflow, SDLC. By the end, Suresh only had to worry about writing code. That foundation enabled 5 years of productive API development.
              </Typography>
              <Typography fontSize={16} fontWeight={600} className="text-white">
                This hackathon is that 30-minute call - but for mobile development.
              </Typography>
              <Typography fontSize={15} className="text-gray-300 leading-relaxed">
                After today, you&apos;ll only need to worry about writing code to build mobile apps.
              </Typography>
            </Stack>
          </Box>
        </Box>
      </SectionContainer>
    </PageSection>
  );
};

