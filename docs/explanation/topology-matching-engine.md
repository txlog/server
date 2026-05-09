# Topology Matching Engine

The Topology Matching Engine is responsible for translating raw hostnames into
structured, searchable dimensions (such as environments, services, and sequence
numbers) that the dashboard can consume.

By abstracting away the rigidity of delimiter-specific formats, the engine
provides a flexible, positional matching system that intelligently groups assets
regardless of how their hostnames are constructed.

## From Static Regex to Dynamic Delimiters

Historically, the system relied on static, hardcoded regular expressions to
extract data from hostnames. These patterns expected a strict, predefined order
of elements and rigid delimiters (like hyphens or periods). If an organization
used a different naming convention, the parser would fail to categorize the
server correctly.

To address this, the engine now utilizes a **"grep-like" dynamic pattern
resolution system**. Rather than relying on static regex capture groups, the
engine dynamically builds search patterns at runtime based on the actual
services and environments registered in the database.

## The Positional Pipeline

The engine evaluates hostnames through a series of prioritized steps, rather
than attempting to parse the entire string at once. This ensures robustness even
in mixed-format environments.

1. **Global Longest Match Extraction:** The engine queries the database for
   known environment and service identifiers. It searches the hostname string
   for these identifiers, prioritizing the longest matches first. This prevents
   overlapping substrings (e.g., matching "api" when the actual service is
   "api-gateway") from causing false positives.
2. **Positional Anchoring:** Once the environment and service substrings are
   identified within the hostname, the engine anchors them positionally.
3. **Sequence Number Isolation:** With the context of where the environment and
   service reside, the engine uses negative lookbehinds and non-greedy matching
   to isolate the sequence number (or pod ID). This resolves mixed-pod numbering
   issues (e.g., confusing pod `18` with pod `8`).

## The `:any` Placeholder

A key component of the new architecture is the `:any` placeholder in hostname
templates.

In many naming conventions, organizations inject arbitrary strings (like
geographic region codes, cluster IDs, or department names) into the hostname
that are not strictly relevant for the server's primary classification.

The `:any` placeholder acts as an **"ignore-and-pivoting" segment**. When the
engine encounters `:any` in a registered template, it explicitly ignores
whatever text occupies that position, allowing it to pivot and accurately
capture the crucial data points (`:env`, `:svc`, `:seq`) that follow or precede
it, without requiring the administrator to write complex regular expressions to
skip those segments.

## Logical OR (`|`) Delimiters

To drastically reduce the administrative burden of maintaining templates, the
database engine supports dynamic array parsing using PostgreSQL's native
`unnest` function.

This architecture enables the use of the logical OR operator (`|`) directly
inside the `match_value` of a template
definition. Instead of creating five separate template rules for five distinct
microservices that share the same structural format, an administrator can create
a single rule whose service definition contains
`auth-api|payment-api|notification-api`.

During parsing, the engine "explodes" this pipe-delimited string into an array
and evaluates the hostname against each candidate dynamically.
