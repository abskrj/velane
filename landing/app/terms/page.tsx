import React from 'react';
import Link from 'next/link';
import Image from 'next/image';

export default function TermsOfService() {
  const lastUpdated = "June 9, 2026";

  return (
    <div className="min-h-screen bg-[#FAFAFA] text-zinc-900">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-black/5 bg-[#FAFAFA]/80 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-4">
          <Link href="/" className="flex items-center gap-2">
            <Image src="/logo.png" alt="Velane Logo" width={24} height={24} className="rounded" />
            <span className="text-xl font-medium tracking-tight">Velane</span>
          </Link>
          <div className="flex items-center gap-6">
            <Link
              href="https://github.com/abskrj/velane"
              target="_blank"
              rel="noreferrer"
              className="text-sm text-zinc-500 transition-colors hover:text-zinc-900 hidden sm:inline-block"
            >
              Docs
            </Link>
            <Link
              href="https://app.velane.sh"
              className="rounded-full bg-zinc-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition-transform hover:scale-105 hover:bg-zinc-800"
            >
              Start building
            </Link>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-[680px] px-6 py-16">
        <div className="mb-14">
          <h1 className="text-4xl font-medium tracking-[-0.5px]">Terms of Service</h1>
          <p className="mt-3 text-sm text-zinc-500">Last updated: {lastUpdated}</p>
        </div>

        <div className="space-y-10 text-[15px] leading-[1.75] text-zinc-700">
          <div className="space-y-5 text-base text-zinc-800">
            <p>
              These Terms of Service ("Terms") govern your access to and use of the Velane website, the hosted platform at app.velane.sh, the Velane API, and related services (collectively, the "Service") provided by Velane ("we", "us", or "our").
            </p>
            <p>
              By accessing or using the Service, you agree to be bound by these Terms. If you do not agree, you may not use the Service.
            </p>
          </div>

          {/* Section 1 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">1. Description of Service</h2>
            <p>
              Velane is an AI agent code runtime that allows you to write, deploy, and invoke Bun and Python code snippets as production APIs, with built-in support for 800+ integrations, versioning, environments, and first-class MCP integration for coding agents.
            </p>
          </section>

          {/* Section 2 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">2. Accounts and Tenants</h2>
            <ul className="list-disc pl-6 space-y-2.5">
              <li>You must create an account and tenant to use most features of the Service.</li>
              <li>You are responsible for maintaining the confidentiality of your credentials and for all activity that occurs under your account.</li>
              <li>You may invite team members to your tenant. You are responsible for their actions within the tenant.</li>
            </ul>
          </section>

          {/* Section 3 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">3. Acceptable Use</h2>
            <p className="mb-3">You agree not to:</p>
            <ul className="list-disc pl-6 space-y-2">
              <li>Use the Service for any illegal purpose or in violation of any laws.</li>
              <li>Attempt to gain unauthorized access to any part of the Service or related systems.</li>
              <li>Interfere with or disrupt the integrity or performance of the Service.</li>
              <li>Upload or transmit malicious code, including viruses, worms, or trojans.</li>
              <li>Use the Service to process or store data in a manner that violates the rights of others.</li>
            </ul>
          </section>

          {/* Section 4 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">4. Your Code and Data</h2>
            <div className="space-y-4">
              <p>
                You retain all ownership rights to the code snippets, configurations, and data you upload or create using the Service ("Your Content").
              </p>
              <p>
                You grant us a limited license to host, run, and display Your Content solely as necessary to provide the Service to you.
              </p>
              <p>
                You are solely responsible for the legality, accuracy, and appropriateness of Your Content.
              </p>
            </div>
          </section>

          {/* Section 5 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">5. Integrations and Third-Party Services</h2>
            <div className="space-y-4">
              <p>
                The Service allows you to connect to third-party APIs and services. You are responsible for complying with the terms of service and privacy policies of any third-party services you connect.
              </p>
              <p>
                We are not responsible for the availability, accuracy, or practices of third-party services.
              </p>
            </div>
          </section>

          {/* Section 6 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">6. Intellectual Property</h2>
            <div className="space-y-4">
              <p>
                The Service, including all software, designs, text, graphics, and trademarks (excluding Your Content), is owned by us or our licensors. You may not copy, modify, distribute, or create derivative works from the Service without our prior written permission, except as expressly allowed by these Terms or applicable open-source licenses.
              </p>
              <p>
                The core Velane software is available under the AGPL-3.0 license (and a commercial license option). See the repository for details.
              </p>
            </div>
          </section>

          {/* Section 7 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">7. Fees and Payment</h2>
            <p>
              Certain features of the Service may be offered for a fee. If you purchase a paid plan, you agree to pay all applicable fees. Fees are non-refundable except as required by law or as explicitly stated by us.
            </p>
          </section>

          {/* Section 8 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">8. Disclaimer of Warranties</h2>
            <p className="mb-3">
              THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, WHETHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
            </p>
            <p>
              We do not warrant that the Service will be uninterrupted, error-free, or completely secure.
            </p>
          </section>

          {/* Section 9 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">9. Limitation of Liability</h2>
            <p>
              TO THE MAXIMUM EXTENT PERMITTED BY LAW, WE SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS OR REVENUES, WHETHER INCURRED DIRECTLY OR INDIRECTLY, OR ANY LOSS OF DATA, USE, GOODWILL, OR OTHER INTANGIBLE LOSSES, RESULTING FROM YOUR USE OF THE SERVICE.
            </p>
          </section>

          {/* Section 10 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">10. Termination</h2>
            <div className="space-y-4">
              <p>
                We may suspend or terminate your access to the Service at any time, with or without cause, and with or without notice. You may stop using the Service at any time by deleting your account.
              </p>
              <p>
                Upon termination, your right to use the Service will immediately cease. Provisions that by their nature should survive termination shall survive.
              </p>
            </div>
          </section>

          {/* Section 11 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">11. Governing Law</h2>
            <p>
              These Terms shall be governed by and construed in accordance with the laws of the jurisdiction in which Velane operates, without regard to its conflict of law provisions.
            </p>
          </section>

          {/* Section 12 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">12. Changes to These Terms</h2>
            <p>
              We may modify these Terms at any time. We will post the revised Terms on this page and update the "Last updated" date. Continued use of the Service after changes constitutes acceptance of the new Terms.
            </p>
          </section>

          {/* Section 13 */}
          <section>
            <h2 className="mb-4 text-xl font-medium tracking-tight text-zinc-900">13. Contact</h2>
            <p>
              For questions about these Terms, please contact us at:
            </p>
            <p className="mt-2">
              <a href="mailto:abhi@velane.sh" className="font-medium text-zinc-900 underline hover:no-underline">abhi@velane.sh</a>
            </p>
          </section>
        </div>
      </main>

      {/* Minimal footer */}
      <footer className="border-t border-black/5 py-10">
        <div className="mx-auto max-w-5xl px-6 text-center text-sm text-zinc-500">
          © {new Date().getFullYear()} Velane. All rights reserved.
        </div>
      </footer>
    </div>
  );
}
