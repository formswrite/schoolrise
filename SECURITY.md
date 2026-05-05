# Security policy

SchoolRise stores student data — including names, ages, guardians, and assessment scores. Vulnerabilities that could expose that data are taken seriously and treated as the highest priority.

## Reporting a vulnerability

**Do not open public issues for security vulnerabilities.**

Email **security@formswrite.com** with:

1. A description of the vulnerability and the affected component (service name, file path, function if known).
2. Steps to reproduce against a local `encore run` or a reference deployment.
3. The impact: which data could be read, modified, or deleted; whether the vulnerability is pre- or post-authentication; what privilege level is required.
4. A proof-of-concept request, payload, or script if possible.

We acknowledge reports within **two business days** and aim to ship a fix within **fourteen days** for critical issues, **thirty days** for high-severity, and **ninety days** for medium- and low-severity. We will keep you updated on progress and credit you in the release notes when the fix ships, unless you prefer to remain anonymous.

## Supported versions

Until v1.0.0, only the latest minor release receives security updates. After v1.0.0 we will publish a backport policy here.

## Scope

In scope:
- Encore service code (`auth`, `tenancy`, `people`, `academics`, `enrollment`, `forms`, `assessment`, `progression`, `imports`, `notifications`, `ai`)
- Shared Go packages under `pkg/`
- The SvelteKit frontend in `apps/web/`
- The default `deploy/docker-compose.yml` and `infra-config/selfhost.json`
- The first-boot bootstrap (`pkg/seed/bootstrap.go`)
- Authentication, authorisation, session token handling, assignment-token signing
- SQL injection, XSS, CSRF, SSRF, IDOR, broken access control
- Dependency vulnerabilities surfaced by `govulncheck` or Dependabot

Out of scope:
- Vulnerabilities in third-party services (Resend, OpenAI, GitHub, Postgres, Encore.go itself) — report directly to the upstream maintainer
- Self-hosted deployments running modified forks
- Social-engineering attacks on the maintainers
- Physical attacks on ministry infrastructure

## What we ask

- Test only against your own local `encore run` or your own self-hosted deployment.
- Do not access, modify, or exfiltrate data belonging to a partner ministry.
- Give us reasonable time to fix before public disclosure.

## What you can expect

- Acknowledgement within two business days.
- A clear severity assessment and timeline.
- Credit in the release that ships the fix, unless you prefer anonymity.
- For critical issues, the fix ships before public disclosure of the vulnerability details.

For commercial-support arrangements that include guaranteed response SLAs and private patch distribution, email **commercial@formswrite.com**.
