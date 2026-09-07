# Security policy

## Reporting

Please report vulnerabilities privately to the project maintainers rather than opening a public issue. Until a dedicated security address is published, use GitHub's private vulnerability reporting feature on the repository.

## Local trust boundary

RepoSweep reads filenames and filesystem activity from the selected repository. Future analyzers may read source content. The server therefore binds to `127.0.0.1` by default.

Do not expose RepoSweep on a shared or public network. The prototype does not provide authentication, authorization, TLS, or origin-based request protection. Treat agent event input as untrusted; the server bounds JSON bodies, rejects unknown fields, and the dashboard escapes rendered values.
