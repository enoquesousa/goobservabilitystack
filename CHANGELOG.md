# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado no [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Não lançado]

### Planejado

- Pipeline CI no GitHub Actions (build, test, lint, govulncheck).
- Taxa de amostragem OTel configurável via variável de ambiente.
- Imagem Docker multi-arquitetura publicada no ghcr.io.
- Suporte a mTLS no exportador gRPC do Jaeger.

---

## [1.0.0] — 2025-01-01

### Adicionado

- Pacote `internal/middleware`:
  - `Config` — struct de dependências injetáveis para o middleware RED (contador, histograma, tracer).
  - `RED` — middleware HTTP que implementa a instrumentação de Taxa, Erros e Duração via `http_requests_total` e `http_request_duration_seconds`; envolve cada requisição em um span de servidor OTel e injeta headers W3C `traceparent` / `tracestate` na resposta.
  - `statusRecorder` — wrapper minimalista de `http.ResponseWriter` que captura o código de status escrito pelos handlers.
- Pacote `internal/tracing`:
  - `New` — constrói um `TracerProvider` OTel com exportador OTLP gRPC. Retorna uma função de shutdown para flush limpo de spans.
- Adições em `internal/metrics`:
  - Histograma `process_duration_seconds` — latência de processamento simulada com durações aleatórias.
  - Gauge `go_goroutines_active` — número atual de goroutines ativas no sistema.
  - Counter vec `http_requests_total` — contador RED de taxa/erro rotulado por `method`, `route`, `status_code`.
  - Histogram vec `http_request_duration_seconds` — histograma RED de duração rotulado por `method`, `route`.
- Adições em `cmd/server/main.go`:
  - Middleware RED conectado ao `http.ServeMux` da aplicação.
  - `TracerProvider` OTel inicializado no startup.
- Diretório `deploy/`:
  - `deploy/prometheus/prometheus.yml` — config de scrape do Prometheus.
  - `deploy/grafana/provisioning/` — datasource Prometheus e dashboard auto-provisionados.
- `docker-compose.yml` — orquestra `app`, `prometheus`, `jaeger` e `grafana`.
- `Dockerfile` — build multi-estágio; imagem final roda com usuário não-root.
- `tests/middleware/red_integration_test.go` — testes de integração para o middleware RED.

### Alterado

- `Worker` agora emite um span OTel por ciclo além de incrementar o contador de eventos.
- `Registry.NewRegistry` registra o conjunto completo de métricas RED na construção.
- Target `up` do `Makefile` agora invoca `docker compose up --build`.

---

## [0.2.0] — 2025-01-01

### Segurança

- Atualizado `google.golang.org/protobuf` para resolver vulnerabilidade conhecida de upstream.

---

## [0.1.0] — 2025-01-01

### Adicionado

- Pacote `internal/metrics`:
  - `Registry` — encapsula um `prometheus.Registry` customizado com `app_events_total` registrado.
  - `Worker` — incrementa um contador Prometheus em um intervalo fixo.
- `cmd/server/main.go`:
  - Servidor HTTP da aplicação na porta `8080`.
  - Servidor HTTP de métricas dedicado na porta `9091`.
  - Graceful shutdown com timeout de 10 segundos.
- Pacote `tests/metrics` com testes unitários para o registro e worker.
- `Makefile` com targets: `deps`, `build`, `test`, `lint`, `up`.
