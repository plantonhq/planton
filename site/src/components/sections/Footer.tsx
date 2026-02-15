import React from "react";
import Image from "next/image";
import Link from "next/link";
import { Github, Scale, MessageSquare, AlertCircle } from "lucide-react";

export default function Footer() {
  const links = [
    { name: "Documentation", href: "/docs", external: false },
    { name: "Getting Started", href: "/docs/getting-started", external: false },
    { name: "Components", href: "/docs/catalog", external: false },
    { name: "GitHub", href: "https://github.com/plantonhq/openmcf", external: true },
  ];

  return (
    <footer className="bg-slate-900 border-t border-slate-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid md:grid-cols-4 gap-8">
          <div className="md:col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <Image src="/icon.png" alt="OpenMCF mark" width={36} height={36} className="h-9 w-auto object-contain" />
              <Image src="/text-logo.svg" alt="OpenMCF" width={155} height={32} className="h-8 w-auto object-contain" />
            </div>
            <p className="text-slate-400 max-w-md mb-6">
              Kubernetes‑style manifests for multi‑cloud infrastructure. Define once. Deploy anywhere.
            </p>
            <div className="flex items-center gap-2 text-sm text-slate-500">
              <Scale className="w-4 h-4" />
              <span>Apache-2.0 License</span>
            </div>
          </div>
          <div>
            <h3 className="font-semibold text-white mb-4">Resources</h3>
            <div className="space-y-3">
              {links.map((link) => (
                link.external ? (
                  <a
                    key={link.name}
                    href={link.href}
                    className="block text-slate-400 hover:text-white transition-colors"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {link.name}
                  </a>
                ) : (
                  <Link
                    key={link.name}
                    href={link.href}
                    className="block text-slate-400 hover:text-white transition-colors"
                  >
                    {link.name}
                  </Link>
                )
              ))}
            </div>
          </div>
          <div>
            <h3 className="font-semibold text-white mb-4">Community</h3>
            <div className="space-y-3">
              <a
                href="https://github.com/plantonhq/openmcf"
                target="_blank"
                rel="noopener noreferrer"
                className="block text-slate-400 hover:text-white transition-colors"
              >
                <div className="flex items-center gap-2">
                  <Github className="w-4 h-4" />
                  GitHub Repository
                </div>
              </a>
              <a 
                href="https://github.com/plantonhq/openmcf/discussions"
                target="_blank"
                rel="noopener noreferrer"
                className="block text-slate-400 hover:text-white transition-colors"
              >
                <div className="flex items-center gap-2">
                  <MessageSquare className="w-4 h-4" />
                  Discussions
                </div>
              </a>
              <a 
                href="https://github.com/plantonhq/openmcf/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="block text-slate-400 hover:text-white transition-colors"
              >
                <div className="flex items-center gap-2">
                  <AlertCircle className="w-4 h-4" />
                  Issues
                </div>
              </a>
            </div>
          </div>
        </div>
        <div className="border-t border-slate-800 mt-12 pt-8 text-center">
          <p className="text-slate-500">© 2025 OpenMCF. Open‑source under Apache-2.0 License.</p>
        </div>
      </div>
    </footer>
  );
}


