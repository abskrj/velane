import React from 'react';
import Image from 'next/image';
import Link from 'next/link';

const featuredIntegrations = [
  { name: "GitHub",     color: "bg-zinc-900",   icon: "GH" },
  { name: "Slack",      color: "bg-[#4A154B]",  icon: "#" },
  { name: "Salesforce", color: "bg-[#0176D3]",  icon: "SF" },
  { name: "Stripe",     color: "bg-[#635BFF]",  icon: "S" },
  { name: "Notion",     color: "bg-zinc-900",   icon: "N" },
  { name: "Linear",     color: "bg-[#5E6AD2]",  icon: "L" },
  { name: "HubSpot",    color: "bg-[#FF7A59]",  icon: "H" },
  { name: "Zendesk",    color: "bg-[#03363D]",  icon: "Z" },
];

export default function Home() {
  return (
    <div className="min-h-screen bg-[#FAFAFA] text-zinc-900 selection:bg-zinc-200">
      <header className="sticky top-0 z-50 border-b border-black/5 bg-[#FAFAFA]/80 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <Image src="/logo.png" alt="Velane Logo" width={24} height={24} className="rounded" />
            <div className="text-xl font-medium tracking-tight">Velane</div>
          </div>
          <div className="flex items-center gap-6">
            <a
              className="text-sm text-zinc-500 transition-colors hover:text-zinc-900 hidden sm:inline-block"
              href="https://github.com/abskrj/velane"
              target="_blank"
              rel="noreferrer"
            >
              Docs
            </a>
            <a
              className="hidden sm:flex items-center gap-2 rounded-full border border-black/10 px-3 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/5 hover:text-zinc-900"
              href="https://github.com/abskrj/velane"
              target="_blank"
              rel="noreferrer"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
              </svg>
              Star
            </a>
            <a
              className="rounded-full bg-zinc-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition-transform hover:scale-105 hover:bg-zinc-800"
              href="https://app.velane.sh"
            >
              Start building
            </a>
          </div>
        </div>
      </header>

      <main>
        {/* Hero Section */}
        <section className="mx-auto w-full max-w-5xl px-6 min-h-[calc(100vh-80px)] flex flex-col justify-center py-20 md:py-32">
          <div className="grid gap-16 md:grid-cols-2 md:items-center">
            <div className="max-w-xl">
              <h1 className="text-5xl font-medium tracking-tight text-zinc-900 leading-[1.1] md:text-6xl">
                Don't write workflows.<br/>Let your agent do it.
              </h1>
              <p className="mt-6 text-lg leading-relaxed text-zinc-600">
                Add Velane to Cursor or Claude via MCP. Your AI handles the rest—creating APIs, pulling integrations, debugging logs, and promoting to production.
              </p>
              
              <div className="mt-10 flex flex-wrap items-center gap-4">
                <a
                  className="rounded-full bg-zinc-900 px-6 py-3 text-sm font-medium text-white shadow-md transition-all hover:scale-105 hover:bg-zinc-800 hover:shadow-lg"
                  href="https://app.velane.sh"
                >
                  Connect your agent
                </a>
                <a
                  className="rounded-full border border-black/10 px-6 py-3 text-sm font-medium text-zinc-600 transition-colors hover:bg-black/5"
                  href="https://github.com/abskrj/velane"
                  target="_blank"
                  rel="noreferrer"
                >
                  View GitHub
                </a>
              </div>
            </div>

            {/* The Velane Circle - Humanized & Crafted */}
            <div className="relative mx-auto w-full max-w-sm">
              <div className="relative aspect-square w-full">
                {/* Background glow */}
                <div className="absolute inset-0 rounded-full bg-zinc-200/50 blur-3xl animate-pulse" />
                
                {/* Main structure */}
                <div className="absolute inset-4 rounded-full border border-black/5 bg-white/50 shadow-xl backdrop-blur-xl transition-transform duration-700 hover:scale-105">
                  
                  {/* Center Node */}
                  <div className="absolute left-1/2 top-1/2 flex h-24 w-24 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full bg-zinc-900 text-white shadow-2xl">
                    <span className="font-medium tracking-tight">Velane</span>
                  </div>

                  {/* Satellite Nodes */}
                  <div className="absolute left-1/2 top-8 -translate-x-1/2 rounded-full border border-black/5 bg-white/90 px-4 py-2 text-xs font-medium text-zinc-600 shadow-sm backdrop-blur-md">
                    iPaaS
                  </div>
                  <div className="absolute bottom-12 left-6 rounded-full border border-black/5 bg-white/90 px-4 py-2 text-xs font-medium text-zinc-600 shadow-sm backdrop-blur-md">
                    Code Runtime
                  </div>
                  <div className="absolute bottom-12 right-6 rounded-full border border-black/5 bg-white/90 px-4 py-2 text-xs font-medium text-zinc-600 shadow-sm backdrop-blur-md">
                    Agent First
                  </div>
                  
                  {/* Decorative connecting lines (SVG) */}
                  <svg className="absolute inset-0 h-full w-full -z-10 opacity-20" viewBox="0 0 100 100">
                    <line x1="50" y1="50" x2="50" y2="20" stroke="black" strokeWidth="0.5" strokeDasharray="2 2" />
                    <line x1="50" y1="50" x2="25" y2="75" stroke="black" strokeWidth="0.5" strokeDasharray="2 2" />
                    <line x1="50" y1="50" x2="75" y2="75" stroke="black" strokeWidth="0.5" strokeDasharray="2 2" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Integrations */}
        <section className="border-y border-black/5 bg-white min-h-screen flex flex-col justify-center py-12">
          <div className="mx-auto w-full max-w-5xl px-6">
            <div className="text-center mb-12">
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-zinc-400 mb-3">Integrations</p>
              <h2 className="text-3xl font-medium tracking-tight text-zinc-900">800+ tools. Zero tokens in your code.</h2>
              <p className="mt-4 text-lg text-zinc-600 max-w-lg mx-auto">
                Your agent discovers connected accounts and calls any API through Velane — no OAuth, no credentials, no glue.
              </p>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 max-w-4xl mx-auto">
              {featuredIntegrations.map((item) => (
                <div
                  key={item.name}
                  className="group flex items-center gap-4 rounded-2xl border border-black/5 bg-white p-5 shadow-sm transition-all hover:border-black/10 hover:shadow-md hover:-translate-y-px"
                >
                  <div
                    className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-sm font-semibold tracking-[-0.3px] text-white ${item.color}`}
                  >
                    {item.icon}
                  </div>
                  <span className="font-medium text-zinc-800 group-hover:text-zinc-950">{item.name}</span>
                </div>
              ))}
            </div>

            <div className="mt-10 text-center">
              <a
                href="https://nango.dev/api-integrations/"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 text-sm font-medium text-zinc-600 hover:text-zinc-900"
              >
                Browse the full catalog on Nango <span aria-hidden="true">→</span>
              </a>
            </div>
          </div>
        </section>

        {/* Storytelling / Features - MCP Focus */}
        <section className="mx-auto w-full max-w-5xl px-6 min-h-screen flex flex-col justify-center py-24">
          <div className="mb-16 max-w-2xl">
            <h2 className="text-3xl font-medium tracking-tight text-zinc-900">
              Your IDE is now your deployment pipeline. <br className="hidden sm:block" />
              Fully automated.
            </h2>
            <p className="mt-4 text-lg text-zinc-600">
              Ship production workflows and integrations for your AI agents without ever leaving your IDE.
            </p>
          </div>

          <div className="grid gap-12 md:grid-cols-3">
            <div className="group">
              <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-full bg-zinc-100 text-zinc-900 transition-colors group-hover:bg-zinc-900 group-hover:text-white">
                1
              </div>
              <h3 className="text-xl font-medium text-zinc-900">Create & Connect</h3>
              <p className="mt-3 leading-relaxed text-zinc-600">
                Your agent writes the Bun or Python code and pulls external API documentation for 800+ integrations directly into its context via MCP.
              </p>
            </div>
            <div className="group">
              <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-full bg-zinc-100 text-zinc-900 transition-colors group-hover:bg-zinc-900 group-hover:text-white">
                2
              </div>
              <h3 className="text-xl font-medium text-zinc-900">Run & Debug</h3>
              <p className="mt-3 leading-relaxed text-zinc-600">
                Cursor or Claude can invoke snippets, read live execution logs, and inspect past invocations to fix errors autonomously without leaving the chat.
              </p>
            </div>
            <div className="group">
              <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-full bg-zinc-100 text-zinc-900 transition-colors group-hover:bg-zinc-900 group-hover:text-white">
                3
              </div>
              <h3 className="text-xl font-medium text-zinc-900">Promote</h3>
              <p className="mt-3 leading-relaxed text-zinc-600">
                Once testing passes, your agent manages the release—promoting versions across dev, staging, and production environments instantly.
              </p>
            </div>
          </div>
        </section>

        {/* Technical Proof */}
        <section className="bg-zinc-900 min-h-screen flex flex-col justify-center py-24 text-white selection:bg-zinc-700">
          <div className="mx-auto grid w-full max-w-5xl gap-16 px-6 md:grid-cols-2 md:items-center">
            <div>
              <h2 className="text-3xl font-medium tracking-tight">
                Just add to your MCP config
              </h2>
              <p className="mt-4 text-zinc-400 leading-relaxed">
                Connect Velane to your preferred AI assistant. It immediately discovers your tenant's connected integrations, snippets, and execution environment.
              </p>
              <div className="mt-8 flex flex-col gap-4 border-l border-zinc-700 pl-6">
                <div>
                  <span className="text-sm font-medium text-zinc-300">Zero Glue Code</span>
                  <p className="text-sm text-zinc-500">The agent calls integrations without dealing with OAuth flows.</p>
                </div>
                <div>
                  <span className="text-sm font-medium text-zinc-300">Secure Sandboxes</span>
                  <p className="text-sm text-zinc-500">All code executes remotely in isolated runtimes.</p>
                </div>
              </div>
            </div>
            
            <div className="rounded-2xl border border-white/10 bg-black/50 p-2 shadow-2xl">
              <div className="flex items-center gap-2 border-b border-white/10 bg-white/5 px-4 py-3">
                <div className="h-3 w-3 rounded-full bg-red-500/80" />
                <div className="h-3 w-3 rounded-full bg-yellow-500/80" />
                <div className="h-3 w-3 rounded-full bg-green-500/80" />
                <span className="ml-2 text-xs font-mono text-zinc-500">mcp.json</span>
              </div>
              <pre className="overflow-x-auto p-6 text-sm leading-relaxed text-zinc-300 font-mono">
                {'{'}{'\n'}
                {'  '}<span className="text-pink-400">"mcpServers"</span>: {'{'}{'\n'}
                {'    '}<span className="text-blue-400">"velane"</span>: {'{'}{'\n'}
                {'      '}<span className="text-pink-400">"url"</span>: <span className="text-emerald-300">"http://localhost:8090/mcp"</span>,{'\n'}
                {'      '}<span className="text-pink-400">"headers"</span>: {'{'}{'\n'}
                {'        '}<span className="text-blue-400">"Authorization"</span>: <span className="text-emerald-300">"Bearer vl_xxxx"</span>{'\n'}
                {'      '}{'}'}{'\n'}
                {'    '}{'}'}{'\n'}
                {'  '}{'}'}{'\n'}
                {'}'}
              </pre>
            </div>
          </div>
        </section>

        {/* Final CTA */}
        <section className="mx-auto w-full max-w-5xl px-6 min-h-screen flex flex-col justify-center py-32 text-center">
          <h2 className="text-4xl font-medium tracking-tight text-zinc-900">
            Ready to bring your agents to life?
          </h2>
          <p className="mx-auto mt-6 max-w-xl text-lg text-zinc-600">
            Join the teams building the next generation of reliable, agent-driven applications on Velane.
          </p>
          <div className="mt-10 flex justify-center gap-4">
            <a
              className="rounded-full bg-zinc-900 px-8 py-4 text-sm font-medium text-white shadow-lg transition-transform hover:scale-105 hover:bg-zinc-800"
              href="https://app.velane.sh"
            >
              Start building for free
            </a>
          </div>
        </section>
      </main>

      <footer className="relative border-t border-black/5 bg-white py-16 overflow-hidden">
        {/* Large background "Velane" watermark */}
        <div className="absolute bottom-[-20px] left-0 right-0 -z-10 pointer-events-none select-none text-center">
          <span className="block text-[120px] md:text-[180px] lg:text-[220px] xl:text-[260px] font-bold tracking-[-0.065em] text-zinc-900 opacity-[0.035] leading-[0.75]">
            Velane
          </span>
        </div>

        <div className="mx-auto grid max-w-5xl grid-cols-2 gap-8 px-6 md:grid-cols-3 lg:gap-12 relative z-10">
          <div className="col-span-2 flex flex-col items-start gap-6 lg:col-span-1">
            <div className="text-xl font-medium tracking-tight text-zinc-900">Velane</div>
            <p className="text-sm text-zinc-500">
              The agent-first code runtime. Build, connect, and ship APIs on autopilot.
            </p>
          </div>
          
          <div className="flex flex-col gap-4">
            <span className="text-sm font-semibold text-zinc-900">Product</span>
            <a href="https://app.velane.sh" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Platform</a>
            <a href="https://nango.dev/api-integrations/" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Integrations</a>
          </div>

          <div className="flex flex-col gap-4">
            <span className="text-sm font-semibold text-zinc-900">Resources</span>
            <a href="https://docs.velane.sh" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Documentation</a>
            <a href="https://github.com/abskrj/velane/issues" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Issue Tracker</a>
            <a href="https://github.com/abskrj/velane" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">GitHub</a>
          </div>
        </div>

        <div className="mx-auto mt-16 flex max-w-5xl flex-col items-center justify-between gap-4 border-t border-black/5 px-6 pt-8 sm:flex-row relative z-10">
          <div className="text-sm text-zinc-500">
            © {new Date().getFullYear()} Velane. All rights reserved.
          </div>
          <div className="flex gap-6">
            <Link href="/privacy" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Privacy Policy</Link>
            <Link href="/terms" className="text-sm text-zinc-500 transition-colors hover:text-zinc-900">Terms of Service</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
