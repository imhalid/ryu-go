import { Link, createFileRoute } from "@tanstack/react-router"
import { ArrowRightIcon } from "lucide-react"

import { buttonVariants } from "@/components/ui/button"

export const Route = createFileRoute("/")({
  component: Home,
})

// The families of docs/README.md, each by its first page.
const families = [
  { title: "Theming", text: "Light and dark, three size steps, three corners and a brand accent, for a window or a region.", slug: "theming" },
  { title: "Controls", text: "Buttons, fields, selects and cards: one raised pill, one inset well.", slug: "button" },
  { title: "Selection", text: "Checks that draw themselves, dots that open, a pill that slides between views.", slug: "checkbox" },
  { title: "Forms", text: "Text that grows, numbers with a stepper, a calendar, tags as pills.", slug: "textarea" },
  { title: "Layers", text: "Dialogs that drag, sheets, menus and popovers on glass, toasts with a beam of light.", slug: "dialog" },
  { title: "Chat", text: "A conversation that grows message by message, and its composer.", slug: "chat" },
  { title: "Drag and drop", text: "Lists, boards and files, with cards that lift and lean as they go.", slug: "sortable" },
  { title: "Feedback", text: "Progress, meters, spinners and skeletons, in the same wells.", slug: "progress" },
  { title: "Data", text: "A virtualized table: sort, move and resize columns, select, open rows.", slug: "table" },
  { title: "Workspace", text: "Experimental: areas that split, join and take docked tabs, edge to edge, as Blender's.", slug: "workspace" },
]

function Home() {
  return (
    <main>
      <section className="grid gap-12 px-4 py-20 sm:px-10 md:py-24 lg:grid-cols-[1fr_1.1fr] lg:items-center">
        <div>
          <p className="label inline-flex items-center gap-2">
            <span className="size-1.5 rounded-full bg-brand" />
            For MyGo {__MYGO_VERSION__} · macOS, Linux, Windows, browsers
          </p>
          <h1 className="mt-6 text-4xl leading-[1.05] font-semibold tracking-[-0.04em] sm:text-5xl">
            Interfaces you can feel.
            <span className="block text-muted-foreground">Native, in Go alone.</span>
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-8 text-muted-foreground">
            A design system with weight: controls lit from above, wells pressed into the surface, cards that lift, lean and land, checks that draw
            themselves. For MyGo's native UI, drawn on the GPU without a webview or cgo; every preview here is the real component, running live.
          </p>
          <div className="mt-10 flex flex-wrap items-center gap-3">
            <span className="inline-flex h-11 items-center gap-3 rounded-lg border border-dashed bg-card pr-4 pl-4 font-mono text-[13.5px] text-muted-foreground">
              <span className="text-brand-ink select-none">$</span>
              go get · coming soon
            </span>
            <Link to="/docs" className={buttonVariants({ size: "lg", className: "h-11 px-5" })}>
              Read the docs <ArrowRightIcon />
            </Link>
          </div>
        </div>
        <figure className="ui-preview my-0!">
          <div className="ui-preview-titlebar">
            <span className="ui-preview-lights" aria-hidden="true">
              <i className="close" />
              <i className="minimize" />
              <i className="zoom" />
            </span>
            <span className="ui-preview-title">Theming</span>
          </div>
          <iframe className="ui-preview-frame" src="/docs-demo?demo=theming" title="A preview of Ryugo's theming" style={{ height: 460 }} />
        </figure>
      </section>

      <ul className="grid gap-px border-t bg-border sm:grid-cols-2">
        {families.map((family) => (
          <li key={family.title} className="bg-background">
            <Link to="/docs/$" params={{ _splat: family.slug }} className="group flex h-full flex-col px-4 py-8 transition-colors hover:bg-code sm:px-10">
              <h2 className="flex items-center justify-between font-semibold tracking-tight">
                {family.title}
                <ArrowRightIcon className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-foreground" />
              </h2>
              <p className="mt-2 max-w-md text-sm leading-6 text-muted-foreground">{family.text}</p>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  )
}
