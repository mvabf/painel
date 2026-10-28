# Configuração

O painel funciona sem nenhuma configuração: os PRs vêm do `gh` e o nome da
saudação vem do Git. A configuração serve para ligar a agenda e ajustar
detalhes.

## Onde fica

```
~/.config/painel/config
```

O formato é uma chave por linha, `CHAVE=valor`. Linhas em branco e linhas que
começam com `#` são ignoradas. Aspas em volta do valor são opcionais.

```sh
# ~/.config/painel/config
NAME=Ana
ICS_URL="https://calendar.google.com/calendar/ical/.../private-.../basic.ics"
EMAIL=ana@empresa.com
```

## Variáveis de ambiente

Toda chave pode ser definida também como variável de ambiente, com o prefixo
`PAINEL_`. **A variável tem prioridade sobre o arquivo.** Isso é útil para
testar algo sem editar o config:

```sh
PAINEL_PR_INTERVAL=30s painel
PAINEL_AGENDA=ics PAINEL_ICS_URL=~/Downloads/teste.ics painel
```

## Referência

| Chave | Padrão | Descrição |
|---|---|---|
| `NAME` | primeiro nome de `git config --global user.name` | Nome na saudação. Se não houver nenhum, a saudação sai sem nome |
| `ICS_URL` | (vazio) | Link `http(s)://` de um calendário iCal ou caminho de um arquivo `.ics`. Aceita `~/` no começo. Veja [agenda.md](agenda.md) |
| `EMAIL` | (vazio) | Seu e-mail no calendário. **Só para iCal:** esconde os eventos que você recusou. O Google já sabe quem você é |
| `AGENDA` | automático | Força a fonte da agenda: `google` ou `ics`. Sem essa chave, usa o Google se houver login e o iCal se houver `ICS_URL` |
| `PR_INTERVAL` | `2m` | De quanto em quanto tempo buscar os PRs |
| `CAL_INTERVAL` | `5m` | De quanto em quanto tempo buscar a agenda |
| `ENV_FILE` | (vazio) | Outro arquivo `CHAVE=valor` de onde ler os nomes padrão do Jira (veja abaixo). Aceita `~/` |
| `JIRA_URL` | `JIRA_BASE_URL` | Endereço do seu Jira Cloud, ex.: `https://empresa.atlassian.net` |
| `JIRA_EMAIL` | `JIRA_EMAIL` | E-mail da sua conta Atlassian |
| `JIRA_TOKEN` | `JIRA_API_TOKEN` | Token de API da Atlassian |
| `JIRA_JQL` | `assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC` | Busca mostrada na caixa |
| `JIRA_INTERVAL` | `5m` | De quanto em quanto tempo buscar o Jira |

Nas chaves do Jira, a coluna "Padrão" indica o nome padrão usado como
alternativa (veja [Jira](#jira)). A caixa do Jira só aparece com URL, e-mail e
token preenchidos.

Os intervalos usam o formato de duração do Go: `30s`, `90s`, `2m`, `1h`,
`1m30s`. Um valor inválido ou zero volta para o padrão.

> Evite `PR_INTERVAL` abaixo de `30s`. Cada ciclo é uma consulta à API do
> GitHub, e o seu login tem um limite por hora compartilhado com outras
> ferramentas. Veja [funcionamento.md](funcionamento.md#limites).

## Jira

O painel usa o Jira Cloud com autenticação básica (e-mail + token de API). Crie
o token em
[id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens).
Ele tem as mesmas permissões da sua conta, mas o painel só faz buscas.

Cada valor é procurado nesta ordem:

1. variável de ambiente `PAINEL_JIRA_URL` / `PAINEL_JIRA_EMAIL` / `PAINEL_JIRA_TOKEN`;
2. chave `JIRA_URL` / `JIRA_EMAIL` / `JIRA_TOKEN` no `config`;
3. variável de ambiente com o nome padrão `JIRA_BASE_URL` / `JIRA_EMAIL` / `JIRA_API_TOKEN`;
4. o mesmo nome padrão dentro do arquivo indicado em `ENV_FILE`.

Os itens 3 e 4 existem para você **não copiar o token** para o config se ele já
estiver em outro lugar (outras ferramentas usam os mesmos nomes):

```sh
# ~/.config/painel/config
ENV_FILE=~/.secrets/tokens.env
```

```sh
# ~/.secrets/tokens.env
TOKEN_JIRA=abc123...
export JIRA_BASE_URL=https://empresa.atlassian.net
JIRA_EMAIL=ana@empresa.com
JIRA_API_TOKEN=${TOKEN_JIRA}
```

### Exemplos de `JIRA_JQL`

```sh
# Só um projeto (atenção: alguns nomes, como INT, são palavras reservadas e precisam de aspas)
JIRA_JQL=project = "INT" AND assignee = currentUser() AND statusCategory != Done

# O que está na sprint ativa
JIRA_JQL=assignee = currentUser() AND sprint in openSprints() ORDER BY status

# Bugs de alta prioridade do time
JIRA_JQL=project = "APP" AND type = Bug AND priority in (Highest, High) AND statusCategory != Done
```

## Formato dos arquivos

O `config` e o arquivo de `ENV_FILE` seguem as mesmas regras:

- uma chave por linha, `CHAVE=valor`; `export ` no começo é aceito;
- linhas vazias e linhas que começam com `#` são ignoradas;
- aspas em volta do valor inteiro são removidas, e aspas no meio ficam
  (`JIRA_JQL=project = "INT"` funciona);
- `$VAR` e `${VAR}` são substituídos pelo valor de uma linha anterior do mesmo
  arquivo ou do ambiente. Entre aspas simples (`'$VAR'`), o texto fica literal.

## Outros arquivos da pasta

Além do `config`, o painel usa estes arquivos em `~/.config/painel/`:

| Arquivo | Quem cria | Para quê |
|---|---|---|
| `google-client.json` | você, baixando do Google Cloud | Identifica o "app" no login do Google |
| `google-token.json` | `painel auth` | Seu login no Google Calendar |

Os dois são credenciais. Veja [seguranca.md](seguranca.md).

## O que não é configurável (ainda)

- O horário do expediente na barra de progresso é fixo em **9h–18h**, de
  segunda a sexta.
- O idioma da interface é português.
- O número de itens por caixa: até 6 PRs e 8 eventos.

Se precisar de algum desses, abra uma issue.
