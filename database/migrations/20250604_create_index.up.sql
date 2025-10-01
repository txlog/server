CREATE INDEX IF NOT EXISTS idx_executions_ranked_optim ON public.executions (hostname, executed_at DESC) INCLUDE (machine_id, needs_restarting);
