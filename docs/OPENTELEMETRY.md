# Guia de Teste - OpenTelemetry no Txlog Server

Este guia mostra como testar a implementação do OpenTelemetry no Txlog Server.

## Teste Rápido com Jaeger

A maneira mais fácil de testar é usar o Jaeger All-in-One localmente.

### 1. Iniciar Jaeger

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

- **16686**: Jaeger UI
- **4318**: OTLP HTTP endpoint

### 2. Configurar Txlog Server

Adicione ao seu arquivo `.env`:

```bash
# OpenTelemetry Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=txlog-server
OTEL_SERVICE_VERSION=dev
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=development
```

### 3. Iniciar o Txlog Server

```bash
make run
```

Você verá nos logs:

```text
INFO OpenTelemetry: initialized successfully
INFO OpenTelemetry: exporting to http://localhost:4318
```

### 4. Fazer Requisições

Faça algumas requisições para gerar traces:

```bash
# Página principal
curl http://localhost:8080/

# API endpoint
curl http://localhost:8080/v1/version

# Swagger docs
curl http://localhost:8080/swagger/index.html
```

### 5. Visualizar Traces no Jaeger

Abra o navegador em: <http://localhost:16686>

1. No campo "Service", selecione **txlog-server**
2. Clique em "Find Traces"
3. Você verá todas as requisições HTTP com seus detalhes:
   - Duração
   - Status HTTP
   - Rota
   - Método
   - **Queries SQL executadas** (como spans filhos)

### 6. Visualizar SQL Queries nos Traces

Cada requisição HTTP que executa queries SQL mostrará:

**Informações capturadas:**

- Texto da query SQL (sem parâmetros sensíveis)
- Tempo de execução
- Número de linhas afetadas
- Erros (se houver)
- Conexões ao banco de dados

**Exemplos de spans SQL que você verá:**

- `sql:query` - SELECT queries
- `sql:exec` - INSERT, UPDATE, DELETE
- `sql:prepare` - Prepared statements
- `sql:begin` - Início de transação
- `sql:commit` - Commit de transação
- `sql:rollback` - Rollback de transação

**Atributos capturados:**

- `db.system`: "postgresql"
- `db.name`: nome do banco de dados
- `db.statement`: texto da query SQL
- `db.operation`: tipo de operação (SELECT, INSERT, etc.)

### 7. Visualizar Logs Correlacionados

Os logs também são enviados para o Jaeger com correlação de trace:

- Clique em um trace específico
- Você verá os spans HTTP
- Os logs estarão correlacionados com os trace_id e span_id

## Teste sem OpenTelemetry

Para testar que a aplicação funciona sem OpenTelemetry:

1. **Remova** ou **comente** a variável `OTEL_EXPORTER_OTLP_ENDPOINT` do `.env`

2. Inicie o servidor:

```bash
make run
```

3. Você verá nos logs:

```text
INFO OpenTelemetry: disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)
```

4. A aplicação funciona normalmente, sem enviar telemetria

## Teste com OpenTelemetry Collector

Para ambientes de produção, recomenda-se usar um OpenTelemetry Collector.

### 1. Criar arquivo `otel-collector-config.yaml`

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  logging:
    loglevel: debug
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, jaeger]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

### 2. Docker Compose

```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14250:14250"

  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4318:4318"
    depends_on:
      - jaeger

  txlog-server:
    image: cr.rda.run/txlog/server:main
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
      - OTEL_SERVICE_NAME=txlog-server
      - OTEL_RESOURCE_ATTRIBUTES=deployment.environment=production
      # ... outras variáveis de ambiente
    ports:
      - "8080:8080"
    depends_on:
      - otel-collector
```

### 3. Iniciar

```bash
docker-compose up -d
```

## Verificação de Traces

### O que procurar nos traces

1. **HTTP Spans**:
   - Nome do span: rota HTTP (ex: `GET /v1/version`)
   - Atributos:
     - `http.method`: GET, POST, etc.
     - `http.status_code`: 200, 404, etc.
     - `http.route`: rota da requisição
     - `http.target`: URL completa

2. **SQL Spans** (dentro de HTTP spans):
   - Nome do span: operação SQL (ex: `sql:query SELECT`, `sql:exec INSERT`)
   - Atributos:
     - `db.system`: "postgresql"
     - `db.name`: nome do banco de dados
     - `db.statement`: texto da query SQL
     - `db.operation`: SELECT, INSERT, UPDATE, DELETE, etc.
   - Hierarquia:
     - HTTP Request (span raiz)
       - SQL Query 1 (span filho)
       - SQL Query 2 (span filho)
       - SQL Transaction (span filho)
         - SQL Query 3 (span neto)
         - SQL Query 4 (span neto)

3. **Logs Correlacionados**:
   - Atributo `trace_id`: ID do trace
   - Atributo `span_id`: ID do span
   - Permite correlacionar logs com requisições específicas

## Troubleshooting

### "Failed to initialize telemetry"

Verifique:

- O endpoint OTLP está acessível?
- O formato do endpoint está correto? (<http://host:port>)
- Não inclua `/v1/traces` no endpoint - o SDK adiciona automaticamente

### Traces não aparecem no Jaeger

1. Verifique os logs do servidor:
   - Deve aparecer "OpenTelemetry: initialized successfully"

2. Verifique se o Jaeger está recebendo dados:

   ```bash
   docker logs jaeger
   ```

3. Teste a conectividade:

   ```bash
   curl -v http://localhost:4318/v1/traces
   ```

### Performance

O overhead do OpenTelemetry é tipicamente < 5%:

- Traces são enviados em batch (não bloqueante)
- Logs são processados assincronamente
- Graceful shutdown garante que dados não sejam perdidos

## Integração com Serviços em Produção

### Grafana Cloud

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-prod-us-central-0.grafana.net/otlp
OTEL_EXPORTER_OTLP_HEADERS=authorization=Basic $(echo -n "userid:api-key" | base64)
```

### New Relic

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.nr-data.net
OTEL_EXPORTER_OTLP_HEADERS=api-key=YOUR_LICENSE_KEY
```

### Datadog

Primeiro configure o Datadog Agent para aceitar OTLP:

```yaml
# datadog.yaml
otlp_config:
  receiver:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
```

Depois configure o Txlog Server:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://datadog-agent:4318
```

## Métricas de Sucesso

Após configurar, você deve ver:

✅ Traces de todas as requisições HTTP
✅ **Traces de todas as queries SQL executadas**
✅ Logs correlacionados com trace_id
✅ Latência de requisições HTTP e SQL
✅ Taxas de erro (status 4xx, 5xx e erros SQL)
✅ Throughput (requisições por segundo)
✅ **Performance de queries individuais**
✅ **Identificação de queries lentas (N+1 queries, etc.)**

### Exemplos do que você pode monitorar

**Requisições HTTP:**

- Endpoint mais lento
- Endpoints com mais erros
- Throughput por endpoint

**Queries SQL:**

- Queries mais lentas
- Queries mais frequentes
- Queries com erros
- Tempo médio de execução por tipo de query
- Número de queries por requisição (detectar N+1)

**Transações:**

- Tempo de commit/rollback
- Taxa de sucesso de transações
- Deadlocks e lock waits

## Próximos Passos

1. Configure alertas baseados em latência (HTTP e SQL)
2. Crie dashboards para visualização
3. Use trace sampling em produção para reduzir custo
4. Configure service level objectives (SLOs)
5. **Identifique e otimize queries lentas usando os traces**
6. **Configure alertas para queries que excedem tempo esperado**
