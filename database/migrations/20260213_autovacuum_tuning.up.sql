-- Autovacuum tuning for high-churn tables
ALTER TABLE executions SET (autovacuum_vacuum_scale_factor = 0.05);
ALTER TABLE executions SET (autovacuum_analyze_scale_factor = 0.02);
ALTER TABLE transaction_items SET (autovacuum_vacuum_scale_factor = 0.05);
ALTER TABLE transaction_items SET (autovacuum_analyze_scale_factor = 0.02);
