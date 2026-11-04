// Writes the site, prerendered by `vite build` into dist/client, as
// Cloudflare's Build Output for `cf deploy --prebuilt`: static assets only,
// with no Worker code, served at ryugo.halid.dev. cf's own build runs the
// prerender inside the Workers runtime, where the docs can't be read from
// disk, so the site is built by Vite here and only uploaded by cf.
import path from "node:path"
import { cleanBuildOutputDir, writeAssets, writeRootConfig, writeWorkerConfig } from "@cloudflare/build-output-utils"

const root = import.meta.dirname

await cleanBuildOutputDir(root)
await writeAssets({ root, sourceDirectory: path.join(root, "dist", "client") })
await writeWorkerConfig({
  root,
  config: {
    name: "ryugo",
    compatibilityDate: "2026-09-01",
    assets: {
      // docs/button/index.html is served at /docs/button.
      htmlHandling: "drop-trailing-slash",
      // dist/client/404.html, with a 404 status.
      notFoundHandling: "404-page",
    },
    domains: ["ryugo.halid.dev"],
  },
})
await writeRootConfig(root, undefined, { isPreview: false, mode: "production" })
