'use client';

import { Box, Typography, Stack } from '@mui/material';
import {
  PageSection,
  SectionContainer,
} from '@/components/landing-page/v1-2025-02-15-0800/components';

export const FAQ = () => {
  return (
    <PageSection>
      <SectionContainer>
        <Stack spacing={3} className="max-w-4xl mx-auto text-center mb-12">
          <Typography
            fontSize={48}
            fontWeight={700}
            className="text-white"
          >
            Frequently Asked Questions
          </Typography>
          <Typography fontSize={18} className="text-gray-400">
            Got questions? We&apos;ve got answers.
          </Typography>
        </Stack>

        <Box className="max-w-4xl mx-auto space-y-4">
          <FAQItem
            question="What happens in Phase 1 vs Phase 2?"
            answer="Phase 1 (First Half - ~4 hours): Swaroop leads as teacher and technical decision-maker. We make all the foundational choices together: platform (React Native vs Flutter vs Native), dev environment setup, project structure, git workflow, build system, deployment, state management, testing, backend strategy, and common libraries. Everyone leaves with a working dev environment and clear understanding of the full SDLC. Phase 2 (Second Half - ~4-6 hours): Everyone builds using Cursor AI (vibe coding). Apply the Phase 1 decisions to build actual apps - Vehicle Mileage Tracker, Construction Expense Tracker, or your own ideas. Focus exclusively on writing code with AI assistance. Think of Phase 1 as the foundation that enables Phase 2. After this hackathon, you'll only need to worry about writing code to build mobile apps."
          />
          <FAQItem
            question="Do I need to know mobile development?"
            answer="Not at all! That's exactly why we have Phase 1. Swaroop will teach and guide all the technical decisions - from choosing the platform to setting up your dev environment to understanding the full software development lifecycle. This hackathon draws from a 2020 experience: when Suresh knew zero about building APIs, Swaroop had a 30-minute call explaining all foundational choices (Java, PostgreSQL, gRPC, git workflow). By the end, Suresh only had to worry about writing code. That foundation enabled 5 years of productive API development. We're doing the same for mobile apps. Phase 1 establishes the foundation. Phase 2 is where you write code with Cursor AI. Beginners are absolutely welcome - this is a learning experience."
          />
          <FAQItem
            question="What if I don't have an app idea?"
            answer="That's totally fine! You can join one of the two initial projects (Vehicle Mileage Tracker or Apartment Expense Tracker), or team up with someone else who has an idea. The goal is to learn and build together."
          />
          <FAQItem
            question="Do I need to bring anything?"
            answer="Just your laptop and enthusiasm! We'll handle development environment setup together during the hackathon. No need to prepare or install anything in advance."
          />
          <FAQItem
            question="Can I join remotely?"
            answer="Absolutely! This is a hybrid event. Whether you're joining from the office or from home, you'll be able to participate fully. We'll make sure remote participants are well-integrated."
          />
          <FAQItem
            question="What if I can't stay the full day?"
            answer="No problem! While it's a full-day event (8-10 hours), you can join for whatever time works for you. Even a few hours of participation is valuable."
          />
          <FAQItem
            question="Will there be food and breaks?"
            answer="Yes! We'll have regular breaks throughout the day. For in-person participants, we'll arrange food. Remote participants, keep your favorite snacks handy!"
          />
          <FAQItem
            question="What role does Swaroop play?"
            answer="Swaroop is the teacher and technical decision-maker during Phase 1. He'll guide all foundational choices about platform, tools, workflow, and best practices. Think of him as the technical lead establishing the mobile development standards for Planton. During Phase 2, he'll be available to help, but everyone works on building their own apps using the foundation from Phase 1."
          />
          <FAQItem
            question="Are these apps related to Planton?"
            answer="No! These are generic personal and business apps - a Vehicle Mileage Tracker (for tracking odometer readings) and a Construction Expense Tracker (for real estate construction projects). They're not Planton product features. The goal is to learn mobile app development by building real, useful apps. If your apps need backends, we'll decide in Phase 1 whether to use local storage or cloud backends like Firebase, Supabase, or custom APIs."
          />
          <FAQItem
            question="What platforms are we targeting?"
            answer="This will be decided during Phase 1! Swaroop will help us choose between React Native, Flutter, or Native iOS/Android development based on our goals and constraints. We're leaning toward cross-platform development (apps that work on both Android and iOS), but the final decision will be made together during the hackathon."
          />
          <FAQItem
            question="Is this competitive or collaborative?"
            answer="100% collaborative! This isn't a competition - it's a learning and building experience. The goal is to establish mobile development workflows and help each other succeed."
          />
          <FAQItem
            question="How do I RSVP?"
            answer="Super informal! Just react or comment on the Slack announcement so we can get a rough headcount. Or just show up on the day - no formal registration needed."
          />
        </Box>
      </SectionContainer>
    </PageSection>
  );
};

const FAQItem = ({ 
  question, 
  answer 
}: { 
  question: string; 
  answer: string;
}) => (
  <Box className="bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all">
    <Typography fontSize={18} fontWeight={700} className="text-white mb-3">
      {question}
    </Typography>
    <Typography fontSize={15} className="text-gray-400 leading-relaxed">
      {answer}
    </Typography>
  </Box>
);










