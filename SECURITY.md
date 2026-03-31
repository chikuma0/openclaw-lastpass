# Security Policy

## Supported Versions

Security fixes are only guaranteed for the latest released version and the current `main` branch.

## What To Report

Please report issues such as:

- secret values being logged or persisted unexpectedly
- unsafe handling of command arguments or environment variables
- installer behavior that could lead to executing the wrong asset
- path or file-permission issues that expose resolver mappings
- protocol behavior that leaks secrets outside the expected command contract

## How To Report

Please do not open a public issue with secrets, tokens, vault contents, or private reproduction details.

Prefer GitHub's private vulnerability reporting for this repository if it is available. If private reporting is not available, open a minimal public issue without sensitive details and ask for a secure reporting path.

## Handling Guidance

- Never include real secrets in issues, pull requests, or screenshots.
- Redact entry names if they would reveal sensitive internal systems.
- If a bug can be reproduced with a fake token or fake mapping, use the fake version.

## Project Boundaries

This project is read-only and deliberately small. That reduces some classes of risk, but it does not remove the need for careful review around:

- subprocess execution through `lpass`
- machine-readable output contracts
- file permissions on local mapping files
- install and release asset integrity
