'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const AppIdeas = () => {
  return (
    <div id="app-ideas">
      <PageSection>
        <SectionContainer>
        <Stack spacing={3} className="max-w-4xl mx-auto text-center mb-12">
          <Typography
            fontSize={48}
            fontWeight={700}
            className="text-white"
          >
            Initial App Ideas
          </Typography>
          <Typography fontSize={18} className="text-gray-400">
            Two real use cases to get started - but bring your own ideas too!
          </Typography>
        </Stack>

        <Box className="space-y-8 max-w-5xl mx-auto">
          {/* App Idea 1 */}
          <AppIdeaCard
            number="01"
            title="Vehicle Mileage Tracker"
            stakeholder="Avinash"
            description="A simple app to track vehicle mileage over time. Currently maintained manually in logs, this app will make it easy to record odometer readings and visualize mileage trends."
            features={[
              'Quick odometer entry with date/time',
              'Automatic mileage calculation',
              'Historical trend visualization',
              'Simple, clean interface',
              'Works on both Android and iOS'
            ]}
            icon="🚗"
            color="from-white to-[#a0a0a0]"
          />

          {/* App Idea 2 */}
          <AppIdeaCard
            number="02"
            title="Construction Partner Expense Tracker"
            stakeholder="Sai"
            description="For real estate construction projects in Hyderabad. When Sai and his partners build apartments, they currently track construction expenses on paper. This app will digitize expense tracking, making it easy to see who invested/spent what during construction and calculate profit shares at project completion."
            features={[
              'Track construction expenses by partner',
              'Record who spent what with receipts/notes',
              'See real-time investment totals per partner',
              'Calculate profit distribution at project end',
              'Replace paper-based expense reconciliation',
              'Simple interface for quick expense entry'
            ]}
            icon="🏗️"
            color="from-white to-[#666]"
          />

          {/* Your Idea Card */}
          <Box className="bg-gradient-to-br from-white/10 to-[#666]/10 backdrop-blur-sm border border-white/10 rounded-2xl p-8">
            <Box className="text-center">
              <Typography fontSize={48} className="mb-4">
                💡
              </Typography>
              <Typography fontSize={28} fontWeight={700} className="text-white mb-3">
                Your Idea Here!
              </Typography>
              <Typography fontSize={16} className="text-gray-300 max-w-2xl mx-auto">
                Have a mobile app idea you&apos;ve been wanting to build? This is your chance! Whether it&apos;s a productivity tool, a hobby project, or something completely new - bring it to the hackathon.
              </Typography>
            </Box>
          </Box>
        </Box>
      </SectionContainer>
    </PageSection>
    </div>
  );
};

const AppIdeaCard = ({ 
  number,
  title, 
  stakeholder,
  description,
  features,
  icon,
  color
}: { 
  number: string;
  title: string; 
  stakeholder: string;
  description: string;
  features: string[];
  icon: string;
  color: string;
}) => (
  <Box className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-8 hover:bg-white/10 transition-all hover:border-white/20">
    <Box className="flex items-start gap-6">
      {/* Icon */}
      <Box className={`w-20 h-20 rounded-2xl bg-gradient-to-br ${color} flex items-center justify-center flex-shrink-0 shadow-lg`}>
        <Typography fontSize={36}>{icon}</Typography>
      </Box>
      
      {/* Content */}
      <Box className="flex-1">
        <Box className="flex items-center gap-3 mb-3">
          <Typography fontSize={14} fontWeight={700} className="text-[#a0a0a0] tracking-wider">
            APP {number}
          </Typography>
          <Box className="w-1.5 h-1.5 rounded-full bg-white" />
          <Typography fontSize={14} className="text-gray-400">
            Requested by <span className="text-[#a0a0a0] font-medium">{stakeholder}</span>
          </Typography>
        </Box>
        
        <Typography fontSize={26} fontWeight={700} className="text-white mb-3">
          {title}
        </Typography>
        
        <Typography fontSize={15} className="text-gray-300 mb-6 leading-relaxed">
          {description}
        </Typography>
        
        <Box>
          <Typography fontSize={13} fontWeight={600} className="text-[#a0a0a0] mb-3 uppercase tracking-wider">
            Key Features
          </Typography>
          <Stack spacing={2}>
            {features.map((feature, index) => (
              <Box key={index} className="flex items-start gap-2">
                <Box className="w-1.5 h-1.5 rounded-full bg-white mt-2 flex-shrink-0" />
                <Typography fontSize={14} className="text-gray-400">
                  {feature}
                </Typography>
              </Box>
            ))}
          </Stack>
        </Box>
      </Box>
    </Box>
  </Box>
);









