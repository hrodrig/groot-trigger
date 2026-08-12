# Security policy — groot-trigger

## Scope

HTTP companion that creates Kubernetes Jobs running [`groot`](https://github.com/hrodrig/groot) collect.
Covers the trigger binary, its API key / rate-limit / proxy handling, and the published container image.

Treat collect Job output and object-storage credentials as sensitive. Never log API keys or cloud secrets
(see `docs/SPECIFICATIONS.md`).

## Supported versions

| Version | Supported |
| ------- | --------- |
| Latest release (when published) | Yes |
| Older releases | No — upgrade |

Until the first tagged release, treat `develop` as the security surface.

## Reporting a vulnerability

**Do not open a public issue** for undisclosed vulnerabilities.

- Preferred: GitHub Security Advisories on this repository (once the remote exists).
- Alternatively: contact the maintainer via [github.com/hrodrig](https://github.com/hrodrig).

Include description, reproduction, affected versions, and impact.
