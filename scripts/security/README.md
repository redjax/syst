# Security Scripts <!-- omit in toc -->

Bash & Powershell scripts for Go security tools.

## Table of Contents <!-- omit in toc -->

- [Tools](#tools)
- [Installation Scripts](#installation-scripts)
- [Security Scanning](#security-scanning)
  - [Basic Usage](#basic-usage)
  - [Windows PowerShell](#windows-powershell)
- [Common Gosec Rules](#common-gosec-rules)
- [Running Other Security Tools](#running-other-security-tools)
  - [govulncheck](#govulncheck)
  - [osv-scanner](#osv-scanner)
- [CI/CD Integration](#cicd-integration)
- [Suppressing False Positives](#suppressing-false-positives)

## Tools

- **gosec** - Go security analyzer that scans code for common security issues.
- **govulncheck** - Go vulnerability scanner that checks for known vulnerabilities in dependencies.
- **osv-scanner** - Dependency vulnerability scanner using the OSV database.

## Installation Scripts


**Linux/macOS:**

```bash
./scripts/security/tools/install-gosec.sh
./scripts/security/tools/install-govulncheck.sh
./scripts/security/tools/install-osv-scanner.sh
```

**Windows:**

```powershell
.\scripts\security\tools\install-gosec.ps1
.\scripts\security\tools\install-govulncheck.ps1
.\scripts\security\tools\install-osv-scanner.ps1
```

## Security Scanning

Individual scan scripts for each security tool are provided in the `scans/` directory.

### Basic Usage

**Scan everything:**
```bash
./scripts/security/scans/gosec.sh
```

**Scan specific rules:**
```bash
# Single rule
./scripts/security/scans/gosec.sh --rule G304

# Multiple rules
./scripts/security/scans/gosec.sh --rule G304 --rule G301 --rule G104
```

**Filter by severity:**
```bash
# Only high severity issues
./scripts/security/scans/gosec.sh --severity high

# Multiple severities
./scripts/security/scans/gosec.sh --severity high --severity medium
```

**Change output format:**
```bash
# JSON output
./scripts/security/scans/gosec.sh --format json

# HTML report
./scripts/security/scans/gosec.sh --format html > security-report.html

# YAML output
./scripts/security/scans/gosec.sh --format yaml
```

**Combine filters:**
```bash
# Specific rules with JSON output
./scripts/security/scans/gosec.sh --rule G304 --rule G301 --format json

# High severity issues only, as JSON
./scripts/security/scans/gosec.sh --severity high --format json
```

### Windows PowerShell

```powershell
# Scan everything
.\scripts\security\scans\gosec.ps1

# Specific rules
.\scripts\security\scans\gosec.ps1 -Rule G304
.\scripts\security\scans\gosec.ps1 -Rule G304,G301,G104

# Filter by severity
.\scripts\security\scans\gosec.ps1 -Severity high

# JSON output
.\scripts\security\scans\gosec.ps1 -Format json

# Combined
.\scripts\security\scans\gosec.ps1 -Rule G304 -Format json
```

## Common Gosec Rules

| Rule | Description | Severity |
|------|-------------|----------|
| G104 | Unhandled errors | LOW |
| G107 | URL provided to HTTP request | MEDIUM |
| G110 | Decompression bomb | MEDIUM |
| G115 | Integer overflow | HIGH |
| G201 | SQL injection via string concatenation | MEDIUM |
| G204 | Subprocess launched with variable | MEDIUM |
| G301 | Poor file permissions (directory) | MEDIUM |
| G302 | Poor file permissions (file) | MEDIUM |
| G304 | File path provided as taint input | MEDIUM |
| G306 | Poor file permissions (write) | MEDIUM |
| G401 | Weak crypto (MD5) | MEDIUM |
| G501 | Import of weak crypto | MEDIUM |

## Running Other Security Tools

### govulncheck

Check for known vulnerabilities in Go code:

```bash
# Using the scan script
./scripts/security/scans/govulncheck.sh

# With JSON output
./scripts/security/scans/govulncheck.sh --format json

# PowerShell
.\scripts\security\scans\govulncheck.ps1
```

### osv-scanner

Scan dependencies for vulnerabilities:

```bash
# Using the scan script
./scripts/security/scans/osv-scanner.sh

# With JSON output
./scripts/security/scans/osv-scanner.sh --format json

# Scan specific lockfile
./scripts/security/scans/osv-scanner.sh --lockfile go.sum

# PowerShell
.\scripts\security\scans\osv-scanner.ps1
```

## CI/CD Integration

These scripts can be integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Install security tools
  run: |
    ./scripts/security/tools/install-gosec.sh
    ./scripts/security/tools/install-govulncheck.sh

- name: Run security scan
  run: ./scripts/security/scans/gosec.sh --format json

- name: Check vulnerabilities
  run: ./scripts/security/scans/govulncheck.sh
```

## Suppressing False Positives

Use `#nosec` comments to suppress false positives:

```go
// #nosec G304 - CLI tool reads user-specified files by design
content, err := os.ReadFile(userPath)
```

Always include a comment explaining why the suppression is valid.
