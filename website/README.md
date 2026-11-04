# Mizan website

The home page and, at `/docs`, the guides of [`../docs`](../docs), with a
live preview of every component, built with TanStack Start and shadcn/ui
after [MyGo's website](https://github.com/egoist/mygo/tree/main/website).
Every page is prerendered at build time, and Cloudflare Workers serves the
result as static assets, with no Worker code.

```sh
bun install                    # in website/
bun run dev                    # http://localhost:3000
bun run build                  # dist/client
bun run preview                # dist/client, served by wrangler as in production
bun run deploy                 # build, then wrangler deploy
```

Go must be installed: the build compiles the previews to WebAssembly.

Pages of the docs are rendered from the markdown when the site is built:
the order of the sidebar comes from the lists of `docs/README.md`, and links
between pages, such as `button.md#color-and-fill`, become links of the
site. Lists of pages, `- [Title](page.md): what it covers`, become cards.

A comment `<!-- preview button 190 -->` in a page becomes a window of
macOS, as high as it says, running that demo of `website/demo` (a Go app of
native UI compiled to WebAssembly); a button of its title bar opens it in
a tab of its own. A page shows its own demo (`button` on
`button.md`) and others named after it (`button-colors`):
`go test ./website/demo` checks that, and builds every demo.

```sh
go run ./website/demo                   # every demo, in a sidebar, on a desktop
go run ./website/demo button            # one
DEMO_SHOT=all DEMO_DIR=/tmp/shots go test -run TestShots ./website/demo   # PNGs of them
```
