# CodeCrashReview
<!-- impeccable:product-schema 1 -->

## Platform
web

## Product Purpose
Aplicativo para revisar e corrigir repositórios locais autorizados ou GitHub com dois agentes independentes e concorrentes. Um avaliador autônomo ou operador humano compara as propostas e suas provas. A solução vencedora pode ser publicada no repositório de origem ou implantada em destino autorizado.

## Users
Operador responsável pelos repositórios e pela aprovação das correções. A conta contato@laytech.com.br foi indicada pelo usuário para operar implantações; a autorização deve estar vinculada à identidade persistida.

## Operating Context
Infraestrutura Docker existente, nginx-app, PostgreSQL e inferência preferencial pelo FreeLLM-Local. Acesso web pelo domínio codecrashreview.laytech.com.br. Código de terceiros é conteúdo não confiável.

## Capabilities and Constraints
A proposta deve apresentar topologia, design, plano de implantação, critérios de aprovação, objetivo alcançável, metodologia, evidência e passos executados. Competidores devem ter especialidades compatíveis. A prova funcional, e não o tempo decorrido, determina a conclusão. Recursos ainda não validados não podem aparecer como operacionais.

## Brand Commitments
Interface limpa, polida, moderna, em português. Nome CodeCrashReview.

## Product Principles
- Prova executada se distingue de alegação do modelo.
- Aprovação e publicação são rastreáveis.
- Testes e isolamento protegem os repositórios e o host.
- Resultados, falhas e limites são visíveis.

## Evidence on Hand
Relatórios versionados em docs/validation e execuções locais em artifacts/validation. A documentação da versão 0.2.0 registra seu escopo limitado; a entrega ampliada exige nova validação.

## Interface confirmada
O usuário escolheu uma área de trabalho com lista de revisões e comparação lado a lado.
