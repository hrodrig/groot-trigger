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

## Known exposure — CVE-2026-98432 (Linux kernel IPv4 TCP UAF)

**This is a host/kernel issue, not a bug in the trigger.** Go 1.24+ enables
Multipath TCP (MPTCP) on Linux listening sockets by default
(`GODEBUG multipathtcp=2`). **CVE-2026-98432** is a use-after-free in the Linux
kernel IPv4 TCP options path (CVSS 9.8) with a public remote-root exploit
against hosts exposing an MPTCP listener.

| Layer | State in this repo |
|-------|--------------------|
| Binary | The trigger listener calls `net.ListenConfig.SetMultipathTCP(false)`, so it never opens an MPTCP socket. This overrides both the system default and `GODEBUG`. |
| Module default | `go.mod` sets `godebug multipathtcp=0` for every binary in the module. |
| Deploy manifests | `deploy/k8s/always/deployment.yaml` sets `GODEBUG=multipathtcp=0` (defense in depth, also covers an older image tag). |

**The real fix is a patched kernel.** Disabling MPTCP only closes the published
MPTCP exploit path — ordinary IPv4 TCP can still reach the underlying race.
Fixed kernel versions: `5.10.272`, `5.15.223`, `6.1.190`, `6.6.159`,
`6.12.113`, `6.18.56`, `7.2.10`, `7.3` or later.

Notes:

- Binding to `127.0.0.1` does **not** disable MPTCP.
- `GODEBUG=multipathtcp=1` in the environment does **not** re-enable MPTCP for
  the trigger: the explicit `SetMultipathTCP(false)` wins.
- The trigger does not need MPTCP; there is no supported way to turn it back on.

## Reporting a vulnerability

**Do not open a public issue** for undisclosed vulnerabilities.

- Preferred: GitHub Security Advisories on this repository (once the remote exists).
- Alternatively: contact the maintainer via [github.com/hrodrig](https://github.com/hrodrig).

Include description, reproduction, affected versions, and impact.
