# How to Manage and Clean Inactive Servers

Over time, as infrastructure scales, servers are decommissioned, re-imaged, or
retired. If these servers are not properly cleaned up, they will continue to
appear in the dashboard and skew your vulnerability and package distribution
metrics.

Txlog provides a built-in mechanism to bulk-remove servers that have stopped
reporting data.

## When to Use This Feature

You should use the Inactive Server Cleanup feature when:

- You have recently migrated workloads and terminated old VMs.
- Your dashboard shows a high number of "Inactive" or "Stale" assets.
- You want to free up database storage by deleting historical transaction data
  for servers that no longer exist.

## Performing the Cleanup

The cleanup operation identifies any server that has not sent an execution
payload (check-in) in the last **15 days**.

To perform the cleanup:

1. Log in to the **Administration Panel**.
2. Navigate to the **Settings** or **Database Management** section.
3. Locate the **Inactive Server Cleanup** tool.
4. The UI will display a preview of how many servers are currently considered
   inactive (not seen in > 15 days).
5. Click the button to confirm the cleanup.

> [!CAUTION] **Data Destruction:** This operation is irreversible. It performs a
> cascading delete in a single database transaction. When an inactive server is
> deleted, all of its associated data is permanently removed, including:
>
> - Its entry in the `assets` table.
> - All historical `executions` (logs of when it checked in).
> - All `transactions` (logs of DNF/YUM package updates).
> - All `transaction_items` (the specific packages installed or removed).

## Automatic Re-Enrollment

If a server is accidentally deleted because it was offline for more than 15 days
(e.g., a laptop that was powered off during a vacation), there is no need to
manually intervene on the agent side.

As soon as the server comes back online and the agent runs its next scheduled
task, it will send a new execution payload to the Txlog server. The server will
treat it as a newly enrolled asset, create a new record, and begin tracking its
packages and vulnerabilities from scratch.
