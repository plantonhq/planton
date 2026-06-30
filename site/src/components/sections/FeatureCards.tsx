import React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Globe, Shield, Code, Zap, Lock, Puzzle } from "lucide-react";

const features = [
  {
    icon: Globe,
    title: "Consistent across clouds",
    description:
      "Same YAML structure across AWS, GCP, Azure, and Kubernetes. Provider-specific fields without artificial abstractions. Every manifest follows the Kubernetes pattern: apiVersion, kind, metadata, spec, status.",
  },
  {
    icon: Shield,
    title: "Catch errors before deployment",
    description:
      "Strongly-typed schemas with field-level validation and sensible defaults. Configuration errors caught immediately with clear messages—before any cloud APIs are called.",
  },
  {
    icon: Code,
    title: "Battle-tested modules included",
    description:
      "Curated Pulumi and Terraform modules maintained for every component. Choose your preferred execution engine—same manifests work with both. Stop writing raw IaC code.",
  },
  {
    icon: Zap,
    title: "Production-ready workflow",
    description:
      "Complete lifecycle management: preview, apply, refresh, destroy. Stack isolation with org/project/stack namespacing. One CLI for all operations across any cloud.",
  },
  {
    icon: Lock,
    title: "Security built-in",
    description:
      "Standard cloud provider authentication via environment variables. Consistent resource labeling for governance. Native support for cross-resource references.",
  },
  {
    icon: Puzzle,
    title: "Fully extensible",
    description:
      "Fork and customize any module. Add new components. Generate SDKs from APIs via Buf Schema Registry. Override any field at runtime with --set flags.",
  },
];

export default function FeatureCards() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
            Why Planton?
          </span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto">
          Built for DevOps engineers, platform teams, and infrastructure architects who need consistency across clouds
        </p>
      </div>
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        {features.map((feature, index) => {
          const IconComponent = feature.icon;
          return (
            <Card
              key={index}
              className="bg-slate-900/50 border-slate-700 hover:bg-slate-900/70 transition-all duration-300 group"
            >
              <CardContent className="p-8">
                <div className="flex items-center gap-4 mb-4">
                  <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-cyan-400 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <IconComponent className="w-6 h-6 text-white" />
                  </div>
                  <h3 className="text-xl font-bold text-white">{feature.title}</h3>
                </div>
                <p className="text-slate-400 leading-relaxed">{feature.description}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}


