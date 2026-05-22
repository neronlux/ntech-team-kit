import type { Plugin } from "@opencode-ai/plugin"

export const CIWatcherPlugin: Plugin = async ({ client, $ }) => {
  const POLL_INTERVAL_MS = 60_000
  const MAX_POLLS = 30
  const ENV_ENABLE = "OPENCODE_NTECH_CI_WATCH"

  let timer: ReturnType<typeof setTimeout> | null = null
  let pollCount = 0

  async function poll() {
    try {
      const branch = await $`git branch --show-current`.quiet().text()
      const currentBranch = branch.trim()
      if (!currentBranch) return

      let prJson: string
      try {
        prJson = await $`gh pr view --json number,url,headRefName,state`.quiet().text()
      } catch {
        return
      }

      const pr = JSON.parse(prJson)
      if (!pr || pr.state !== "OPEN") return

      const checksJson = await $`gh pr checks --json name,bucket,state,workflow,link`.quiet().text()
      const checks = JSON.parse(checksJson)

      const pending = checks.filter((c: any) => c.state === "pending" || c.bucket === "pending")
      const failed = checks.filter((c: any) => c.state === "failure" || c.state === "failed" || c.bucket === "fail" || c.bucket === "failed")
      const passed = checks.filter((c: any) => c.state === "success" || c.state === "passed" || c.bucket === "pass" || c.bucket === "passed")

      if (failed.length > 0) {
        const names = failed.map((c: any) => c.name).join(", ")
        await client.app.log({
          body: {
            service: "ntech-ci-watcher",
            level: "error",
            message: `CI FAILED on ${currentBranch}: ${names}`,
            extra: { pr: pr.url, failedChecks: failed },
          },
        })
        stop()
        return
      }

      if (pending.length === 0 && passed.length > 0) {
        await client.app.log({
          body: {
            service: "ntech-ci-watcher",
            level: "info",
            message: `All CI checks passed on ${currentBranch}`,
            extra: { pr: pr.url },
          },
        })
        stop()
        return
      }

      pollCount++
      if (pollCount >= MAX_POLLS) {
        await client.app.log({
          body: {
            service: "ntech-ci-watcher",
            level: "warn",
            message: `CI watch stopped after ${MAX_POLLS} polls for ${currentBranch}`,
            extra: { pr: pr.url, pendingChecks: pending.map((c: any) => c.name) },
          },
        })
        stop()
      }
    } catch (err: any) {
      await client.app.log({
        body: {
          service: "ntech-ci-watcher",
          level: "error",
          message: `CI watch error: ${err.message || String(err)}`,
        },
      })
    }
  }

  function start() {
    if (timer) return
    pollCount = 0
    poll()
    timer = setInterval(poll, POLL_INTERVAL_MS)
  }

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  return {
    event: async ({ event }) => {
      if (event.type === "session.idle") {
        const enabled = process.env[ENV_ENABLE]
        if (!enabled || enabled === "0" || enabled === "false") return

        start()
      }
    },
  }
}
