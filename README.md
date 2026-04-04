# goobservabilitystack

> Implementação de referência para observabilidade em Go — métricas RED, tracing OTel, Prometheus, Jaeger e Grafana em uma stack única e composta.

![CI](https://github.com/ESousa97/goobservabilitystack/actions/workflows/ci.yml/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/ESousa97/goobservabilitystack)
![Go Reference](https://pkg.go.dev/badge/github.com/ESousa97/goobservabilitystack.svg)
![License](https://img.shields.io/github/license/ESousa97/goobservabilitystack)
![Go Version](https://img.shields.io/github/go-mod/go-version/ESousa97/goobservabilitystack)
![Last Commit](https://img.shields.io/github/last-commit/ESousa97/goobservabilitystack)

---

`goobservabilitystack` é uma base mínima e correta para instrumentação de serviços HTTP em Go. O projeto demonstra o padrão RED (Rate, Errors, Duration) usando métricas Prometheus e tracing distribuído OpenTelemetry, com separação clara entre as camadas de transporte, lógica de negócio e telemetria. Projetado para baixo overhead de CPU e memória em infraestruturas de alto throughput.

## Demonstração

Inicie a stack completa e verifique o serviço em execução:

```bash
# Health check
curl -i http://localhost:8080/healthz

# Métricas Prometheus
curl -s http://localhost:9091/metrics | grep http_requests_total

# Inspecione um trace no Jaeger
# Acesse http://localhost:16686 no seu navegador
```

![Preview do dashboard Grafana](assets/dashboard_preview.png)

## Stack Tecnológico

| Componente | Papel |
|---|---|
| Go 1.25+ | Runtime da aplicação |
| [prometheus/client_golang](https://github.com/prometheus/client_golang) | Registro de métricas customizadas, contadores RED e histogramas |
| [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) | Tracing distribuído com propagação de contexto W3C |
| [Jaeger](https://www.jaegertracing.io/) | Backend de tracing (receptor OTLP gRPC) |
| [Prometheus](https://prometheus.io/) | Backend de métricas e scraper |
| [Grafana](https://grafana.com/) | Frontend de dashboard com painéis pré-provisionados |
| Docker Compose | Orquestração local da stack completa de observabilidade |

## Pré-requisitos

- Go >= 1.25
- Docker e Docker Compose (para a stack completa)
- `golangci-lint`: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `govulncheck`: `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Instalação e Uso

### Como binário

```bash
go install github.com/ESousa97/goobservabilitystack/cmd/server@latest
```

### A partir do source

```bash
git clone https://github.com/ESousa97/goobservabilitystack.git
cd goobservabilitystack
make deps
make build
./bin/goobservabilitystack
```

O servidor da aplicação inicia em `http://localhost:8080` e o endpoint de métricas em `http://localhost:9091/metrics`.

### Stack completa de observabilidade

```bash
make up
```

Este comando compila o binário Go e inicia Prometheus, Jaeger e Grafana via Docker Compose.

| Serviço | URL |
|---|---|
| Aplicação | http://localhost:8080 |
| Métricas | http://localhost:9091/metrics |
| Prometheus | http://localhost:9090 |
| Jaeger UI | http://localhost:16686 |
| Grafana | http://localhost:3000 (admin / admin) |

## Makefile Targets

| Target | Descrição |
|---|---|
| `make deps` | Baixa e organiza as dependências do módulo Go |
| `make build` | Compila o binário do servidor para `bin/goobservabilitystack` |
| `make test` | Executa todos os testes com o race detector ativado |
| `make lint` | Executa `golangci-lint` em todos os pacotes |
| `make up` | Compila e inicia a stack completa com Docker Compose |

## Arquitetura

O projeto segue uma estrutura modular focada em separação de preocupações:

- `cmd/server/main.go`: Ponto de entrada; orquestra dependências, inicia servidores HTTP e gerencia o graceful shutdown.
- `internal/metrics/`: Registro customizado Prometheus e definições de métricas (sem estado global).
- `internal/middleware/`: Middleware HTTP RED; envolve handlers com métricas de taxa/erro/duração e spans OTel.
- `internal/tracing/`: Provedor OTel Tracer; configurado para exportação via gRPC para o Jaeger.
- `deploy/`: Configurações de infraestrutura (Prometheus e provisionamento do Grafana).

**Decisões arquiteturais:**

- Isolamento de métricas via `prometheus.NewRegistry()` para evitar poluição global.
- Gerenciamento de ciclo de vida de goroutines via `context.Context`.
- Servidor de métricas em porta dedicada para isolar tráfego de monitoramento.
- Uso estrito de timeouts em servidores HTTP para resiliência de recursos.

## API Reference

A documentação detalhada dos pacotes está disponível em [pkg.go.dev](https://pkg.go.dev/github.com/ESousa97/goobservabilitystack).

| Método | Caminho | Porta | Descrição |
|---|---|---|---|
| `GET` | `/healthz` | `8080` | Health check — retorna `200 OK` |
| `GET` | `/metrics` | `9091` | Exposição de métricas no formato Prometheus |

## Configuração

| Variável | Tipo | Padrão | Descrição |
|---|---|---|---|
| `HTTP_PORT` | string | `8080` | Porta do servidor HTTP da aplicação |
| `METRICS_PORT` | string | `9091` | Porta de scrape do Prometheus `/metrics` |
| `JAEGER_ENDPOINT` | string | `localhost:4317` | Endpoint gRPC do coletor Jaeger |
| `SERVICE_NAME` | string | `goobservabilitystack` | Nome do serviço para atributos OpenTelemetry |

## Roadmap

- [x] Registro Prometheus customizado com métricas RED
- [x] Tracing OpenTelemetry com exportador Jaeger gRPC
- [x] Middleware HTTP RED com propagação de contexto W3C
- [x] Stack Docker Compose com dashboard Grafana pré-provisionado
- [ ] Pipeline CI no GitHub Actions (build, test, lint, govulncheck)
- [ ] Taxa de amostragem OTel configurável via variável de ambiente
- [ ] Imagem Docker multi-arquitetura publicada no ghcr.io
- [ ] Suporte a mTLS no exportador gRPC do Jaeger

## Contribuindo

Link para o [CONTRIBUTING.md](CONTRIBUTING.md) com detalhes sobre como configurar o ambiente e abrir pull requests.

## Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.

## Autor

Enoque Sousa — [enoquesousa.vercel.app](https://enoquesousa.vercel.app) · [github.com/ESousa97](https://github.com/ESousa97)
