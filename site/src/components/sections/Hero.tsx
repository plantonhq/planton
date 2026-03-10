"use client";
import React, { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Copy, Check, ArrowRight } from "lucide-react";

export default function Hero() {
	const [copied, setCopied] = useState(false);
	const installCommand = "brew install plantonhq/tap/openmcf";

	const copyToClipboard = () => {
		navigator.clipboard.writeText(installCommand);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};

	return (
		<div className="relative overflow-hidden font-sans">
			{/* Deep purple gradient to match production */}
			<div className="absolute inset-0 bg-gradient-to-b from-[#0b0713] via-[#140a26] to-[#0a0912]" />
			{/* Radial magenta glow */}
			<div className="absolute inset-0 pointer-events-none bg-[radial-gradient(1200px_500px_at_50%_8%,rgba(217,70,239,0.18),rgba(0,0,0,0))]" />

			<div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="min-h-[calc(100vh-64px)] flex items-center justify-center">
					<div className="text-center max-w-4xl mx-auto w-full px-2 pt-20">
						{/* Badges */}
						<div className="flex flex-wrap justify-center gap-2 mb-10">
							<Badge variant="outline" className="rounded-full border-white/15 text-white/80 bg-white/10">Apache-2.0</Badge>
							<Badge variant="outline" className="rounded-full border-white/15 text-white/80 bg-white/10">KRM/Protobuf/Buf</Badge>
							<Badge variant="outline" className="rounded-full border-white/15 text-white/80 bg-white/10">Pulumi/OpenTofu</Badge>
							<Badge variant="outline" className="rounded-full border-white/15 text-white/80 bg-white/10">CLI-first</Badge>
							<Badge variant="outline" className="rounded-full border-white/15 text-white/80 bg-white/10">CI/CD-ready</Badge>
						</div>

						{/* Heading */}
						<h1 className="text-3xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 w-full break-words">
							<span className="block">
								<span className="text-white">An </span>
								<span className="text-[#7a4183]">Open </span>
								<span className="text-white">Source </span>
							</span>
							<span className="block">
								<span className="text-[#7a4183]">M</span>
								<span className="text-white">ulti‑</span>
								<span className="text-[#7a4183]">C</span>
								<span className="text-white">loud </span>
								{/* <span className="text-white">Infrastructure </span> */}
								<span className="text-[#7a4183]">F</span>
								<span className="text-white">ramework</span>
							</span>
						</h1>

						{/* Subheading */}
						<p className="text-base sm:text-xl lg:text-2xl text-slate-200/90 font-medium leading-relaxed max-w-3xl mx-auto mb-10">
						Stop learning different tools for every cloud. Write declarative YAML once, deploy to AWS, GCP, Azure, and Kubernetes with the same CLI and workflow. No vendor lock‑in, no artificial abstractions—just consistent infrastructure deployment everywhere.
						</p>

						{/* CTAs */}
						<div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
							<Button
								size="lg"
								onClick={copyToClipboard}
								className="cursor-pointer rounded-full bg-gradient-to-r from-[#7a4183] to-[#d946ef] text-white font-mono text-xs sm:text-sm px-4 sm:px-8 py-3 h-auto shadow-lg shadow-fuchsia-800/30 transform-gpu transition-all hover:brightness-110 hover:-translate-y-0.5 focus-visible:ring-2 focus-visible:ring-fuchsia-400/50 w-full sm:w-auto"
							>
								<span className="flex items-center whitespace-nowrap overflow-hidden text-ellipsis">
								{copied ? (
									<>
										<Check className="w-4 h-4 mr-2 flex-shrink-0" />
										Copied!
									</>
								) : (
									<>
										<Copy className="w-4 h-4 mr-2 flex-shrink-0" />
										<span className="whitespace-nowrap overflow-hidden text-ellipsis">{installCommand}</span>
									</>
								)}
								</span>
							</Button>

					<Link href="/docs/catalog">
						<Button
							size="lg"
							variant="outline"
							className="cursor-pointer rounded-full border border-white/20 text-slate-100 bg-transparent hover:bg-white/5 hover:border-white/40 px-8 py-3 h-auto transform-gpu transition-all hover:-translate-y-0.5 focus-visible:ring-2 focus-visible:ring-fuchsia-400/40"
						>
							Browse 360+ Components from 17 Providers
							<ArrowRight className="w-4 h-4 ml-2" />
						</Button>
					</Link>
				</div>

					{/* Tagline */}
					<div className="mt-16 pt-12 border-t border-white/10">
						<p className="text-2xl sm:text-3xl font-bold text-center mb-8">
							<span className="text-slate-300">Define once.</span>
							<span className="ml-3 bg-gradient-to-r from-[#f0abfc] via-[#f472b6] to-[#d946ef] bg-clip-text text-transparent">Deploy anywhere.</span>
						</p>
						
				{/* Cloud Provider Icons */}
				<div className="flex flex-wrap items-center justify-center gap-4 sm:gap-6 md:gap-8 mt-8">
					<Image src="/images/providers/aws.svg" alt="AWS" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/gcp.svg" alt="GCP" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/azure.svg" alt="Azure" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/kubernetes.svg" alt="Kubernetes" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/digital-ocean.svg" alt="DigitalOcean" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/civo.svg" alt="Civo" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/cloudflare.svg" alt="Cloudflare" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/confluent.svg" alt="Confluent" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/mongodb-atlas.svg" alt="MongoDB Atlas" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
					<Image src="/images/providers/auth0.svg" alt="Auth0" width={40} height={40} className="h-6 sm:h-8 md:h-10 w-auto opacity-80 hover:opacity-100 transition-opacity" />
				</div>
					</div>
					</div>
				</div>
			</div>
		</div>
	);
}


