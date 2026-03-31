# Vision

`openclaw-lastpass` is a small, read-only LastPass bridge for OpenClaw and direct CLI workflows.

It exists for people who are already on LastPass and want a safer workflow for machine secrets than docs, spreadsheets, or scattered notes. It is not trying to replace every secret manager decision a team might make. It is trying to solve one narrow problem cleanly: map a small set of machine-usable secret IDs to LastPass entries and resolve them through `lpass`.

## Project Priorities

- small over broad
- explicit over magical
- review-first over auto-apply
- read-only over vault mutation
- trust and auditability over feature count

## Audience

The best fit is:

- developers already using LastPass locally
- teams adopting OpenClaw's exec-provider flow
- people who want a practical bridge before deciding whether they need a larger secret-management change later

## Boundaries

This project is not:

- a general password manager
- a browser extension
- a browser automation tool
- a full-vault export or migration system
- a universal auth or secret-management platform

## Contribution Direction

Future changes should preserve the same shape:

- read-only
- least-privilege
- metadata-first discovery
- human-reviewed apply
- simple install and onboarding

If a proposed feature makes the project broader, more magical, or harder to audit, it is probably the wrong fit.
