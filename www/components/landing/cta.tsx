import { ArrowRight, Github, BookOpen } from 'lucide-react'
import Link from 'next/link'

export function CTA() {
  return (
    <section className="relative py-24 sm:py-28">
      <div className="mx-auto max-w-6xl px-6">
        <div className="relative overflow-hidden rounded-3xl border border-fd-border/70 bg-fd-card/65 p-10 sm:p-12 lg:p-16">
          <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(70%_70%_at_85%_10%,rgba(56,189,248,0.16),transparent_78%)]" />
          <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(55%_55%_at_15%_90%,rgba(20,184,166,0.14),transparent_75%)]" />

          <div className="relative text-center">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-fd-muted-foreground">
              Get started
            </p>
            <h2 className="mx-auto mt-4 max-w-2xl font-[family-name:var(--font-home-display)] text-3xl font-semibold tracking-[-0.03em] text-fd-foreground sm:text-4xl lg:text-5xl">
              Ready to streamline your MCP operations?
            </h2>
            <p className="mx-auto mt-5 max-w-xl text-sm leading-relaxed text-fd-muted-foreground sm:text-base">
              Download the desktop app or explore the documentation to get started with elastic runtime management.
            </p>

            <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Link
                href="https://github.com/wibus-wee/mcpv/releases"
                target="_blank"
                rel="noopener noreferrer"
                className="group inline-flex h-12 items-center justify-center gap-2 rounded-xl bg-fd-foreground px-7 text-sm font-medium text-fd-background transition-colors hover:bg-fd-foreground/90"
              >
                <Github className="h-4 w-4" />
                Download from GitHub
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
              </Link>
              <Link
                href="/docs"
                className="inline-flex h-12 items-center justify-center gap-2 rounded-xl border border-fd-border bg-fd-background/80 px-7 text-sm font-medium text-fd-muted-foreground transition-colors hover:text-fd-foreground"
              >
                <BookOpen className="h-4 w-4" />
                Read Documentation
              </Link>
            </div>

            <p className="mt-8 text-xs text-fd-muted-foreground">
              Currently available for macOS · Linux and Windows support coming soon
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}
