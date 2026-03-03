# Relatório de Pesquisa: Txlog Server

## 1. Visão Geral do Projeto

O **Txlog Server** é o servidor central projetado para receber, armazenar e relatar dados enviados por instâncias
do "Txlog Agent". Ele atua como uma central analítica e de monitoramento de infraestrutura, com foco elementar
no rastreamento de operações de gerenciamento de pacotes (instalações, atualizações, remoções) e no estado geral
das execuções sistêmicas das máquinas conectadas de um ambiente distribuído.

## 2. Arquitetura Técnica e Tecnologias Base

- **Linguagem**: O servidor é desenvolvido inteiramente em Go (Golang).
- **Framework Web**: Utiliza o framework `gin-gonic/gin` para expor duas frentes: a API RESTful e o
  portal/dashboard Web iterativo.
- **Banco de Dados**: Depende de um banco transacional PostgreSQL, cujo acesso é orquestrado através do pacote
  puro `lib/pq` (`database/sql`). Traz consigo um motor embutido (`golang-migrate/migrate/v4`) que realiza a
  análise e a imposição automatizada do esquema de dados a cada inicialização via Scripts SQL.
- **Frontend**: Não é uma SPA tipica (Single Page Application, Ex: React). Pelo contrário, adota a poderosa engine
  Server-Side do Go (`html/template`) para processar e renderizar as visões utilizando a pasta `templates/`. O
  **Tailwind CSS** suporta o visual das estruturas HTML geradas usando um build via executável standalone.
- **Portabilidade Total**: Arquivos estáticos essenciais, como CSS, Imagens, arquivos HTML (templates) e até mesmo
  migrações de banco em SQL, são todos compilados para dentro do binário resultante por intermédio da biblioteca
  `embed.FS` nativa. Isto gera um executável agnóstico e independente do host, perfeito para Docker e Kubernetes.

## 3. Modelos de Dados (Entidades do Domínio)

Ao analisar o pacote `models`, identificam-se de imediato os artefatos de controle chave:

- **Assets (Máquinas)**: Cadastro e acompanhamento de ponta a ponta dos computadores corporativos (identificados
  univocamente pelo `MachineID` e `Hostname`). O Server sabe exatamente qual arquitetura, sistema operacional e
  qual a versão do *agent* está em execução neste Asset.
- **Transactions**: Agregado macro das ações efetuadas frequentemente por um sistema de repositórios (yum, apt,
  dnf, zypper). Guarda informações úteis como linha de comando submetida, usuário que autorizou, tempo decorrido,
  saída padrão de terminal obtida (scriptlet output) e códigos numéricos de retorno do processo.
- **Transaction Items**: Um detalhamento atômico associado a cada *Transaction*. Para cada requisição de alteração
  de pacotes feita pela máquina, detalha rigorosamente e rastreia o pacote (Ex: `name`, `version`, `release`,
  `epoch`, `arch`, `repo`).
- **Executions**: Um log vitalício reportando os contatos periódicos (pulsações) que o Agente do Txlog propôs ao
  servidor central. Exibe se houve êxito de contato, a quantidade de remessas e sinaliza bandeiras imperativas,
  como uma booleana `needs_restarting` provendo o alinhamento de servidores carentes de reboots pendentes.

## 4. Endpoints e Interface do Usuário (UI/API)

- **Módulo Administrador & Relatórios (Visualização)**: Injetado por funções customizadas Go Template (Ex:
  transformações visuais `formatDateTime`, `timeStatusClass`), e hospedadas nas rotas limpas do gin (`/` index,
  `/assets`, `/packages`, e as telas complexas analíticas como `/analytics/compare`, `adoption`, `freshness`).
- **Módulo de Ingestão (`/v1`)**: Com proteção garantida por uma autorização com validade cruzada com API Key
  Middleware, esses canais servem inteiramente aos Txlog Agents. Destaques aos pontos focais `POST /v1/transactions`
  (que inclusive cadastra as transações já ligadas de maneira segura através de transações atômicas `tx.Begin()`
  com *Rollback*) e `POST /v1/executions`.

## 5. Autenticação e Segurança Multinível

A aplicação suporta múltiplos conectores para segurança e autorização de interfaces.

1. **Nativo/Tratamento Indiferenciado**: Opcionalmente funcional para ambientes locais sem requisitos extras.
2. **OIDC (OpenID Connect)**: Integração e permissibilidade atrelada a provedores de terceiros modernos OAuth.
3. **LDAP**: Módulo extensivo customizado a se ligar aos serviços complexos do Active Directory/OpenLDAP visando
   autorizar o corpo funcional de uma empresa na adoção deste painel. Permite consultas elaboradas baseadas unicamente
   em Filtros (`LDAP_USER_FILTER`, `LDAP_GROUP_FILTER`) e distingue Grupos Administrativos de Grupos de Leitura.

*Agentes requerem chaves de API restritas vinculadas via interface administrativa que valida todas as entradas
aos endpoints `/v1/`.*

## 6. Agendador Background (Subsistema Scheduler)

Hospedado de forma nativa pela biblioteca Cron `github.com/mileusna/crontab` atuante de maneira perptenua em
sua goroutine isolada:

- **`housekeepingJob`**: Job empenhado rotineiramente a higienizar o banco logístico, purgando do sistema as
  execuções antigas sob uma regra regida por ambiente (`CRON_RETENTION_DAYS`), varrendo registros e
  `transaction_items` orfãos cujas máquinas repousam por 90+ dias na condição inoperante.
- **`refreshMaterializedViewsJob`**: Como lidamos com logs agressivos de transações, este processo executa um
  `REFRESH MATERIALIZED VIEW CONCURRENTLY` com constância máxima (5 minutos). Ele atualiza caches precomputados e
  pré-paginados relativas a dashboards (`mv_package_listing`, `mv_dashboard_agent_stats`, etc.), tirando o impacto
  da agregação em tempo real do banco na hora exata em que o cliente dá F5 na aba de relatórios.
- **`statsJob` e `latestVersionJob`**: Ocupam-se, respectivamente, da coleta de estatísticas demográficas de
  instalações dos últimos 30 dias na base PostgreSQL e da verificação pública, na URL corporativa externa, de qual o
  último versionamento global suportado pelo `txlog-agent`.

*Detalhe de alta resiliência: Devido à possível implantação paralela em Nuvem/Kubernetes com várias cópias do
Server (Replicas>1), os cron jobs estão amparados por um **Distributed Lock** amarrado às tabelas de PostgreSQL
(`cron_lock`), assegurando que duas ou mais instâncias gêmeas nunca disparam as manutenções perigosas em horários
sobrepostos.*

## 7. Conclusões e Destaques Importantes

A robustez identificada sublinha um projeto idealizado com padrões sólidos de produção. A performance via
**Materialized Views Concorrentes** com um limite engessado (`SetMaxOpenConns` em 25 posições máximas de
conexão) comprova uma arquitetura feita num pilar de performance antecipada à *Storms* massivos de eventos emitidos
subitamente por todos os relatórios de ativos nas redes locais. Ademais, a portabilidade impulsionada pelo `embed`
unida aos scripts prontos e visões acentuam ser uma proposta que prioriza a "facilidade do deploy na linha
de chegada" para DevOps Administrators.
