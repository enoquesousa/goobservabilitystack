# Política de Segurança

## Versões Suportadas

| Versão | Suportado |
| ------- | ------------------ |
| latest (master) | :white_check_mark: |

## Reportando uma Vulnerabilidade

Se você descobrir uma vulnerabilidade de segurança no `goobservabilitystack`, reporte-a de forma responsável abrindo um [GitHub Security Advisory](https://github.com/ESousa97/goobservabilitystack/security/advisories/new) em vez de uma issue pública.

Inclua no seu relatório:

- Uma descrição da vulnerabilidade e seu impacto potencial.
- Passos para reproduzir ou uma prova de conceito (PoC) mínima.
- Versões ou commits afetados, se conhecidos.

Você pode esperar um reconhecimento dentro de **5 dias úteis** e um plano de resolução ou mitigação em até **30 dias**, dependendo da complexidade.

## Considerações de Segurança

### Exposição do endpoint de métricas

O endpoint `/metrics` (porta padrão `9091`) expõe dados internos do runtime — contagem de goroutines, taxas de requisição, taxas de erro e histogramas de latência. Este endpoint não deve ser acessível a partir de redes não confiáveis.

Mitigações recomendadas:

- Vincule o servidor de métricas a uma interface não pública (ex: `127.0.0.1` ou uma rede interna).
- Coloque o endpoint de métricas atrás de uma regra de firewall ou política de rede que permita acesso apenas do seu scraper Prometheus.
- No Kubernetes, use uma `NetworkPolicy` que restrinja a porta `9091` ao namespace de monitoramento.

### Servidor HTTP da Aplicação

O servidor da aplicação (porta padrão `8080`) serve `/healthz` e futuras rotas. Se exposto externamente:

- Implante atrás de um proxy reverso ou balanceador de carga com terminação TLS.
- Não exponha a porta HTTP bruta diretamente à internet.

### Variáveis de Ambiente

Configurações sensíveis são passadas via variáveis de ambiente (`HTTP_PORT`, `METRICS_PORT`, `JAEGER_ENDPOINT`, `SERVICE_NAME`). Ao implantar com Docker Compose ou Kubernetes:

- Não comite arquivos `.env` contendo valores reais no controle de versão.
- Use objetos `Secret` do Kubernetes ou um gerenciador de segredos (Vault, AWS Secrets Manager, etc.) para credenciais em vez de entradas simples de `ConfigMap`.

### Jaeger / Exportador OTel

Dados de trace são exportados via gRPC para o `JAEGER_ENDPOINT` (padrão `localhost:4317`). Em produção:

- Ative mTLS no endpoint do coletor Jaeger.
- Restrinja o acesso à rede para que apenas a aplicação possa alcançar a porta `4317`.
- Traces podem conter metadados de requisição (rotas, status codes); evite incluir informações de identificação pessoal (PII) em atributos de span.

### Gerenciamento de Dependências

Dependências são fixadas no `go.sum` e gerenciadas via módulos Go. Para auditar vulnerabilidades conhecidas:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

As atualizações de dependências (incluindo patches de segurança) são automatizadas via Dependabot.

### Higiene da Imagem Docker

Ao construir uma imagem de container:

- Use uma imagem base mínima (ex: `gcr.io/distroless/static` ou `scratch`) para reduzir a superfície de ataque.
- Execute o processo como um usuário não root.
- Não copie o toolchain do Go ou código-fonte para o estágio final da imagem.
