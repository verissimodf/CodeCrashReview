# Critérios para conclusão do objetivo do projeto

Operadora autorizada: cadastro existente de `contato@laytech.com.br`, vinculado pelo UUID; nunca promover uma conta com base apenas no e-mail recebido em uma requisição.

| Entrega | Prova obrigatória |
|---|---|
| Repositórios Go, Python e JavaScript | Teste real de projeto fora do avaliador da arena, baseline e candidato isolados, testes originais preservados, hashes e saídas persistidos |
| Dependências reproduzíveis | Imagem de validação identificada e preparação limitada a manifests/lockfiles aprovados, sem executar scripts de instalação enviados pelo repositório |
| Dois participantes compatíveis | Catálogo de modelos/perfis, especialidade explícita, inferências independentes e registro do modelo realmente solicitado |
| Avaliação autônoma e humana | Pontuações comparáveis, empate explícito quando não houver diferença, decisão bloqueada se faltar prova, operação humana autenticada |
| Publicar no repositório de origem | Branch/hotfix e tag/release opcionais, revisão de origem conferida, credencial transitória, confirmação do SHA no remoto, recibo persistido |
| Implantar a proposta aprovada | Operadora autorizada, destino configurado, isolamento, verificação funcional do serviço implantado e rollback comprovado |
| Interface navegável | Criar revisão, acompanhar, comparar, decidir e publicar/implantar com resultados reais; teste desktop e móvel |
| Segurança e operação | Isolamento entre usuários, limites de recursos, ausência de segredos em artefatos, recuperação de trabalhos, backup/restauração e trilha de auditoria |
| Entrega final | Testes locais e CI aprovados; imagens e código identificados; documentação sincronizada; Develop e main atualizados |

## Topologia e capacidade inicial

Manter nginx-app, frontend, API Go, PostgreSQL, Redis e FreeLLM-Local. O broker recebe operações tipadas pelo socket Unix e executa código somente em containers sem privilégios. A aplicação não recebe acesso direto ao Docker. Credenciais de publicação não são fornecidas ao modelo nem ao executor de testes.

Operação inicial: uma revisão ativa por worker, duas inferências concorrentes, um executor de prova serializado e limitado por memória/CPU, snapshots limitados a 20 MiB/1000 arquivos. Metadados e fila permanecem no PostgreSQL; Redis atende limitação de requisições. Leituras de estado são pequenas e não justificam cache adicional que possa exibir uma aprovação obsoleta. Escalar primeiro verticalmente; múltiplos workers dependem de leases com exclusão e limites globais.

O host único continua sendo um ponto único de falha. Saúde, reinício, backup e rollback melhoram recuperação, mas não são alta disponibilidade entre hosts. A arquitetura não presume equipamentos ou domínios adicionais que não estejam disponíveis.

## Método

Implementar por contrato, reproduzir falhas antes da correção, executar baseline e candidato com os mesmos testes, registrar evidências do executor, impedir promoção sem prova e validar a publicação no destino. Tempos são observabilidade e limites de proteção de recursos, nunca critério de vitória. Casos sem dependências ou perfil disponível devem falhar de forma explícita, sem aprovação fictícia.
