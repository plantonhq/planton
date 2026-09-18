'use client';

import React from 'react';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { Lightbulb, Package, GitBranch, Zap, ChevronRight } from 'lucide-react';

export default function InfraChartsIntro() {
  return (
    <div className="h-full overflow-y-auto">
      <div className="max-w-7xl mx-auto p-8 space-y-8">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="text-center"
        >
          <h1 className="text-4xl md:text-5xl font-black text-gray-900 mb-4">
            Infra Charts
          </h1>
          <div className="inline-block bg-gradient-to-r from-white to-[#666] text-white px-6 py-3 rounded-full mb-4">
            <p className="text-lg font-bold">
              Helm Charts for Kubernetes. Infra Charts for Cloud.
            </p>
          </div>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            You&apos;ve seen how individual Lego Blocks work. Now let&apos;s see how we combine multiple blocks into complete, production-ready environments.
          </p>
        </motion.div>

        {/* The Problem */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.6 }}
          className="bg-red-50 border-2 border-red-200 rounded-2xl p-8"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center gap-3">
            <div className="w-10 h-10 bg-red-600 rounded-lg flex items-center justify-center">
              <Lightbulb className="w-6 h-6 text-white" />
            </div>
            The Challenge
          </h2>
          <div className="space-y-3 text-gray-700">
            <p className="text-lg">
              Real production environments aren&apos;t just one resource. An ECS environment needs:
            </p>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mt-4">
              {[
                'VPC with subnets',
                'Security groups',
                'Application Load Balancer',
                'SSL certificates',
                'Route53 DNS',
                'ECS cluster',
                'IAM roles',
                'ECR repositories',
                'And more...'
              ].map((item, i) => (
                <div key={i} className="bg-white rounded-lg px-4 py-2 text-sm font-medium border border-red-200">
                  • {item}
                </div>
              ))}
            </div>
            <p className="text-lg mt-4 font-semibold text-red-900">
              Manually deploying and connecting these is complex, error-prone, and time-consuming.
            </p>
          </div>
        </motion.div>

        {/* The Inspiration */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4, duration: 0.6 }}
          className="bg-gradient-to-r from-[#151515] to-[#151515] border-2 border-[#2a2a2a] rounded-2xl p-8"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center gap-3">
            <div className="w-10 h-10 bg-[#111] rounded-lg flex items-center justify-center">
              <Package className="w-6 h-6 text-white" />
            </div>
            The Inspiration: Helm Charts
          </h2>
          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white rounded-xl p-6 border border-[#2a2a2a]">
              <h3 className="font-bold text-lg mb-3 text-white">Kubernetes Helm Charts</h3>
              <p className="text-gray-700 mb-3">
                Package multiple Kubernetes resources together as reusable templates
              </p>
              <div className="space-y-2 text-sm text-gray-600">
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4 text-white" />
                  <span>Deploy complete applications</span>
                </div>
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4 text-white" />
                  <span>Parameterized templates</span>
                </div>
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4 text-white" />
                  <span>Reusable across environments</span>
                </div>
              </div>
            </div>

            <div className="bg-gradient-to-br from-white to-[#666] rounded-xl p-6 text-white">
              <h3 className="font-bold text-lg mb-3">Infra Charts</h3>
              <p className="mb-3 text-white/90">
                Package multiple Lego Blocks together as reusable infrastructure templates
              </p>
              <div className="space-y-2 text-sm">
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4" />
                  <span>Deploy complete environments</span>
                </div>
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4" />
                  <span>Parameterized templates</span>
                </div>
                <div className="flex items-center gap-2">
                  <ChevronRight className="w-4 h-4" />
                  <span>Reusable across projects</span>
                </div>
              </div>
            </div>
          </div>
        </motion.div>

        {/* What It Does */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6, duration: 0.6 }}
          className="bg-[#151515] border-2 border-[#2a2a2a] rounded-2xl p-8"
        >
          <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-3">
            <div className="w-10 h-10 bg-[#10b981] rounded-lg flex items-center justify-center">
              <Zap className="w-6 h-6 text-white" />
            </div>
            What Infra Charts Do
          </h2>
          <div className="grid md:grid-cols-3 gap-6">
            <div className="bg-white rounded-xl p-6 border border-[#2a2a2a]">
              <GitBranch className="w-8 h-8 text-[#10b981] mb-3" />
              <h3 className="font-bold mb-2">Orchestrated Deployment</h3>
              <p className="text-sm text-gray-700">
                Automatically handles dependencies and deploys resources in the correct order using a DAG
              </p>
            </div>
            <div className="bg-white rounded-xl p-6 border border-[#2a2a2a]">
              <Package className="w-8 h-8 text-[#10b981] mb-3" />
              <h3 className="font-bold mb-2">Complete Environments</h3>
              <p className="text-sm text-gray-700">
                Package all resources needed for production: VPC, security, compute, storage, DNS, SSL
              </p>
            </div>
            <div className="bg-white rounded-xl p-6 border border-[#2a2a2a]">
              <Zap className="w-8 h-8 text-[#10b981] mb-3" />
              <h3 className="font-bold mb-2">20 Minutes to Production</h3>
              <p className="text-sm text-gray-700">
                What used to take weeks of manual work now deploys in ~20 minutes
              </p>
            </div>
          </div>
        </motion.div>

        {/* Real Example - ECS Chart DAG */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.8, duration: 0.6 }}
          className="bg-white rounded-2xl border-2 border-gray-200 overflow-hidden demo-card-shadow"
        >
          <div className="bg-gradient-to-r from-[#999] to-red-600 px-8 py-5">
            <h3 className="font-bold text-2xl text-white mb-1">Real Example: AWS ECS Chart</h3>
            <p className="text-white/90">
              9 interconnected resources deployed as an orchestrated DAG—all from a simple 7-parameter form
            </p>
          </div>
          <div className="p-8 bg-gray-50">
            <Image
              src="/demo/_site/images/ecs-chart-dag.png"
              alt="ECS Infra Chart DAG showing VPC, Route53, security groups, ALB, certificates, ECS cluster, IAM roles, ECR repo, and ECS service with their dependency arrows and deployment order"
              width={2674}
              height={1430}
              className="w-full h-auto rounded-xl border-2 border-gray-200 demo-card-shadow"
            />
          </div>
        </motion.div>

        {/* Customer Success */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.0, duration: 0.6 }}
          className="bg-gradient-to-r from-white to-[#666] rounded-2xl p-8 text-white text-center"
        >
          <div className="max-w-3xl mx-auto">
            <h2 className="text-3xl font-bold mb-4">Real Impact</h2>
            <div className="grid md:grid-cols-3 gap-6 mb-6">
              <div>
                <div className="text-4xl font-black mb-2">3</div>
                <div className="text-sm text-white/90">Different clients</div>
              </div>
              <div>
                <div className="text-4xl font-black mb-2">~20 min</div>
                <div className="text-sm text-white/90">Complete infrastructure</div>
              </div>
              <div>
                <div className="text-4xl font-black mb-2">&lt;1 hour</div>
                <div className="text-sm text-white/90">Code to live URL</div>
              </div>
            </div>
            <p className="text-lg text-white/90 italic">
              &ldquo;Every single time they need to duplicate terraform code → solved&rdquo;
            </p>
          </div>
        </motion.div>
      </div>
    </div>
  );
}

