# Security Policy

## Supported Versions

Only the [latest release](https://github.com/redjax/syst/releases) is supported with security updates.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please use [GitHub's private vulnerability reporting](https://github.com/redjax/syst/security/advisories/new) to submit a report.

You should receive a response within 72 hours. If the vulnerability is confirmed, a fix will be released as soon as possible.

Please include:

- A description of the vulnerability
- Steps to reproduce
- Affected versions
- Any potential impact

## Security Scanning

This project runs automated security scans on a daily schedule:

- **OSV Scanner** — dependency vulnerability scanning via [OSV.dev](https://osv.dev/)
- **govulncheck** — Go-specific vulnerability analysis
- **CodeQL** — static analysis for code quality and security
- **Gitleaks** — secret detection in source and commit history
