# Security Policy

## Supported Versions

Security fixes are applied to the latest release on the default branch (`main`). Older tags generally do not receive backports unless a release is still marked as supported below.

| Version | Supported |
| ------- | --------- |
| Latest release (`main` / newest `v*`) | Yes |
| Older releases | No |

## Reporting a Vulnerability

**Do not** open a public GitHub issue for security vulnerabilities.

Please report vulnerabilities privately using one of these channels:

1. **GitHub Security Advisories** (preferred):  
   [Report a vulnerability](https://github.com/turahe/pkg/security/advisories/new)
2. If private reporting is unavailable, contact the repository maintainers through a private channel and wait for acknowledgment before any public disclosure.

### What to include

- Affected package(s) and package path (for example `github.com/turahe/pkg/jwt`)
- Module / library version or commit SHA
- Clear description of the issue and impact
- Steps to reproduce (PoC, minimal program, or configuration)
- Any known workarounds

### What to expect

- Acknowledgment within **7 days** when possible
- An initial assessment and next steps (fix, decline, or request more info)
- Coordinated disclosure: please give maintainers reasonable time to patch and release before public discussion

We may credit reporters in release notes or advisories when appropriate, unless you ask to remain anonymous.

## Scope

In scope for this repository:

- Vulnerabilities in published Go packages under `github.com/turahe/pkg/...`
- Issues in example wiring or defaults that could cause insecure production use when followed as documented

Out of scope (unless they reveal a library defect):

- Misconfiguration of consumer applications that import this module
- Vulnerabilities solely in third-party dependencies (report those upstream; we track them via `govulncheck` / Dependabot)
- Denial-of-service via unbounded resource use in caller-controlled workloads without a concrete library bug

## Automated security checks

This project runs continuous checks on `main` and pull requests:

- [`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) — known Go module vulnerabilities
- [CodeQL](https://codeql.github.com/) — `security-and-quality` queries for Go
- Dependabot — dependency and Actions updates

See [docs/development.md](docs/development.md) and [`.github/workflows/security.yml`](.github/workflows/security.yml).

## Prefer responsible disclosure

Thank you for helping keep `github.com/turahe/pkg` and its consumers safe.
