# Lembretes

Lembretes são avisos que você mesmo marca: "15:30 ligar pro fornecedor". Eles
ficam numa caixa própria na tela e, na hora, o painel mostra um balão e
**toca um alarme**.

## Criando

Na tela, aperte `n`. Um campo abre no rodapé; escreva quando e o quê e aperte
`enter` (`esc` cancela):

| Você escreve | Quando avisa |
|---|---|
| `15:30 ligar pro fornecedor` | Hoje às 15:30 (amanhã, se já passou) |
| `15h café` ou `15h45 café` | Hoje às 15:00 / 15:45 |
| `+20m alongar` | Daqui a 20 minutos (também `+1h`, `+1h30m`, `+10min`) |
| `amanhã 9:00 revisar o PR` | Amanhã às 9:00 |
| `hoje 18h sair no horário` | Hoje às 18:00 (dá erro se já passou) |
| `25/12 10:00 ligar pra família` | Em 25 de dezembro (no ano que vem, se a data já passou) |

Fora da tela, pelo terminal:

```sh
painel lembrar 15:30 ligar pro fornecedor
painel lembrar +20m café
painel lembrar          # lista os pendentes
```

O painel aberto percebe na hora o lembrete criado pelo terminal.

## Na tela

A caixa **Lembretes** mostra os pendentes por horário. Os da próxima hora
aparecem em amarelo, com quanto falta ("15:20 em 20 min"). Os de outros dias
mostram o dia ("amanhã 09:00", "sex 14:00", "25/12 10:00").

Quando chega a hora, o lembrete ganha um `!` em vermelho e **continua na tela**
até você concluir: selecione com as setas e aperte `x`. Se o texto tiver um
link, `enter` abre o link.

## O alarme

Diferente das [notificações](notificacoes.md), os lembretes:

- **Não dependem do `NOTIFY`.** Foi você que pediu o aviso, então ele sempre
  vem.
- **Sempre tocam som**, e o som é tocado à parte do balão, para tocar mesmo
  com o **Não perturbe** ligado. No Linux, o som `alarm-clock-elapsed` do tema
  do sistema (via `canberra-gtk-play` ou, na falta dele, `paplay`); no macOS,
  o "Glass".
- **Avisam ao abrir**, se venceram com o painel fechado. Cada lembrete toca
  uma vez só.

## O arquivo

Os lembretes ficam em `~/.config/painel/lembretes`, um por linha, e dá para
editar à mão:

```
- 2026-10-30 15:30 ligar pro fornecedor
! 2026-10-30 09:00 este já tocou
```

`-` é pendente e `!` já tocou. Concluir com `x` apaga a linha. Linhas fora
desse formato são ignoradas.

## Limites

- O painel só toca **enquanto está aberto**. O que vencer com ele fechado
  toca quando você abrir.
- Não há lembrete que se repete (todo dia, toda segunda). Para isso, a agenda
  é o lugar certo.
