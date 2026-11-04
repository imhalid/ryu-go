import { Link } from "@tanstack/react-router"

import { site } from "@/lib/site"

export function SiteFooter() {
  return (
    <footer className="frame label flex flex-wrap items-center justify-between gap-4 border-t px-4 py-6 sm:px-6">
      <span>
        Ryugo · MyGo {__MYGO_VERSION__} ·{" "}
        <a href={`${site.repo}/blob/main/LICENSE`} className="hover:text-foreground">
          MIT License
        </a>
      </span>
      <nav className="flex flex-wrap gap-x-6 gap-y-2">
        <Link to="/docs" className="hover:text-foreground">
          Docs
        </Link>
        <a href="https://mygo.egoist.dev" className="hover:text-foreground">
          MyGo
        </a>
        <a href={site.repo} className="hover:text-foreground">
          GitHub
        </a>
      </nav>
    </footer>
  )
}
