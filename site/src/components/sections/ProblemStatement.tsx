import React from "react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { X, Check, BookOpen } from "lucide-react";

export default function ProblemStatement() {
	return (
		<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div className="text-center mb-16">
				<h2 className="text-4xl font-bold mb-6">
					<span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
						The Multi-Cloud Problem
					</span>
				</h2>
				<p className="text-xl text-slate-400 max-w-3xl mx-auto">
					Deploying the same infrastructure across different clouds means learning completely different tools, CLIs, and workflows
				</p>
			</div>

			{/* Example: Deploying PostgreSQL */}
			<div className="mb-12">
				<h3 className="text-2xl font-bold text-white text-center mb-8">
					Example: Deploying PostgreSQL
				</h3>

				<div className="grid md:grid-cols-3 gap-6 mb-12">
					{/* AWS */}
					<Card className="bg-slate-900/50 border-red-900/30">
						<CardContent className="p-6">
							<div className="flex items-center gap-2 mb-4">
								<X className="w-5 h-5 text-red-400" />
								<h4 className="text-lg font-bold text-white">AWS RDS</h4>
							</div>
							<div className="bg-slate-950 rounded-lg p-4 font-mono text-xs overflow-x-auto">
								<div className="text-slate-400">aws rds create-db-instance \</div>
								<div className="text-slate-400 ml-2">--db-instance-identifier mydb \</div>
								<div className="text-slate-400 ml-2">--db-instance-class db.t3.medium \</div>
								<div className="text-slate-400 ml-2">--engine postgres \</div>
								<div className="text-slate-400 ml-2">--master-username admin \</div>
								<div className="text-slate-400 ml-2">--allocated-storage 100</div>
							</div>
							<p className="text-sm text-slate-500 mt-4">
								Learn: CloudFormation, IAM roles, Security Groups, Parameter Groups...
							</p>
						</CardContent>
					</Card>

					{/* GCP */}
					<Card className="bg-slate-900/50 border-red-900/30">
						<CardContent className="p-6">
							<div className="flex items-center gap-2 mb-4">
								<X className="w-5 h-5 text-red-400" />
								<h4 className="text-lg font-bold text-white">GCP Cloud SQL</h4>
							</div>
							<div className="bg-slate-950 rounded-lg p-4 font-mono text-xs overflow-x-auto">
								<div className="text-slate-400">gcloud sql instances create mydb \</div>
								<div className="text-slate-400 ml-2">--tier db-n1-standard-1 \</div>
								<div className="text-slate-400 ml-2">--database-version POSTGRES_13 \</div>
								<div className="text-slate-400 ml-2">--region us-central1 \</div>
								<div className="text-slate-400 ml-2">--storage-size 100GB</div>
							</div>
							<p className="text-sm text-slate-500 mt-4">
								Learn: Deployment Manager, Cloud IAM, Authorized Networks, Flags...
							</p>
						</CardContent>
					</Card>

					{/* Azure */}
					<Card className="bg-slate-900/50 border-red-900/30">
						<CardContent className="p-6">
							<div className="flex items-center gap-2 mb-4">
								<X className="w-5 h-5 text-red-400" />
								<h4 className="text-lg font-bold text-white">Azure PostgreSQL</h4>
							</div>
							<div className="bg-slate-950 rounded-lg p-4 font-mono text-xs overflow-x-auto">
								<div className="text-slate-400">az postgres server create \</div>
								<div className="text-slate-400 ml-2">--name mydb \</div>
								<div className="text-slate-400 ml-2">--resource-group mygroup \</div>
								<div className="text-slate-400 ml-2">--sku-name GP_Gen5_2 \</div>
								<div className="text-slate-400 ml-2">--storage-size 102400</div>
							</div>
							<p className="text-sm text-slate-500 mt-4">
								Learn: ARM templates, Azure AD, Firewall Rules, Server Parameters...
							</p>
						</CardContent>
					</Card>
				</div>

				<div className="text-center mb-12">
					<p className="text-xl font-bold text-red-400">
						Three different CLIs. Three different terminologies. Three different mental models.
					</p>
					<p className="text-lg text-slate-400 mt-2">
						For the same thing: a PostgreSQL database.
					</p>
				</div>
			</div>

			{/* The Planton Way */}
			<div className="mb-12">
				<h3 className="text-2xl font-bold text-center mb-8">
					<span className="bg-gradient-to-r from-[#f0abfc] via-[#f472b6] to-[#d946ef] bg-clip-text text-transparent">
						The Planton Way
					</span>
				</h3>

				<Card className="bg-slate-900/50 border-purple-900/30 max-w-7xl mx-auto">
					<CardContent className="p-8">
						<div className="flex items-center gap-2 mb-6">
							<Check className="w-6 h-6 text-emerald-400" />
							<h4 className="text-xl font-bold text-white">Same Structure. Same Workflow. Any Cloud.</h4>
						</div>

						<div className="grid md:grid-cols-3 gap-4">
							{/* AWS Example */}
							<div>
								<p className="text-sm font-bold text-purple-400 mb-2">AWS RDS</p>
								<div className="bg-slate-950 rounded-lg p-3 font-mono text-xs whitespace-nowrap overflow-x-auto">
									<div className="text-blue-400">apiVersion: <span className="text-white">aws.planton.dev/v1</span></div>
									<div className="text-blue-400">kind: <span className="text-white">AwsRdsInstance</span></div>
									<div className="text-blue-400">metadata:</div>
									<div className="text-blue-400 ml-2">name: <span className="text-white">mydb</span></div>
									<div className="text-blue-400">spec:</div>
									<div className="text-blue-400 ml-2">engine: <span className="text-white">postgres</span></div>
									<div className="text-blue-400 ml-2">engine_version: <span className="text-white">&quot;15.4&quot;</span></div>
									<div className="text-blue-400 ml-2">instance_class: <span className="text-white">db.t3.medium</span></div>
									<div className="text-blue-400 ml-2">allocated_storage_gb: <span className="text-emerald-400">100</span></div>
								</div>
							</div>

							{/* GCP Example */}
							<div>
								<p className="text-sm font-bold text-purple-400 mb-2">GCP Cloud SQL</p>
								<div className="bg-slate-950 rounded-lg p-3 font-mono text-xs whitespace-nowrap overflow-x-auto">
									<div className="text-blue-400">apiVersion: <span className="text-white">gcp.planton.dev/v1</span></div>
									<div className="text-blue-400">kind: <span className="text-white">GcpCloudSql</span></div>
									<div className="text-blue-400">metadata:</div>
									<div className="text-blue-400 ml-2">name: <span className="text-white">mydb</span></div>
									<div className="text-blue-400">spec:</div>
									<div className="text-blue-400 ml-2">database_engine: <span className="text-white">POSTGRESQL</span></div>
									<div className="text-blue-400 ml-2">tier: <span className="text-white">db-n1-standard-1</span></div>
								</div>
							</div>

							{/* Kubernetes Example */}
							<div>
								<p className="text-sm font-bold text-purple-400 mb-2">Kubernetes</p>
								<div className="bg-slate-950 rounded-lg p-3 font-mono text-xs whitespace-nowrap overflow-x-auto">
									<div className="text-blue-400">apiVersion: <span className="text-white">kubernetes.planton.dev/v1</span></div>
									<div className="text-blue-400">kind: <span className="text-white">PostgresKubernetes</span></div>
									<div className="text-blue-400">metadata:</div>
									<div className="text-blue-400 ml-2">name: <span className="text-white">mydb</span></div>
									<div className="text-blue-400">spec:</div>
									<div className="text-blue-400 ml-2">container:</div>
									<div className="text-blue-400 ml-4">replicas: <span className="text-emerald-400">3</span></div>
									<div className="text-blue-400 ml-4">disk_size: <span className="text-white">&quot;100Gi&quot;</span></div>
								</div>
							</div>
						</div>

					<div className="mt-6 bg-slate-950 rounded-lg p-4 font-mono text-sm text-center">
						<div className="text-slate-400"># Same deployment command for all providers:</div>
						<div className="text-emerald-400 mt-2">planton apply -f postgres.yaml</div>
					</div>

						<div className="mt-6 pt-6 border-t border-slate-700">
							<p className="text-slate-300 text-center mb-6">
								<span className="font-bold text-white">Provider-specific configuration</span> (no artificial abstractions), 
								but <span className="font-bold text-white">consistent structure, workflow, and validation</span> across all clouds.
							</p>
							<div className="text-center">
								<Link href="/docs">
									<Button
										size="lg"
										variant="outline"
										className="rounded-full border border-purple-400/40 text-purple-300 bg-transparent hover:bg-purple-900/20 hover:border-purple-400/60 px-8 py-3 h-auto cursor-pointer"
									>
										<BookOpen className="w-4 h-4 mr-2" />
										Read the Documentation
									</Button>
								</Link>
							</div>
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}

