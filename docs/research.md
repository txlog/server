# Relatório de Pesquisa: Txlog Server

## 1. Visão Geral
O **Txlog Server** é o sistema centralizado do ecossistema Txlog, projetado para receber, armazenar e analisar dados de log de transações de gerenciadores de pacotes (como YUM e DNF) enviados por **Agentes Txlog** instalados em diversos servidores (assets). O sistema provê uma plataforma unificada para auditoria, controle de atualizações, análise de segurança e monitoramento de frota (infraestrutura) baseada em sistemas operacionais RPM-based, em sua maior parte (AlmaLinux, Rocky Linux, Red Hat, etc.).

## 2. Arquitetura e Tecnologias
Construído para ser eficiente e portável, as principais tecnologias do projeto incluem:
- **Linguagem:** Go (Golang) com o framework web **Gin** (`github.com/gin-gonic/gin`).
- **Banco de Dados:** PostgreSQL (`lib/pq`), utilizando raw queries e transações. O banco faz uso de **Materialized Views** e VIEWS complexas para calcular agregações de desempenho e painéis (dashboards).
- **Frontend / Interface do Usuário:** Server-Side Rendering (SSR) utilizando as templates HTML padrão do Go (`html/template`) integradas com **Tailwind CSS**. A build do Tailwind é gerada via CLI standalone para evitar a necessidade do Node.js no ambiente de produção.
- **Autenticação Flexível:** Implementação dual de provedores, suportando simultaneamente ou individualmente **OIDC (OpenID Connect)** e **LDAP / Active Directory**.
- **Agendador Integrado (Scheduler):** Gerencia tarefas periódicas como cron jobs, utilizando a biblioteca `crontab` em memória com suporte a _locks_ distribuídos no banco para evitar conflitos de execução em cenários Multi-Pod (Kubernetes).
- **Documentação da API:** Gerada via Swagger / Swaggo.

## 3. Estrutura de Diretórios
O código do servidor segue um padrão MVC e está muito bem organizado nas seguintes pastas principais:
- `auth/`: Implementa toda a lógica dos provedores OIDC e LDAP, mapeando grupos, criando sessões de usuários e integrando metadados de identidade ao banco de dados interno de usuários do sistema.
- `controllers/`: Gerencia todas as rotas (HTTP/REST) do sistema.
  - O subdiretório `api/v1/` contém a API exposta para consumo pelo _Txlog Agent_.
  - A raiz desta pasta gerencia as visualizações e interações visuais via navegador, desde os dashboards até telas de admin e analíticas.
- `database/`: Conexões com o banco PostgreSQL e seus "migrations" escritos puramente em arquivos `.sql` (`.up.sql` e `.down.sql`).
- `models/`: Definições das entidades fundamentais do sistema (`Transaction`, `TransactionItem`, `Execution`, `User`, `Asset`, `Vulnerability`). Contém também a lógica avançada para inserções como o mecanismo `AssetManager`.
- `scheduler/`: Responsável pelos "Cron Jobs" de limpeza (housekeeping), estatísticas, atualização de Vulnerabilidades (OSV) e atualização do Materialized Views de performance.
- `middleware/`: Interceptadores do Gin (ex: injeção de variáveis de ambiente nas requests, middlewares de Autenticação LDAP/OIDC, autorização de API Keys para Agentes e Admin).
- `util/`: Funções utilitárias abrangentes, formatadores para a View, e comunicação de requisições de API para a base de vulnerabilidades OSV.
- `templates/` e `static/`: Contêm quase toda a casca visual da aplicação, HTMLs com _Tailwind classes_ e pequenos pedaços de CSS.

## 4. Fluxo de Dados e Funcionalidades Core

### Máquinas (Assets)
Qualquer máquina comunicando-se com o Txlog reporta o `MachineID`, `Hostname` e o SO (ex: AlmaLinux 9). O servidor realiza Upserts registrando o momento em que a máquina "foi vista pela última vez" e se ela precisa de "reinicialização" (Needs Restarting) em decorrência de eventos como updates de kernel constatados pelo agente na ponta.

