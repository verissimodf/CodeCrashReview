# Implantação no host app

O arquivo `compose.host.yml` usa PostgreSQL e Redis em rede interna. O frontend tem o alias `codecrash-frontend` na rede externa `CNC_APP_PROD_01`. A aplicação usa o Nginx e o serviço de inferência já disponíveis no host. Configure as credenciais no arquivo local `.env` com permissão 0600.

O cliente Go acessa o FreeLLM-Local em `127.0.0.1:8082/v1` por um socket Unix privado. As duas unidades em `deploy/codecrash-freellm.*` pertencem ao gerenciador systemd do usuário laytech. O Compose monta apenas o diretório do socket no backend. O usuário da imagem precisa corresponder ao proprietário do socket (UID 1000 neste host).

Use `DEFAULT_MODEL=claude-3-freellm-no-thinking/ollama/qwen2.5:7b` para chamadas de ferramentas e a credencial do proxy em `FREELLM_API_KEY`. O modelo já está no Ollama do host. Para geração de texto, a rota equivalente com `deepseek-coder:6.7b` também respondeu.

## Estado verificado em 18/11/2026

PostgreSQL e Redis estão implantados. O PostgreSQL recebeu a migração Goose 1 e contém as seis personas iniciais. O cliente Go recebeu uma chamada de ferramenta estruturada por meio do socket a partir de um contêiner sem rede IP. O teste verificou nome e argumentos da chamada. A arena ainda depende de correções de compilação e da implementação do fluxo de submissão de soluções. A rota no Nginx compartilhado aguarda um frontend e um backend funcionais.

Para verificar os serviços instalados, execute `docker compose -f compose.host.yml ps` na raiz do repositório. Para iniciar apenas os serviços de dados, use `docker compose -f compose.host.yml up -d --wait postgres redis`.


## Arena implantada

O Compose inicia frontend estático, backend Go, PostgreSQL e Redis. O Nginx compartilhado encaminha o hostname HTTP codecrashreview.laytech.com.br para codecrash-frontend. Para acessar na rede, resolva codecrashreview.laytech.com.br para o endereço do host app.

O catálogo atual contém três desafios aritméticos com expressões determinísticas. O avaliador aceita a, b, inteiros, soma, subtração, multiplicação, parênteses e abs/min/max. Uma função solve(a,b) com retorno único também é aceita. O catálogo informa essa sintaxe. Há partidas entre duas contas e duelos assistidos por duas personas no FreeLLM-Local. A avaliação grava submissões, pontuação, vencedor e, nas partidas ranqueadas, histórico Elo em transação.

A execução de repositórios arbitrários e testes multilinguagem ainda exige desenvolvimento adicional.

## Evidência da partida validada

Partida b28bf624-9fb3-4f84-a9a4-ba163a6e62fc criada e disputada pela API publicada em 18/11/2026. Duas contas, ingresso, duas submissões e consulta de resultado passaram. Estado completed, pontuações 100 e 66. Consulta SQL confirmou ratings 1216 e 1184, variações +16/-16 e dois registros em rating_history. A rota /arena respondeu HTTP 200. Os quatro serviços passaram nos healthchecks do Compose.

## Acesso pelo host

O Nginx compartilhado atende a arena em http://codecrashreview.laytech.com.br/arena na LAN e http://codecrashreview.laytech.com.br/arena pelo Tailscale. Os endereços constam no server_name do bloco codecrash-arena em /opt/laytech-ai/volumes/nginx/nginx.conf. O hostname codecrashreview.laytech.com.br requer resolução DNS ou hosts. As rotas usam a mesma origem. Configuração testada e Nginx recarregado em 18/11/2026.

## Revisão de produção em 18/11/2026

Backup diário configurado pelo timer de usuário codecrash-backup.timer às 03h15 no fuso do host. Arquivos com permissão 0600 em /home/laytech/.local/share/codecrashreview/backups. O backup de 23:12:50 UTC foi restaurado em banco temporário com quatro submissões recuperadas. O banco temporário foi removido após a conferência. A gravação em /srv/backup/docker foi impedida por permissão, portanto a cópia atual reside no disco local.

Logs Docker limitados a três arquivos de 10 MB por serviço. Frontend resolve o backend periodicamente pelo DNS Docker para suportar recriação dos serviços. Publicação configurada em HTTP para codecrashreview.laytech.com.br, com DNS ainda pendente de definição do destino. HTTPS continua pendente.

O avaliador permanece limitado aos três desafios aritméticos descritos no catálogo. A execução de repositórios e testes multilinguagem ainda está pendente.


## Publicação pública confirmada em 18/11/2026

URL https://codecrashreview.laytech.com.br/arena. DNS público resolvido e entrada HTTPS pelo Cloudflare Tunnel existente, com encaminhamento ao nginx-app. HTTP público redireciona para HTTPS. O navegador carregou a arena sem erros JavaScript capturados.

Teste pela API HTTPS pública criou duas contas, criou e ingressou na partida e enviou duas soluções. Partida e7aa9ebd-8749-4f35-b199-c483c87cee84 encerrada em completed. Consulta direta no PostgreSQL confirmou pontuações 100 e 66, ratings 1216 e 1184 e variações +16/-16. As contas verify4f1484882e e verifyadb5beab3e são contas de verificação.

A pendência anterior de DNS/HTTPS foi resolvida e substituída por esta validação. O certificado público é servido pelo Cloudflare. A rota entre o túnel e o proxy local utiliza HTTP. Permanece o escopo funcional de três desafios aritméticos, com avaliação restrita documentada acima.
