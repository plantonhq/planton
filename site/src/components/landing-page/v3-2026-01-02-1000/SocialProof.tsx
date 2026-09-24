'use client';

import { Box, Stack, Typography } from '@mui/material';
import Image from 'next/image';
import { FC } from 'react';
import { Section, SectionTitle, Card, Quote, Badge, TestimonialCard } from './shared';

const customers = [
  {
    company: 'TYNYBAY',
    companyLogo: '/_site/images/customers/logos/tynybay.png',
    type: 'IT Consulting Firm',
    cloud: 'AWS, GCP',
    teamSize: '1 DevOps + 1 Dev',
    economics: {
      before: '3 clients max',
      after: '8+ clients',
      savings: '2.5x capacity',
    },
    quote: 'For a client who mandated GCP despite zero team experience, Planton delivered the entire infrastructure. Using Planton for all future client projects.',
    result: 'Scaled from 3 to 8+ concurrent clients without growing the ops team.',
  },
  {
    company: 'iorta TechNext',
    companyLogo: '/_site/images/customers/logos/iorta.svg',
    type: 'BFSI Product Company',
    cloud: 'AWS (ECS)',
    teamSize: '7 developers',
    economics: {
      before: 'Devs blocked on infra',
      after: 'Self-serve deploys',
      savings: 'Junior-led ops',
    },
    quote: 'Planton enabled me to provide a very mature developer experience to our entire 7-person dev team.',
    result: 'Runs production on AWS ECS without growing the ops team.',
  },
];

const testimonials = [
  {
    name: 'Balaji Borra',
    role: 'DevOps Engineer',
    company: 'TYNYBAY',
    location: 'India',
    quote: 'I handle 8+ client projects with Planton\u2014no more rewriting Terraform between clients.',
    avatar: '/_site/images/customers/people/balaji-borra.png',
  },
  {
    name: 'Rakesh Kandhi',
    role: 'Senior Developer',
    company: 'TYNYBAY',
    location: 'India',
    quote: 'I deploy to dev, staging, and prod without waiting on DevOps. Self-service is a game changer.',
    avatar: '/_site/images/customers/people/rakesh-kandhi.jpeg',
  },
  {
    name: 'Sai Saketh',
    role: 'DevOps Engineer',
    company: 'iorta TechNext',
    location: 'India',
    quote: 'Our 7-person team deploys independently without deep AWS expertise.',
    avatar: null,
  },
];

export const SocialProof: FC = () => {
  return (
    <Section id="social-proof">
      <Stack className="items-center text-center mb-12">
        <Badge className="mb-6">PROVEN IMPACT</Badge>
        <SectionTitle>
          Teams Ship Faster with Planton
        </SectionTitle>
      </Stack>

      <Box className="grid md:grid-cols-2 gap-6 mb-10">
        {customers.map((customer, index) => (
          <Card key={index} className="p-0 overflow-hidden">
            <Box className="p-4 border-b border-[#2a2a2a] bg-[#0f0f0f]">
              <Box className="flex items-center gap-4">
                <Box className="w-10 h-10 rounded-lg bg-[#1a1a1a] p-2 flex items-center justify-center">
                  <Image
                    src={customer.companyLogo}
                    alt={customer.company}
                    width={24}
                    height={24}
                    className="w-full h-full object-contain brightness-0 invert"
                  />
                </Box>
                <Box>
                  <Typography className="text-white font-semibold">{customer.company}</Typography>
                  <Typography className="text-xs text-[#666]">{customer.type}</Typography>
                </Box>
                <Box className="ml-auto flex gap-2">
                  <Badge>{customer.cloud}</Badge>
                  <Badge>{customer.teamSize}</Badge>
                </Box>
              </Box>
            </Box>

            <Box className="p-6 bg-[#0a0a0a]">
              <Box className="p-4 rounded-lg bg-[#151515] border border-[#2a2a2a] mb-6">
                <Box className="grid grid-cols-3 gap-4 text-center">
                  <Box>
                    <Typography className="text-xs text-[#555] mb-1">Before</Typography>
                    <Typography className="text-sm text-[#666] line-through">{customer.economics.before}</Typography>
                  </Box>
                  <Box>
                    <Typography className="text-xs text-[#555] mb-1">After</Typography>
                    <Typography className="text-sm text-white font-medium">{customer.economics.after}</Typography>
                  </Box>
                  <Box>
                    <Typography className="text-xs text-[#555] mb-1">Impact</Typography>
                    <Typography className="text-lg font-bold text-white">{customer.economics.savings}</Typography>
                  </Box>
                </Box>
              </Box>

              <Quote text={customer.quote} author={customer.company} companyLogo={customer.companyLogo} />

              <Box className="mt-4 pt-4 border-t border-[#2a2a2a]">
                <Typography className="text-sm text-[#a0a0a0]">
                  <span className="text-white font-medium">Result:</span> {customer.result}
                </Typography>
              </Box>
            </Box>
          </Card>
        ))}
      </Box>

      <Box className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {testimonials.map((testimonial, index) => (
          <TestimonialCard
            key={index}
            name={testimonial.name}
            role={testimonial.role}
            company={testimonial.company}
            quote={testimonial.quote}
            location={testimonial.location}
            avatar={testimonial.avatar}
            className="h-full"
          />
        ))}
      </Box>
    </Section>
  );
};
