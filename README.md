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

# Meu e-mail no calendário, para esconder reuniões que recusei
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
