# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | Yes                |
| < 0.3   | No                 |

## Design Principles

- **SSH keys never leave your machine.** Keys are read from `~/.ssh/` and used in-memory only.
- **localhost only.** The web server binds to `127.0.0.1`. No external network access.
- **No database.** No credentials are stored persistently (except optional saved connections in `~/.ape/config.yaml`, which store host/username only, never keys or passwords).

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT open a public issue.**
2. Email **dck.alx@gmail.com** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
3. You will receive a response within **48 hours**.
4. A fix will be developed privately and released as a patch.

## Scope

The following are in scope:

- Authentication bypass or credential exposure
- Remote code execution
- Path traversal via API endpoints
- SSH key leakage or unintended transmission
- CORS or localhost binding issues

Thank you for helping keep A.P.E secure.
