-- Dashboard materialized views for instant dashboard loading

-- OS distribution stats
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_dashboard_os_stats AS
SELECT os, COUNT(*) AS num_machines
FROM assets
WHERE is_active = TRUE AND os IS NOT NULL
GROUP BY os
ORDER BY num_machines DESC;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_dashboard_os_stats ON mv_dashboard_os_stats (os);

-- Agent version distribution stats
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_dashboard_agent_stats AS
SELECT agent_version, COUNT(*) AS num_machines
FROM assets
WHERE is_active = TRUE AND agent_version IS NOT NULL
GROUP BY agent_version
ORDER BY num_machines DESC;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_dashboard_agent_stats ON mv_dashboard_agent_stats (agent_version);

-- Most updated packages in last 30 days
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_dashboard_most_updated AS
SELECT
    ti.package,
    COUNT(*) AS total_updates,
    COUNT(DISTINCT t.hostname) AS distinct_hosts_updated
FROM public.transaction_items AS ti
JOIN public.transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
JOIN public.assets AS a ON t.machine_id = a.machine_id AND t.hostname = a.hostname
WHERE ti.action = 'Upgrade'
    AND t.end_time >= NOW() - INTERVAL '30 days'
    AND a.is_active = TRUE
GROUP BY ti.package
ORDER BY total_updates DESC
LIMIT 10;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_dashboard_most_updated ON mv_dashboard_most_updated (package);
