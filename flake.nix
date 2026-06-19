{
  description = "Reproducible Go dev environment with build and test support";

  # ============================================================
  # INPUTS
  # ============================================================

  inputs = rec {
    nixpkgs-stable.url = "github:NixOS/nixpkgs/nixos-26.05";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixpkgs = nixpkgs-unstable;
    flake-utils.url = "github:numtide/flake-utils";
  };

  # ============================================================
  # OUTPUTS
  # ============================================================

  outputs =
    inputs@{ self
    , nixpkgs
    , nixpkgs-stable
    , nixpkgs-unstable
    , flake-utils
    , ...
    }:
    let
      # ==========================================================
      # PROJECT CONFIGURATION — edit this section for your project
      # ==========================================================

      # Package metadata
      pname = "peasant-labs-schema";

      # Version is derived from the git short rev: this is a -dev/source build,
      # NOT a tagged release (tagged release binaries come from goreleaser with
      # -X ...version=<tag>). self.shortRev exists only for a clean tree; fall
      # back to dirtyShortRev, then a literal, so `nix build` works on a dirty
      # working tree too. This string is injected into internal/defaults.version
      # via ldflags below (shortRev + ldflags).
      version = self.shortRev or self.dirtyShortRev or "dev";

      # Go package attribute (e.g., go, go_1_26)
      # Set to null to use the default Go version from nixpkgs
      goAttr = "go_1_26";

      # Vendor hash for buildGoModule. Recompute when go.mod/go.sum changes:
      # set this to nixpkgs.lib.fakeHash, run `nix build`, copy the reported `got:`
      # hash back here.
      vendorHash = null;

      # Extra CLI tools available in the dev shell
      devTools = pkgs: with pkgs; [
        gopls # LSP
        gotools # goimports, godoc, etc.
        go-tools # staticcheck
        delve # debugger
        ast-grep # structural code search and lint
        golangci-lint # linter suite
        sqlite # CLI for inspecting analytics store
        goreleaser # validate .goreleaser.yml (`goreleaser check`) + local --snapshot builds
        actionlint # lint GitHub Actions workflow YAML (.github/workflows/*.yml)
        nodejs_24 # Node.js runtime (npm ci, npm run build)
        pnpm # pnpm workspace builds for first-party file: deps (@peasant-labs/transcript-browser + analytics)
        typescript
        typescript-language-server
      ];

      # Native build dependencies (C libraries, system packages)
      # gcc is required for tree-sitter CGo bindings (github.com/tree-sitter/go-tree-sitter)
      nativeBuildDeps = pkgs: with pkgs; [
        gcc
        # pkg-config
        # openssl
        # sqlite
      ];

      # Extra check commands run during `nix build` after go test
      extraCheckPhase = ''
        # go vet ./...
        # staticcheck ./...
      '';

      # Files to install alongside the binary (relative to src)
      extraInstallPhase = ''
        # mkdir -p $out/share/policies
        # cp authz/policies/*.rego $out/share/policies/
      '';

      # ==========================================================
      # IMPLEMENTATION — you shouldn't need to edit below here
      # ==========================================================

      mkOutputs = nixpkgs-channel:
        flake-utils.lib.eachDefaultSystem (system:
          let
            pkgs = import nixpkgs-channel {
              inherit system;
              config.allowUnfree = true;
            };

            goPackage =
              if goAttr != null
              then pkgs.${goAttr}
              else pkgs.go;

            # ----------------------------------------------------------
            # Build
            # ----------------------------------------------------------

            package = pkgs.buildGoModule {
              inherit pname version;
              src = ./.;
              inherit vendorHash;

              # Build ONLY the peasant CLI. The default `./...` enumeration would
              # try to build ./pkg/schema, which is a SEPARATE nested Go module
              # (its own go.mod, pulled into the main module via a local replace)
              # and fails as "main module does not contain package …/pkg/schema".
              # cmd/schema-gen and cmd/release-guard are dev/CI tools, not shipped.
              subPackages = [ "cmd/peasant" ];

              # Build the same way the release binaries do: CGO disabled, so the
              # output is a static, portable binary. tree-sitter (the cgo-only
              # Maximum-redaction backend) is therefore NOT linked; `peasant
              # … --redaction maximum` returns the actionable hard error from
              # pkg/redact (redact.MaximumAvailable == false), consistent with
              # the goreleaser/distro builds. -race in checkPhase is removed
              # below for the same reason (the race detector requires cgo).
              env.CGO_ENABLED = 0;

              # Inject the version into internal/defaults.version (matches the
              # release ldflags) and strip debug info for a smaller binary.
              ldflags = [
                "-s"
                "-w"
                "-X github.com/peasant-labs/peasant/internal/defaults.version=${version}"
              ];

              # The binary embeds web/out via `//go:embed all:web/out`
              # (embed.go). A committed placeholder (web/out/.gitkeep) already
              # satisfies it, but `src = ./.` would also capture any stale local
              # `make web` output, making the build non-deterministic. Reset
              # web/out to a deterministic stub so the nix build never bundles a
              # developer's local front-end build.
              postPatch = ''
                rm -rf web/out
                mkdir -p web/out
                cat > web/out/index.html <<'HTML'
                <!doctype html>
                <title>Peasant — dashboard assets not bundled</title>
                <p>This nix build does not bundle the web dashboard front-end. The CLI works fully; build from source with <code>make build</code> for the embedded UI.</p>
                HTML
              '';

              nativeBuildInputs = nativeBuildDeps pkgs;

              # No -race: the data-race detector requires cgo, and this build is
              # CGO_ENABLED=0. The authoritative race-enabled FULL suite runs in
              # `make check` / CI (incl. the CGO=0 leg).
              #
              # Two packages are excluded because they are environment-dependent
              # INTEGRATION suites that cannot run in nix's hermetic build sandbox
              # (they pass in `make check` / CI, which is the real gate):
              #   - internal/e2e: resolves its testdata via runtime.Caller, which
              #     buildGoModule's `-trimpath` rewrites to a module-relative path
              #     that does not exist at test time → fixture-not-found.
              #   - internal/ingest: ExecGitResolver/CommitDetector tests shell
              #     out to a real `git` and build throwaway repos, neither of
              #     which the sandbox provides.
              # This checkPhase is a packaging sanity gate over the other 21
              # packages (incl. pkg/redact — the CGO=0 redaction seam this build
              # mode actually exercises), not a substitute for the full suite.
              checkPhase = ''
                runHook preCheck
                go test $(go list ./... | grep -vE '/internal/(e2e|ingest)$')
                ${extraCheckPhase}
                runHook postCheck
              '';

              postInstall = extraInstallPhase;

              meta = with pkgs.lib; {
                description = "Local-first coding-agent transcript analytics, redaction, and publishing CLI";
                homepage = "https://github.com/peasant-labs/peasant";
                # PLACEHOLDER license, honestly NON-SPDX: the committed LICENSE is
                # a BSD-3-Clause variant with a branding-preservation clause
                # (adapted from the Open WebUI License), which is NOT any standard
                # SPDX license. Do NOT claim a stock SPDX id (that would be a lie
                # to nixpkgs/validators).
                # free = false: the branding-preservation clause (clause 4 of the
                # LICENSE) restricts altering/removing "Peasant" branding above a
                # 50-user threshold, so the placeholder is arguably NON-free — the
                # conservative, honest truth for a BSD-3+branding variant. The
                # flake already sets config.allowUnfree = true, so this does not
                # block `nix build`. The FINAL license is deferred (#10) and gates
                # the public flip (runbook); revisit free/SPDX at that decision.
                license = {
                  shortName = "LicenseRef-Peasant-Placeholder";
                  fullName = "Peasant License (placeholder; BSD-3-Clause variant with a branding-preservation clause, adapted from the Open WebUI License)";
                  free = false;
                };
                mainProgram = "peasant";
              };
            };

            # ----------------------------------------------------------
            # Development Shell
            # ----------------------------------------------------------

            devShell = pkgs.mkShell {
              name = "${pname}-dev";
              inputsFrom = [ package ];
              packages = (devTools pkgs);

              shellHook = ''
                echo "Go $(go version | cut -d' ' -f3) dev shell"
                export CGO_ENABLED=1
                source .envrc.local
              '';
            };

          in
          {
            packages.default = package;
            packages.${pname} = package;

            devShells.default = devShell;

            # Quick check: nix flake check
            checks.build = package;
          }
        );
    in
    mkOutputs nixpkgs;
}
