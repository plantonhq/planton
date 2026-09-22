'use client';

import {
  Slide,
  SlideHeader,
  Grid,
  Card,
  CardTitle,
  CardText,
  Callout,
} from '@/components/deck';

const capabilities = [
  {
    title: 'Backend SDKs',
    description:
      'Agents, sessions, skills, and execution — delivered as SDKs, not homework. Planton integrates through the official SDKs and runs no agent infrastructure of its own.',
  },
  {
    title: 'Frontend Components',
    description:
      'The hidden hard part of AI products: streaming chat, tool-call rendering, approval prompts. Stigmer ships them as drop-in components for web, desktop, and mobile — the chat in the demo was these components, themed.',
  },
  {
    title: 'Tenant Operations',
    description:
      'Your customers become tenants inside Stigmer, with usage caps and spend limits built in — the operations layer every AI product needs and nobody wants to build.',
  },
];

export default function S09TheEngineBehindTheAssistant() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="One More Thing"
        title="The Conversation You Just Watched Runs on Stigmer"
        subtitle="Stigmer is the platform for building AI agent products — the backend, the frontend, and the tenant operations, so you ship the capability instead of the plumbing."
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {capabilities.map((item) => (
          <Card key={item.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {item.title}
            </CardTitle>
            <CardText>{item.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Planton didn&apos;t build an AI stack. It adopted one — zero
          first-party agent infrastructure, and an AI-native product on every
          surface.
        </p>
      </Callout>
    </Slide>
  );
}
