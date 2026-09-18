'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import { motion } from 'framer-motion';
import { Play, CheckCircle2, Loader2 } from 'lucide-react';

export default function InfraChartDeployment() {
  const [isDeploying, setIsDeploying] = useState(false);

  const handleDeploy = () => {
    setIsDeploying(true);
  };

  return (
    <div className="h-full flex flex-col bg-white">
      {/* Header */}
      <div className="px-8 py-6 border-b border-gray-200 bg-gradient-to-r from-[#151515] to-red-50">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <div className="inline-block bg-gradient-to-r from-white to-[#666] text-white px-4 py-2 rounded-full mb-3 text-sm font-bold">
            Helm Charts for Kubernetes. Infra Charts for Cloud.
          </div>
          <h2 className="text-3xl font-black text-gray-900 mb-2">
            Deploy AWS ECS Environment
          </h2>
          <p className="text-gray-600">
            Configure input parameters to deploy this chart. The template will be rendered with your values and deployed as an orchestrated DAG.
          </p>
        </motion.div>
      </div>

      {/* Main Content - Form + DAG side by side */}
      <div className="flex-1 overflow-hidden flex">
        {/* Left Side - Form (40%) */}
        <div className="w-[40%] border-r border-gray-200 overflow-y-auto p-8 bg-gray-50">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6 }}
          >
              {/* Educational Header */}
              <div className="bg-[#151515] border border-[#2a2a2a] rounded-xl p-6 mb-8">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 bg-[#111] rounded-lg flex items-center justify-center flex-shrink-0">
                    <Play className="w-5 h-5 text-white" />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg text-gray-900 mb-2">
                      Parameterized Infrastructure Template
                    </h3>
                    <p className="text-sm text-gray-700 leading-relaxed">
                      This form represents the values.yaml from the Infra Chart. 
                      Your inputs will be used to render all Lego Block configurations within the chart. 
                      Each block will then be deployed based on the dependency DAG.
                    </p>
                  </div>
                </div>
              </div>

              {/* Form Fields */}
              <div className="bg-white rounded-2xl border-2 border-gray-200 demo-card-shadow overflow-hidden">
                {/* Metadata Section */}
                <div className="border-b border-gray-200 bg-gray-50 px-6 py-3">
                  <h4 className="font-bold text-gray-900">Metadata</h4>
                </div>
                <div className="p-6 space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Environment
                    </label>
                    <input
                      type="text"
                      value="dev"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                    <p className="text-xs text-gray-500 mt-1">Environment name (dev, staging, prod)</p>
                  </div>
                </div>

                {/* Values Section */}
                <div className="border-b border-gray-200 bg-gray-50 px-6 py-3">
                  <h4 className="font-bold text-gray-900">Chart Parameters</h4>
                </div>
                <div className="p-6 space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      First AZ for the public / private subnet pair
                    </label>
                    <input
                      type="text"
                      value="us-east-1a"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Second AZ for the public / private subnet pair
                    </label>
                    <input
                      type="text"
                      value="us-east-1b"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Route 53 Hosted-Zone domain
                    </label>
                    <input
                      type="text"
                      value="example.com"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      DNS name served by the ALB
                    </label>
                    <input
                      type="text"
                      value="app.example.com"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      ECS service & task family name
                    </label>
                    <input
                      type="text"
                      value="nginx"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      ECR repository for your service images
                    </label>
                    <input
                      type="text"
                      value="shopping-cart-service"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Container port the task listens on
                    </label>
                    <input
                      type="text"
                      value="80"
                      readOnly
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-700"
                    />
                  </div>
                </div>
              </div>

              {/* Deploy Button */}
              <div className="mt-8 flex justify-center">
                <button
                  onClick={handleDeploy}
                  disabled={isDeploying}
                  className={`w-full px-6 py-3 rounded-xl font-bold text-base transition-all duration-300 flex items-center justify-center gap-2 ${
                    isDeploying
                      ? 'bg-gray-400 text-white cursor-not-allowed'
                      : 'bg-gradient-to-r from-white to-[#666] text-white hover:from-white hover:to-[#666] shadow-lg hover:shadow-xl'
                  }`}
                >
                  {isDeploying ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Deploying...
                    </>
                  ) : (
                    <>
                      <Play className="w-5 h-5" />
                      Deploy Infra Chart
                    </>
                  )}
                </button>
              </div>
            </motion.div>
          </div>

          {/* Right Side - DAG Visualization (60%) */}
          <div className="w-[60%] bg-white overflow-y-auto">
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6 }}
              className="h-full flex flex-col"
            >
              {/* DAG Header */}
              <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-[#151515] to-[#151515]">
                <h3 className="font-bold text-lg text-gray-900 mb-1">
                  Deployment DAG Preview
                </h3>
                <p className="text-sm text-gray-600">
                  9 resources will be deployed in dependency order
                </p>
              </div>

              {/* DAG Image */}
              <div className="flex-1 p-6 flex items-center justify-center bg-gray-50">
                <div className="w-full max-w-3xl">
                  <Image
                    src="/demo/_site/images/ecs-chart-dag.png"
                    alt="ECS Infra Chart DAG showing 9 interconnected resources: VPC, Route53, Security Group, Certificate, ALB, ECS Cluster, IAM Role, ECR Repo, and ECS Service with dependency arrows"
                    width={2674}
                    height={1430}
                    className="w-full h-auto rounded-lg border-2 border-gray-200 demo-card-shadow"
                  />
                </div>
              </div>

              {/* Stats Footer */}
              {isDeploying && (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.3 }}
                  className="px-6 py-4 bg-[#151515] border-t border-[#2a2a2a]"
                >
                  <div className="flex items-center gap-3 mb-3">
                    <CheckCircle2 className="w-5 h-5 text-[#10b981]" />
                    <span className="font-semibold text-gray-900">Deployment Started</span>
                  </div>
                  <div className="grid grid-cols-3 gap-4 text-center text-sm">
                    <div>
                      <div className="text-2xl font-black text-white">9</div>
                      <div className="text-gray-600">Resources</div>
                    </div>
                    <div>
                      <div className="text-2xl font-black text-white">~20</div>
                      <div className="text-gray-600">Minutes</div>
                    </div>
                    <div>
                      <div className="text-2xl font-black text-white">
                        <Loader2 className="w-6 h-6 animate-spin inline-block" />
                      </div>
                      <div className="text-gray-600">In Progress</div>
                    </div>
                  </div>
                  <p className="text-center text-gray-600 mt-3 text-xs">
                    Each node deploys when dependencies are satisfied. Independent resources execute in parallel.
                  </p>
                </motion.div>
              )}
            </motion.div>
          </div>
      </div>
    </div>
  );
}