### Execuções (Executions)
Representam as submissões periódicas de log que os agentes reportam, sejam elas acompanhadas ou não de pacotes recém-instalados. Servem como heartbeat avançado informando versão do agente, sistema operacional exato daquele payload da máquina e quaisquer erros que ocorreram do lado do agente.

### Transações (Transactions) e Pacotes (Transaction Items)
O coração do problema resolvido pelo Txlog. Cada instalação/remoção feita no pacote (via comando dnf upgrade, por exemplo) gera uma "Transação", cujo detalhes salvam:
- Ação executada (Instalação, Downgrade, Erase, Upgrade).
- Usuário que comandou e a linha de comando original efetuada.
- Código de retorno.
- Detalhes (itens) pacote por pacote (Nome, Versão, Repositório e Arquitetura).
Essa agregação ajuda a manter os audit logs da infraestrutura visíveis dentro de sua totalidade.

### Motor de Vulnerabilidades (Security / OSV)
O sistema possui uma engine de segurança que correlaciona de forma cruzada as "Transações" das máquinas com o banco de dados global público de vulnerabilidades de código aberto do Google OSV (Open Source Vulnerability).
Quando há um _Upgrade_:
- O `scheduler` em background consulta cada pacote e versão recém-aplicada mapeando contra o "ecossistema" do Linux.
- Recupera-se a pontuação CVSS e gravidade (Low, Medium, High, Critical) da vulnerabilidade.
- Calcula-se então, numa métrica denominada "Scoreboard", qual foi o Risco Mitigado (`risk_score_mitigated`) com aquela atualização de pacote, mostrando ao sysadmin relatórios de pacotes corrigidos (Vulns Fixed) ou introduzidos por downgrade.

## 5. Rotinas Automatizadas (Scheduler)
As rotinas funcionam utilizando _locks distríbuidos_ via Postgres inserindo linhas em uma tabela `cron_lock` para garantir que apenas um container servidor em um cluster executará o job.
As tarefas incluem:
- **housekeepingJob:** Limpeza configurável de Execuções e pacotes "órfãos" (por inatividade prolongada de máquinas desativadas) dependente do valor `CRON_RETENTION_DAYS`.
- **statsJob:** Contabilização massiva do número de execuções mensais, quantidade de pacotes modificados, para facilitar e deixar veloz a renderização diária.
- **UpdateVulnerabilitiesJob:** Sincronização batch com a API externa da base OSV periodicamente de acordo com a regra CRON (`CRON_OSV_EXPRESSION`). Subdivide em pedaços (Chunks) e faz a ingestão das correções e "score" de vulnerabilidades do pacote.
- **refreshMaterializedViewsJob:** Recalcula visualizações materializadas das queries que sustentam o endpoint e interface de Dashboard de Pacotes.

## 6. Autenticação e Segurança da API
Por padrão, o Txlog Server roda **sem autenticação** liberando todo o consumo (inclusive para leitura), porém ele embute as bibliotecas do **OAuth2 / OIDC** e o pacote **go-ldap** de Active Directory.
Uma vez providenciado os `.envs` como `LDAP_HOST` ou `OIDC_CLIENT_ID`, o sistema:
1. "Trava" a Web Interface, exigindo login.
2. Intercepta requests para os endpoints `/v1/` através de `APIKeyAuth` via Headers com API Keys auto-geradas a nível de banco de dados, sendo essa a única maneira dos "Agents" continuarem efetuando suas postagens (reports). Diferentemente da Interface Web que passa a aceitar Cookies de Sessão criados no banco de dados providos por `CreateUserSession`.
O serviço de LDAP faz parse das permissões via `LDAP_ADMIN_GROUP` ou `LDAP_VIEWER_GROUP` definindo qual o escopo (is_admin) do usuário no portal.

## Conclusão
O **Txlog Server** é uma base madura, extremamente acoplada em eficiência via goroutines e sem "bloat" de dependências de frontend. O software concentra a complexidade nas interações com PostgreSQL e abstrai com maestria a união do *gerenciamento de inventário de máquinas*, *auditoria de comandos DNF* e a *consciência cibernética (Vulnerability Risk Assessment)* do pacote de maneira consolidada para o Administrador de Sistemas de plataforma Linux.
