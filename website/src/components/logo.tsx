import type * as React from "react"

import { cn } from "@/lib/utils"

/** The name, after a dot of the accent: a balance's pivot. */
export function Wordmark({ className, ...props }: React.ComponentProps<"span">) {
  return (
    <span className={cn("inline-flex items-center gap-2 font-semibold tracking-tight", className)} {...props}>
      <span aria-hidden="true" className="size-2.5 rounded-full bg-brand shadow-[0_0_0_3px_color-mix(in_oklch,var(--brand)_22%,transparent)]" />
      Mizan
    </span>
  )
}
