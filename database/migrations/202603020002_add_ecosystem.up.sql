TRUNCATE TABLE package_vulnerabilities;
ALTER TABLE package_vulnerabilities DROP CONSTRAINT IF EXISTS package_vulnerabilities_pkey CASCADE;
ALTER TABLE package_vulnerabilities ADD COLUMN ecosystem VARCHAR(255) DEFAULT '';
ALTER TABLE package_vulnerabilities ADD PRIMARY KEY (package_name, version, vulnerability_id, ecosystem);
