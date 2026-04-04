# Contribuindo

Obrigado pelo seu interesse em contribuir para o `goobservabilitystack`.

## Pré-requisitos

- Go 1.25 ou posterior
- Docker e Docker Compose
- `golangci-lint` — instalar via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `govulncheck` — instalar via `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Fluxo de Trabalho

1. Faça um fork do repositório e crie uma branch a partir da `master`.
2. Realize suas alterações.
3. Execute a suite de testes localmente antes de abrir um pull request (veja abaixo).
4. Abra um pull request contra a `master` com uma descrição clara do que a mudança faz e por quê.

## Verificações Locais

Todas as seguintes verificações devem passar antes que um pull request seja aberto:

```bash
# Instalar / organizar dependências
make deps

# Build
make build

# Testes (com race detector ativado)
make test

# Lint
make lint

# Auditoria de vulnerabilidades
govulncheck ./...
```

## Padrões de Código

### Geral

- Escreva código Go idiomático. Siga o estilo descrito em [Effective Go](https://go.dev/doc/effective_go) e [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Mantenha as alterações focadas. Um pull request deve fazer apenas uma coisa.
- Não introduza estado global. Passe registros, contadores e outras dependências explicitamente via construtores.

### Documentação

- Toda função, tipo, constante e variável exportada **DEVE** ter um comentário Godoc.
- Os comentários devem ser frases completas começando com o nome do item documentado.
- Use links de documentação nativos do Go 1.19+ (ex: `[Registry]`, `[NewWorker]`) ao referenciar outros símbolos exportados.

### Métricas Prometheus

- Nomes de métricas devem ser em `snake_case`.
- Contadores devem ter o sufixo `_total`.
- Métricas que medem tempo devem usar `_seconds`; métricas que medem bytes devem usar `_bytes`.
- Registre métricas no `*prometheus.Registry` customizado retornado por `metrics.NewRegistry()`, nunca no registro global padrão.
- Novos tipos de métricas devem ser cobertos por testes usando `prometheus/testutil`.

### Testes

- Coloque os testes em `tests/<package>/` como um pacote externo `_test` (ex: `package metrics_test`).
- Ative o race detector: `go test -race ./...`.
- Não use mocks para o registro Prometheus ou contadores em testes — use instâncias reais de `prometheus.NewCounter` / `prometheus.NewRegistry`.
- Testes baseados em tabelas (table-driven tests) são preferidos para cobertura de caminhos de erro.

### Mensagens de Commit

Use o formato [Conventional Commits](https://www.conventionalcommits.org/):

```
<tipo>(<escopo>): <descrição curta>

[corpo opcional]
```

Tipos comuns: `feat`, `fix`, `test`, `refactor`, `build`, `docs`, `chore`.

Exemplos:

```
feat(metrics): adiciona contador RED http_requests_total
fix(worker): para o ticker antes de retornar de Run
test(registry): adiciona testes table-driven para erros de registro
docs: atualiza README com tabela de serviços da stack
```

## Diretrizes de Pull Request

- Alvo: `master`.
- Mantenha os pull requests pequenos e fáceis de revisar. Divida mudanças não relacionadas em PRs separados.
- Linke quaisquer issues relevantes do GitHub na descrição do PR.
- Não faça force-push em uma branch de PR após o início da revisão.

## Reportando Bugs

Abra uma [issue no GitHub](https://github.com/ESousa97/goobservabilitystack/issues) com:

- Um título claro.
- Passos para reproduzir.
- Comportamento esperado vs. real.
- Versão do Go (`go version`) e SO.

Para vulnerabilidades de segurança, siga o processo em [SECURITY.md](SECURITY.md) em vez de abrir uma issue pública.
