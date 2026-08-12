import {
  argument,
  Container,
  dag,
  Directory,
  File,
  func,
  object,
  Secret,
} from "@dagger.io/dagger"

const GO_IMAGE =
  "golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd"
const JQ_IMAGE =
  "ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486"
const PLAYWRIGHT_VERSION = "v0.6100.0"

const SOURCE_EXCLUDES = [
  ".git",
  ".git/**",
  ".dagger/node_modules",
  ".dagger/node_modules/**",
  ".dagger-output",
  ".dagger-output/**",
  ".dagger-input",
  ".dagger-input/**",
]

const CACHE_NAMESPACES = [
  "pr",
  "trusted",
  "branch-hosted",
  "benchmark-hosted",
  "local",
]

@object()
export class GoshtosoAppShells {
  /** Run templ generation drift, tests, vet, and build. */
  @func({ cache: "never" })
  ci(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cacheNamespace = "local",
    runNonce = "local",
  ): Promise<string> {
    this.validateNonce(runNonce)
    return this.goContainer(source, cacheNamespace)
      .withExec(["bash", "scripts/check-dagger-contract.sh"])
      .withEnvVariable("CI_RUN_NONCE", runNonce)
      .withExec(["bash", "scripts/dagger/ci.sh"])
      .stdout()
  }

  /** Generate templates and run the component documentation browser proof. */
  @func({ cache: "never" })
  browser(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cacheNamespace = "local",
    runNonce = "local",
  ): Promise<string> {
    this.validateNonce(runNonce)
    return this.browserContainer(source, cacheNamespace)
      .withExec(["bash", "scripts/check-dagger-contract.sh"])
      .withExec([
        "go",
        "install",
        "github.com/a-h/templ/cmd/templ@v0.3.1020",
      ])
      .withExec(["/go/bin/templ", "generate"])
      .withExec([
        "go",
        "run",
        `github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_VERSION}`,
        "install-deps",
        "chromium",
      ])
      .withExec([
        "go",
        "run",
        `github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_VERSION}`,
        "install",
        "chromium",
      ])
      .withEnvVariable("CI_RUN_NONCE", runNonce)
      .withEnvVariable("COMPONENTDOCSHELL_E2E", "1")
      .withExec(["go", "test", "./example/e2e", "-count=1", "-v"])
      .stdout()
  }

  /** Run the exact cache-state benchmark workload. */
  @func({ cache: "never" })
  benchmark(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cacheNamespace = "local",
    runNonce = "local",
  ): Promise<string> {
    this.validateNonce(runNonce)
    return this.goContainer(source, cacheNamespace)
      .withExec(["bash", "scripts/check-dagger-contract.sh"])
      .withEnvVariable("CI_RUN_NONCE", runNonce)
      .withExec(["bash", "scripts/ci-hostinger-benchmark.sh"])
      .stdout()
  }

  /** Verify an immutable assets release and return updated allowlisted files. */
  @func({ cache: "never" })
  assetsUpdate(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    payload: File,
    githubToken: Secret,
    cacheNamespace = "trusted",
    runNonce = "local",
  ): Directory {
    this.validateNonce(runNonce)
    return this.goContainer(source, cacheNamespace)
      .withFile("/run/assets-payload.json", payload)
      .withExec(["bash", "scripts/check-dagger-contract.sh"])
      .withExec(["bash", "scripts/dagger/assets-validate_test.sh"])
      .withExec(["bash", "scripts/dagger/assets-validate.sh"])
      .withSecretVariable("GH_TOKEN", githubToken)
      .withEnvVariable("CI_RUN_NONCE", runNonce)
      .withExec(["bash", "scripts/dagger/assets-update.sh"])
      .directory("/out")
  }

  private goContainer(
    source: Directory,
    cacheNamespace: string,
  ): Container {
    if (!CACHE_NAMESPACES.includes(cacheNamespace)) {
      throw new Error(`unsupported cache namespace: ${cacheNamespace}`)
    }
    const jq = dag.container().from(JQ_IMAGE).file("/jq")

    return dag
      .container()
      .from(GO_IMAGE)
      .withFile("/usr/local/bin/jq", jq, { permissions: 0o755 })
      .withEnvVariable("GOCACHE", "/root/.cache/go-build")
      .withEnvVariable("GOMODCACHE", "/go/pkg/mod")
      .withEnvVariable("GOWORK", "off")
      .withDirectory("/work", source)
      .withWorkdir("/work")
      .withMountedCache(
        "/go/pkg/mod",
        dag.cacheVolume(`araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-gomod`),
      )
      .withMountedCache(
        "/root/.cache/go-build",
        dag.cacheVolume(`araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-gobuild`),
      )
  }

  private browserContainer(source: Directory, cacheNamespace: string): Container {
    return this.goContainer(source, cacheNamespace)
      .withEnvVariable("PLAYWRIGHT_BROWSERS_PATH", "/ms-playwright")
      .withMountedCache(
        "/ms-playwright",
        dag.cacheVolume(
          `araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-playwright-${PLAYWRIGHT_VERSION}`,
        ),
      )
  }

  private validateNonce(runNonce: string) {
    if (!/^[A-Za-z0-9._:-]{1,160}$/.test(runNonce)) {
      throw new Error("run nonce must be 1-160 safe characters")
    }
  }
}
