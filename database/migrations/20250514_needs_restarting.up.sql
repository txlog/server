ALTER TABLE IF EXISTS public.executions ADD COLUMN needs_restarting boolean;
ALTER TABLE IF EXISTS public.executions ADD COLUMN restarting_reason text;
