# Revisão e entrega de projetos

## Escopo de execução

Perfis: `go-test-v1`, `python-unittest-v1`, `node-test-v1` e o perfil especializado `arena-grade-v1`. O projeto precisa conter testes originais. O executor preserva esses testes, executa o estado original e o candidato e registra os hashes dos arquivos, da suíte, da imagem e as saídas. Aprovação não comprova requisitos ausentes da suíte; cobertura e adequação dos testes continuam parte da avaliação.

O snapshot executável aceita até 1000 arquivos, 20 MiB no total e 1 MiB por arquivo. O contexto enviado ao modelo tem limites próprios configuráveis. Não são enviados segredos conhecidos, diretórios de dependências nem links simbólicos. Código executa sem rede, com usuário sem privilégios, sistema de arquivos somente leitura, memória, CPU, processos e duração técnica limitados. Não há tempo mínimo nem prêmio por velocidade. O limite do processo de teste protege o host e não determina a vitória.

## Dependências

Imagens base atendem projetos que usam bibliotecas padrão. Para dependências externas, o operador usa `scripts/prepare-validator-image.py` com perfil, diretório dos manifests e `--base` fixada por digest. A preparação recebe somente manifests e lockfiles, nunca o código do repositório. Go baixa módulos pelo proxy e verifica sums; Node exige integridade e URLs do registro oficial, desabilita scripts; Python exige versões exatas, hashes e wheels. Referências locais, Git e instalações executáveis não são aceitas automaticamente.

O script retorna o ID imutável da imagem. Configurar `VALIDATOR_IMAGE`, `VALIDATOR_NODE_IMAGE` ou `VALIDATOR_PYTHON_IMAGE` e recriar somente o broker. A inferência não pode escolher imagens, comandos de teste nem instalar dependências no host. O executor de provas permanece sem rede. Atualizações de dependências exigem reconstrução e nova prova.

## Publicação

Publicar exige ser proprietário da revisão concluída, selecionar a proposta aprovada e reproduzir seu hash. A branch de origem deve continuar na revisão analisada. O token GitHub é transitório e não é persistido nem enviado ao modelo. Branches novas devem começar com `review/`, `hotfix/` ou `release/`; tags e releases são opcionais. Referências com outro commit nunca são sobrescritas. O recibo registra origem, snapshot, proposta, commit e destino. Retentativas idênticas reutilizam a operação reservada.

## Implantação

Apenas UUIDs em `REVIEW_OPERATOR_IDS` podem implantar. O arquivo do host `~/.local/share/codecrashreview/deploy-targets.json` define os destinos; uma instalação nova começa com a lista vazia de `deploy/targets.example.json`. A aplicação e o broker recebem esse arquivo somente para leitura. E-mail enviado pelo cliente não concede privilégio.

A receita disponível inicia um serviço Python HTTP na porta interna 8080, com entrada, origem GitHub, porta de loopback e resposta funcional previamente configuradas pelo operador. A imagem deve ser a mesma da prova. Código, imagem, comandos, porta e mounts não ficam à escolha do modelo. O serviço é isolado em rede interna dedicada, sem privilégios e sem segredos do host. A URL de loopback é acessível no próprio host; publicação externa do serviço exige configuração própria de domínio, independente do domínio da arena.

A implantação para o serviço anterior, inicia o candidato e verifica a resposta configurada. Falha remove o candidato e reinicia/verifica o serviço anterior. A versão anterior aprovada é mantida. Não há migração automática de dados: a receita atende serviços sem estado. Uma implantação interrompida sem recibo confirmado deve ser inspecionada antes de nova tentativa; não se presume sucesso.

## Limites operacionais

Um host não equivale a alta disponibilidade entre máquinas. A execução de testes é serializada pelo broker; as duas inferências são independentes e concorrentes. Perfis de especialista representam instruções diferentes, não certificações. Os modelos pedidos e retornados pelo provedor devem ser registrados sem presumir que um alias de roteamento identifique o modelo físico subjacente.
