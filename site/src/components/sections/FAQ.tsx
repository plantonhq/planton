"use client";
import React, { useState } from "react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ChevronDown, ChevronRight, Github } from "lucide-react";

export default function FAQ() {
  const [openItems, setOpenItems] = useState<Set<number>>(new Set());

  const toggleItem = (index: number) => {
    const newOpenItems = new Set(openItems);
    if (newOpenItems.has(index)) newOpenItems.delete(index);
    else newOpenItems.add(index);
    setOpenItems(newOpenItems);
  };

  const faqs = [
    {
      question: "Is this just another abstraction layer that hides cloud-specific details?",
      answer:
        "No. OpenMCF manifests are deliberately provider-specific (AWS RDS has different fields than GCP Cloud SQL). The framework provides consistency in structure and workflow, not artificial abstractions that leak complexity. You write the exact configuration your cloud provider needs, just in a standardized YAML format.",
    },
    {
      question: "Is this locked to a SaaS platform?",
      answer:
        "No. OpenMCF is a standalone CLI tool with zero SaaS dependencies. Your manifests, your credentials, your infrastructure. Everything runs locally or in your CI/CD. There's no platform to sign up for, no API to call, no vendor to depend on.",
    },
    {
      question: "Why would I use this instead of just learning Terraform or Pulumi?",
      answer:
        "If you're deploying across multiple clouds, you'll end up learning Terraform AND Pulumi AND cloud-specific nuances anyway. OpenMCF gives you battle-tested modules, strong validation, and consistent workflow so you can focus on your infrastructure requirements, not tooling complexity.",
    },
    {
      question: "What if I need to customize beyond what the modules provide?",
      answer:
        "All Pulumi and Terraform modules are open source and forkable. Point --module-dir at your fork. Or use the manifests as validated input and call the modules directly from your own IaC code. You're never locked in.",
    },
    {
      question: "Do I need Pulumi/Terraform knowledge to get started?",
      answer:
        "Not to start. You work with declarative manifests; the CLI selects and executes the appropriate Pulumi or OpenTofu modules for you. But knowledge of these tools helps when you need deeper customization.",
    },
    {
      question: "Which cloud providers are supported today?",
      answer:
        "AWS, GCP, Azure, Kubernetes, plus Cloudflare, DigitalOcean, Civo, Confluent, MongoDB Atlas, and more. Browse the full catalog at /docs/catalog to see all 200+ deployment components across 8 providers.",
    },
    {
      question: "How are configuration errors caught before deployment?",
      answer:
        "Each manifest is validated against Protobuf schemas with field-level rules (via Buf ProtoValidate, including CEL expressions). Running 'openmcf validate' catches errors immediately with clear messages—before any cloud APIs are called.",
    },
    {
      question: "Where is infrastructure state stored?",
      answer:
        "OpenMCF creates no state management abstractions—it uses the native mechanisms of your chosen IaC engine. With OpenTofu, you configure backends (local, S3, GCS, azurerm) directly via tofu init. With Pulumi, you use standard Pulumi backends (Pulumi Cloud, S3, GCS, or local filesystem) configured via pulumi login. Whatever backend configuration you set in your environment is what gets used. You have complete control over where and how state is stored.",
    },
    {
      question: "How do credentials work?",
      answer:
        "Provider credentials are read from environment variables (e.g., AWS_ACCESS_KEY_ID, GOOGLE_APPLICATION_CREDENTIALS, KUBECONFIG). The CLI uses standard cloud provider authentication mechanisms. Optionally, you can pass credential file paths via flags like --aws-credential or --gcp-credential. Credentials never leave your local environment.",
    },
    {
      question: "Can I run this in CI/CD?",
      answer:
        "Yes. Install the CLI and your chosen engine (Pulumi or OpenTofu), provide credentials as files or secrets, then run the same commands you use locally. Works in GitHub Actions, GitLab CI, Jenkins, or any CI system.",
    },
  ];

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
            Frequently Asked Questions
          </span>
        </h2>
        <p className="text-xl text-slate-400">Everything you need to know about OpenMCF</p>
      </div>
      <div className="space-y-4">
        {faqs.map((faq, index) => {
          const isOpen = openItems.has(index);
          return (
            <Card key={index} className="bg-slate-900/30 border-slate-700">
              <CardContent className="p-0">
                <button
                  className="w-full text-left p-6 flex items-center justify-between hover:bg-slate-800/50 transition-colors"
                  onClick={() => toggleItem(index)}
                >
                  <h3 className="text-lg font-semibold text-white pr-4">{faq.question}</h3>
                  {isOpen ? (
                    <ChevronDown className="w-5 h-5 text-slate-400 flex-shrink-0" />
                  ) : (
                    <ChevronRight className="w-5 h-5 text-slate-400 flex-shrink-0" />
                  )}
                </button>
                {isOpen && (
                  <div className="px-6 pb-6">
                    <p className="text-slate-400 leading-relaxed">{faq.answer}</p>
                  </div>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Final CTA */}
      <div className="mt-16 text-center">
        <h3 className="text-2xl font-bold text-white mb-4">Ready to Get Started?</h3>
        <p className="text-lg text-slate-400 mb-8">
          Install the CLI and deploy your first resource in minutes
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <div className="bg-slate-900 rounded-lg px-6 py-3 font-mono text-sm border border-slate-700">
            <span className="text-slate-500">$</span> <span className="text-purple-400">brew install</span> <span className="text-white">plantonhq/tap/openmcf</span>
          </div>
          <Link href="https://github.com/plantonhq/openmcf" target="_blank" rel="noopener noreferrer">
            <Button
              size="lg"
              variant="outline"
              className="rounded-full border border-slate-600 text-slate-100 bg-transparent hover:bg-slate-800 hover:border-slate-500 px-6 py-3 h-auto cursor-pointer"
            >
              <Github className="w-4 h-4 mr-2" />
              View on GitHub
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}


