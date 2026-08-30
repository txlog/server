-- Reverse 202603020002_add_ecosystem: drop the ecosystem column and restore the
-- original primary key from 202603020001_add_vulnerabilities.
--
-- The table is truncated, as the up migration also does. Without ecosystem the
-- remaining columns are no longer unique — that is precisely why the column was
-- added to the key — so rows would collide when the old primary key is
-- recreated. The data is repopulated by the OSV job.
TRUNCATE TABLE package_vulnerabilities;

ALTER TABLE package_vulnerabilities DROP CONSTRAINT IF EXISTS package_vulnerabilities_pkey CASCADE;

ALTER TABLE package_vulnerabilities DROP COLUMN IF EXISTS ecosystem;

ALTER TABLE package_vulnerabilities ADD PRIMARY KEY (package_name, version, vulnerability_id);
