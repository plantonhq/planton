import React from "react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Package, Cloud, Zap, Scale, ArrowRight } from "lucide-react";

export default function Stats() {
	const stats = [
		{
			icon: Package,
			value: "200+",
			label: "Deployment Components",
			description: "Pre-built, battle-tested modules for common infrastructure patterns",
		},
		{
			icon: Cloud,
			value: "8",
			label: "Cloud Providers",
			description: "AWS, GCP, Azure, Kubernetes, Cloudflare, DigitalOcean, Civo, and more",
		},
		{
			icon: Zap,
			value: "2",
			label: "IaC Engines",
			description: "Choose between Pulumi or OpenTofu—same manifests work with both",
		},
		{
			icon: Scale,
			value: "100%",
			label: "Open Source",
			description: "Apache 2.0 licensed. All modules, APIs, and CLI are fully transparent",
		},
	];

	return (
		<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div className="text-center mb-16">
				<h2 className="text-4xl font-bold mb-6">
					<span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
						Built for Production
					</span>
				</h2>
				<p className="text-xl text-slate-400 max-w-3xl mx-auto">
					A mature, comprehensive framework ready for real-world multi-cloud deployments
				</p>
			</div>

			<div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
				{stats.map((stat, index) => {
					const IconComponent = stat.icon;
					return (
						<Card
							key={index}
							className="bg-slate-900/50 border-slate-700 hover:border-purple-900/50 transition-all duration-300"
						>
							<CardContent className="p-8 text-center">
								<div className="w-14 h-14 bg-gradient-to-r from-purple-600 to-purple-400 rounded-2xl flex items-center justify-center mx-auto mb-4">
									<IconComponent className="w-7 h-7 text-white" />
								</div>
								<div className="text-4xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent mb-2">
									{stat.value}
								</div>
								<h3 className="text-lg font-bold text-white mb-3">{stat.label}</h3>
								<p className="text-sm text-slate-400 leading-relaxed">{stat.description}</p>
							</CardContent>
						</Card>
					);
				})}
			</div>

			{/* Additional Trust Signals */}
			<div className="mt-16">
				<div className="flex flex-wrap items-center justify-center gap-6 text-slate-400 mb-8">
					<div className="flex items-center gap-2">
						<span className="text-purple-400 font-bold text-xl">✓</span>
						<span>Published on Buf Schema Registry</span>
					</div>
					<div className="flex items-center gap-2">
						<span className="text-purple-400 font-bold text-xl">✓</span>
						<span>Comprehensive Documentation</span>
					</div>
					<div className="flex items-center gap-2">
						<span className="text-purple-400 font-bold text-xl">✓</span>
						<span>Active Development</span>
					</div>
					<div className="flex items-center gap-2">
						<span className="text-purple-400 font-bold text-xl">✓</span>
						<span>Community Driven</span>
					</div>
				</div>
				
				{/* CTA to browse catalog */}
				<div className="text-center">
					<Link href="/docs/catalog">
						<Button
							size="lg"
							className="rounded-full bg-gradient-to-r from-purple-600 to-pink-600 text-white font-semibold px-8 py-3 h-auto shadow-lg shadow-purple-800/30 transform-gpu transition-all hover:brightness-110 hover:-translate-y-0.5 cursor-pointer"
						>
							Browse All Components
							<ArrowRight className="w-4 h-4 ml-2" />
						</Button>
					</Link>
				</div>
			</div>
		</div>
	);
}

