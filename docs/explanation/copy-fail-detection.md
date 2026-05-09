# Copy Fail (CVE-2026-31431) Detection

The system natively tracks and monitors assets for the **CVE-2026-31431**
vulnerability, internally referred to as the "Copy Fail" bug. This document
explains the technical logic behind the detection and tracking mechanism.

## Technical Context

CVE-2026-31431 is a critical vulnerability affecting specific system binaries.
When this vulnerability is present, routine file copy operations may silently
fail, leading to data corruption without raising application-level errors.

## The Detection Pipeline

The Txlog Agent deployed on the servers runs a diagnostic check to determine if
the local kernel and coreutils are vulnerable to the Copy Fail bug. This boolean
result is transmitted to the server inside the payload of routine executions.

When the server receives an execution payload:

1. **Extraction:** The server parses the payload looking for the diagnostic
   flag.
2. **Persistence:** The `copy_fail` boolean field in the `assets` database table
   is updated.
   - If the agent reports the vulnerability exists, the field is set to `TRUE`.
   - If the agent confirms the vulnerability is patched, the field is set to
     `FALSE`.
   - If the agent version is older and does not transmit this data, the field
     remains `NULL`.
3. **Surfacing:** The `UpsertAsset` model logic ensures that this status is
   immediately available to the dashboard and API without requiring a separate
   vulnerability scan job.

## Search and Filtering

Because Copy Fail represents a systemic risk rather than a simple outdated
package, it bypasses the standard OSV (Open Source Vulnerabilities) integration
pipeline. Instead, it is treated as a first-class attribute of the machine's
identity.

This architectural decision enables the "Magic Keyword" search filters
(`copyfail:true` and `copyfail:false`).

When an administrator searches for `copyfail:false`, the backend database query
leverages `COALESCE` to include both machines that explicitly reported `FALSE`
and machines whose status is `NULL` (unknown/unreported). This ensures that
machines are not falsely hidden from views merely because they haven't run the
latest diagnostic tool yet, preventing blind spots.

## Safety Guarantees

By tracking the vulnerability as an intrinsic property of the asset rather than
a synthetic package vulnerability, the system guarantees that:

- The Copy Fail status is immune to OSV cache desynchronization.
- The Coral Pulse badge is calculated in real-time on every check-in.
- The data cannot be accidentally purged during a routine vulnerability database
  rebuild.
