import { Link, useLocation } from "@tanstack/react-router"
import { MenuIcon } from "lucide-react"

import { GitHubIcon } from "@/components/icons"
import { Wordmark } from "@/components/logo"
import { Search } from "@/components/search"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button, buttonVariants } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { site } from "@/lib/site"

const navLink = "text-muted-foreground transition-colors hover:text-foreground data-current:text-foreground"
const mobileNavLink = "min-h-11 px-3 data-current:bg-accent aria-[current=page]:bg-accent"

export function SiteHeader() {
  const page = useLocation({ select: (l) => (l.pathname.startsWith("/docs") ? "docs" : undefined) })
  return (
    <header className="frame sticky top-0 z-40 flex h-14 items-center gap-6 border-b bg-background/80 px-4 backdrop-blur-xl sm:px-6">
      <Link to="/">
        <Wordmark />
      </Link>
      <nav className="hidden items-center gap-5 text-sm sm:flex">
        <Link to="/docs" className={navLink} activeOptions={{ exact: true }} data-current={page === "docs" || undefined}>
          Docs
        </Link>
        <a href="https://mygo.egoist.dev/docs/ui" className={navLink}>
          MyGo
        </a>
      </nav>
      <div className="ml-auto flex items-center gap-1">
        <Search />
        <a href={site.repo} aria-label="GitHub" className={buttonVariants({ variant: "ghost", size: "icon" })}>
          <GitHubIcon />
        </a>
        <ThemeToggle />
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger render={<Button variant="ghost" size="icon" className="sm:hidden" aria-label="Navigation menu" />}>
            <MenuIcon />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-44 sm:hidden">
            <DropdownMenuItem
              className={mobileNavLink}
              render={<Link to="/docs" activeOptions={{ exact: true }} data-current={page === "docs" || undefined} />}
            >
              Docs
            </DropdownMenuItem>
            <DropdownMenuItem className={mobileNavLink} render={<a href="https://mygo.egoist.dev/docs/ui" />}>
              MyGo
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
