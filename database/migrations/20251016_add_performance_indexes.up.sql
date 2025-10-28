-- Para otimizar buscas por machine_id e queries de retenção de dados
CREATE INDEX IF NOT EXISTS idx_executions_machine_id ON public.executions (machine_id);
CREATE INDEX IF NOT EXISTS idx_executions_executed_at ON public.executions (executed_at DESC);

-- Para otimizar a busca da última transação de uma máquina
CREATE INDEX IF NOT EXISTS idx_transaction_items_machine_id_tx_id ON public.transaction_items (machine_id, transaction_id DESC);

-- Para otimizar a busca de pacotes por nome
CREATE INDEX IF NOT EXISTS idx_transaction_items_package_name ON public.transaction_items (package);

-- Para otimizar a ordenação de usuários no painel de administração
CREATE INDEX IF NOT EXISTS idx_users_created_at ON public.users (created_at DESC);

-- Para otimizar a busca de transações por machine_id e hostname
CREATE INDEX IF NOT EXISTS idx_transactions_machine_id_hostname ON public.transactions (machine_id, hostname);

-- Para otimizar a contagem de pacotes por semana
CREATE INDEX IF NOT EXISTS idx_transaction_items_action ON public.transaction_items (action);
CREATE INDEX IF NOT EXISTS idx_transactions_begin_time ON public.transactions (begin_time);
