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

Os intervalos usam o formato de duração do Go: `30s`, `90s`, `2m`, `1h`,
`1m30s`. Um valor inválido ou zero volta para o padrão.

> Evite `PR_INTERVAL` abaixo de `30s`. Cada ciclo é uma consulta à API do
> GitHub, e o seu login tem um limite por hora compartilhado com outras
> ferramentas. Veja [funcionamento.md](funcionamento.md#limites).

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
