"use client";
import React from "react";
import Image from "next/image";
import Link from "next/link";
import Hero from "@/components/sections/Hero";
import ProblemStatement from "@/components/sections/ProblemStatement";
import FeatureCards from "@/components/sections/FeatureCards";
import Stats from "@/components/sections/Stats";
import HowItWorks from "@/components/sections/HowItWorks";
import Quickstart from "@/components/sections/Quickstart";
import ExampleGallery from "@/components/sections/ExampleGallery";
import CLIReference from "@/components/sections/CLIReference";
import CICDSection from "@/components/sections/CICDSection";
import CompareSection from "@/components/sections/CompareSection";
import FAQ from "@/components/sections/FAQ";
import Footer from "@/components/sections/Footer";
import { GitHubStarBadge } from "@/components/ui/GitHubStarBadge";

export default function HomePage() {

  return (
    <div className="min-h-screen bg-slate-950 text-white overflow-x-hidden">
      <nav className="fixed top-0 w-full bg-slate-950/95 backdrop-blur-sm border-b border-slate-800 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-2">
              <Image src="/icon.png" alt="Planton logo" width={36} height={36} className="h-9 w-auto object-contain" priority />
              <Image src="/text-logo.svg" alt="Planton" width={155} height={32} className="h-8 w-auto object-contain hidden sm:block" />
            </div>
            <div className="flex items-center gap-4 sm:gap-6">
                <Link
                  href="/docs"
                  className="text-sm font-medium text-slate-300 hover:text-white transition-colors"
                >
                  Docs
                </Link>
              <GitHubStarBadge repo="plantonhq/planton" />
            </div>
          </div>
        </div>
      </nav>

      <main className="pt-16">
        <section id="hero">
          <Hero />
        </section>
        <section id="problem" className="py-24 bg-slate-900/30">
          <ProblemStatement />
        </section>
        <section id="why" className="py-24">
          <FeatureCards />
        </section>
        <section id="stats" className="py-24 bg-slate-900/30">
          <Stats />
        </section>
        <section id="how" className="py-24">
          <HowItWorks />
        </section>
        <section id="quickstart" className="py-24">
          <Quickstart />
        </section>
        <section id="examples" className="py-24">
          <ExampleGallery />
        </section>
        <section id="cli" className="py-24">
          <CLIReference />
        </section>
        <section id="cicd" className="py-24">
          <CICDSection />
        </section>
        <section id="compare" className="py-24">
          <CompareSection />
        </section>
        <section id="faq" className="py-24">
          <FAQ />
        </section>
      </main>
      <Footer />
    </div>
  );
}


