# Protótipos anteriores, não publicados

Os pacotes `backend/internal/agents`, `backend/internal/sandbox` e `backend/pkg/docker` não eram dependências do executável `cmd/api`. Foram preservados aqui fora do módulo ativo durante a consolidação da arena de revisão. Incluíam simuladores e integrações Docker genéricas sem uso no fluxo operacional.

O validador atual usa um contrato restrito em `deploy/validator`, com testes de aceitação mantidos pelo projeto. Não reincorporar estes protótipos sem revisão, atualização de dependências e testes próprios.
