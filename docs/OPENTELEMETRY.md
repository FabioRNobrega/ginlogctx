# OpenTelemetry vs ginlogctx

## What is OpenTelemetry?

OpenTelemetry is an observability standard and toolkit used to collect and send
telemetry such as:
- traces
- metrics
- logs

In practice, teams often adopt OpenTelemetry to instrument applications with
distributed tracing. That means a single request can be represented as a trace
made of multiple spans across services, databases, queues, and other systems.

This is powerful, but it also introduces extra concepts and moving parts:
- trace IDs
- span IDs
- propagators
- exporters
- collectors
- sampling
- instrumentation libraries

For teams that truly need application performance monitoring and distributed
trace analysis, OpenTelemetry is a strong choice.

## What problem does ginlogctx solve?

`ginlogctx` solves a smaller and simpler problem:

"I want every Gin request to have a request ID, and I want that request ID and
other request-scoped fields to appear automatically in my Logrus logs."

That makes it easier to:
- correlate logs from the same request
- attach business context such as `user_id` or `tenant_id`
- forward a request ID across services
- search and group logs in tools like Datadog

## Why use this instead of OpenTelemetry?

For many projects, especially simpler APIs and services, logs are enough.

If your main goal is:
- request correlation
- debugging
- operational visibility in logs
- tracking requests in Datadog without enabling APM

then `ginlogctx` can be a much simpler fit.

You do not need to introduce:
- trace exporters
- span modeling
- collectors
- distributed tracing configuration
- APM rollout complexity

Instead, you get:
- an out-of-the-box request ID for Gin
- automatic request-scoped log enrichment for Logrus
- a small API surface
- easy adoption in existing services

## When should you prefer OpenTelemetry?

Choose OpenTelemetry when you need more than correlated logs, for example:
- full distributed traces across many services
- span timing and hierarchy
- instrumentation for databases, queues, and external APIs
- tracing-based performance analysis
- vendor-neutral telemetry pipelines

## A practical rule of thumb

Use `ginlogctx` when:
- logs are your main source of observability
- you want request-level tracking with minimal setup
- Datadog log search is enough for your current needs

Use OpenTelemetry when:
- you need full tracing and APM-style visibility
- request correlation in logs is no longer enough

These options are not mutually exclusive. A team can start with `ginlogctx` for
simple request correlation and adopt OpenTelemetry later when the system grows
or deeper tracing becomes necessary.
