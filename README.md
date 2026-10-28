# painel

Dashboard pessoal de terminal: digito `painel` e aparece uma tela que fica se
atualizando com o que importa no meu dia.

- **Relógio grande** com saudação, data por extenso e barra do expediente (9h–18h)
- **Revisões pendentes**: PRs abertos que pediram revisão diretamente para mim,
  mais quantos pedidos foram feitos aos meus times
- **Meus PRs abertos**
- **Agenda de hoje** com destaque para a reunião em andamento e contagem para a próxima

## Requisitos

- Go 1.24+
- [`gh`](https://cli.github.com/) autenticado (`gh auth status`): os PRs vêm de `gh search prs`

## Instalação

```sh
make install   # gera ~/.local/bin/painel
```

## Agenda pelo Google Calendar

Para contas Google Workspace em que o "Endereço secreto no formato iCal" foi
desativado pelo admin. É preciso um client OAuth próprio (uma vez só):

1. Em [console.cloud.google.com](https://console.cloud.google.com), crie um
   projeto (ex.: `painel`).
2. **APIs e serviços > Biblioteca**: ative a **Google Calendar API**.
3. **APIs e serviços > Tela de permissão OAuth**: tipo **Interno** (se a conta
   for Workspace) ou **Externo** com seu e-mail como usuário de teste.
4. **APIs e serviços > Credenciais > Criar credenciais > ID do cliente OAuth**,
   tipo **App para computador**. Baixe o JSON e salve em
   `~/.config/painel/google-client.json`.
5. Rode `painel auth` e faça login no navegador.

O token fica em `~/.config/painel/google-token.json` (permissão 600) e só tem
acesso de leitura aos eventos. Com ele presente, o painel usa o Google em vez
do `ICS_URL` (force com `AGENDA=ics` ou `AGENDA=google`).

## Configuração

Arquivo opcional `~/.config/painel/config` (`CHAVE=valor`). Qualquer chave pode
ser sobrescrita por uma variável de ambiente com o prefixo `PAINEL_`
(ex.: `PAINEL_ICS_URL`).

```sh
# Nome na saudação (padrão: primeiro nome do git config user.name)
NAME=Marcus

# Agenda: no Google Calendar, Configurações > seu calendário >
# "Endereço secreto no formato iCal". Trate como senha.
# Também aceita o caminho de um .ics local, mas uma exportação não atualiza sozinha.
ICS_URL=https://calendar.google.com/calendar/ical/.../basic.ics

# Meu e-mail no calendário (só para iCal; o Google já sabe quem sou), para esconder reuniões que recusei
EMAIL=voce@empresa.com

# Intervalos de atualização (padrões: 2m e 5m)
PR_INTERVAL=2m
CAL_INTERVAL=5m
```

## Atalhos

| tecla | ação |
|-------|------|
| `r` | atualiza tudo agora |
| `p` | abre a lista de revisões no navegador |
| `q` / `esc` | sai |

## Desenvolvimento

```sh
make build && ./painel
make test
```
