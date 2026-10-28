# Como funciona

## De onde vêm os PRs

O painel não guarda token do GitHub. Ele chama o **GitHub CLI** (`gh`), que já
está autenticado na sua máquina, e faz **uma única consulta** à API GraphQL do
GitHub com três buscas (veja `prsQuery` em `prs.go`):

| Busca | Equivalente no site | Vira |
|---|---|---|
| `user-review-requested:@me` | PRs que pediram revisão **para você** | Número principal de **Revisões pendentes** e a lista |
| `review-requested:@me` | O mesmo, **incluindo** pedidos aos seus times | Usado para calcular a linha "+N pedidos aos seus times" |
| `author:@me` | Seus PRs | **Meus PRs abertos** |

Todas filtram só PRs abertos, em qualquer repositório que você consiga ver. Os
totais vêm do GitHub e são exatos, mesmo quando a lista é maior do que cabe na
tela.

### Por que separar "para você" de "para os times"

No GitHub, quando alguém pede revisão a um time (ex.: `@empresa/backend`), o PR
aparece na busca `review-requested` de **todos** os membros. Em times grandes
isso infla o número: dezenas de PRs que outra pessoa do time vai revisar. Por
isso o número em destaque é só o que pediram a você pelo nome, e o volume dos
times fica como informação secundária:

```
14 Revisões pendentes
+52 pedidos aos seus times
```

O "+52" é a diferença entre as duas buscas. Se você já revisou um PR, o GitHub
tira o pedido e ele sai das duas contagens.

### Há quanto tempo cada revisão espera

Em **Revisões pendentes**, cada PR começa com o tempo desde que pediram a sua
revisão, e quem espera há mais tempo fica no topo:

```
 4d checkout#482 Corrige arredondamento do frete grátis
30h api-pedidos#1290 Adiciona retry no webhook de pagamento
 2h app-mobile#77 Nova tela de rastreio
```

| Cor | Esperando há |
|---|---|
| cinza | menos de 1 dia |
| amarelo | de 1 a 3 dias |
| vermelho | 3 dias ou mais |

A data é a do último pedido de revisão feito **a você**. Se alguém pedir de
novo depois de um ajuste, a contagem recomeça. Um push novo no PR não muda
nada. Se o pedido veio por um time, o painel usa o último pedido de revisão do
PR e, se nem isso existir, a última atualização.

Pedidos muito antigos (de PRs que ninguém fechou) também aparecem no topo e
entram na contagem. Há duas saídas:

- **Limpar na origem:** remova-se dos revisores no GitHub ou feche o PR.
- **Esconder no painel:** com `REVIEW_MAX_AGE=7d` no config, só aparecem os
  pedidos dos últimos 7 dias. A contagem passa a considerar só esses, e uma
  linha avisa o resto:

  ```
  5 Revisões pendentes
  +52 pedidos aos seus times · ocultos: 9 com mais de 7 dias
  ```

  O filtro olha os 50 pedidos atualizados mais recentemente, que é o que a
  consulta traz.

### Status dos seus PRs

Cada PR em **Meus PRs abertos** começa com o estado do CI do último commit:

| Ícone | Significado |
|---|---|
| ✓ verde | Todos os checks passaram |
| ✗ vermelho | Algum check falhou |
| ● amarelo | Checks rodando ou esperando |
| · cinza | O repositório não tem checks |

Depois do `repo#número`, podem aparecer marcadores:

- **conflito** (vermelho): o PR tem conflito com a branch de destino.
- **mudanças** (vermelho): alguém pediu mudanças na revisão.
- **aprovado** (verde): a revisão exigida foi aprovada.

Abaixo do título da caixa vem um resumo do que pede atenção, contando também os
PRs que não couberam na lista:

```
8 Meus PRs abertos
1 CI falhando · 4 aprovados
```

Rascunhos não entram no resumo e não mostram status.

> O GitHub calcula o "tem conflito" sob demanda. Logo depois de um push, o
> estado pode estar como desconhecido, e o marcador só aparece em um ciclo
> seguinte.

### Ordem e rascunhos

**Meus PRs abertos** vêm ordenados pela última atualização (a mais recente no
topo). **Revisões pendentes**, pelo tempo de espera (o mais antigo no topo).
Rascunhos (draft) aparecem apagados, com `[draft]` antes do título. Cada caixa
mostra até 6 PRs e depois "+N mais".

## Agenda

Veja [agenda.md](agenda.md) para as fontes. A janela é sempre **hoje, da
meia-noite à meia-noite**, no fuso do computador.

Destaques na caixa e no cabeçalho:

- **Já passou** (o horário de fim ficou para trás): texto riscado e apagado.
- **Acontecendo agora:** amarelo e em negrito. O cabeçalho mostra `agora: <título>`.
- **Próxima:** quando nada está acontecendo, o cabeçalho mostra
  `próxima: <título> em 25 min`.
- **Dia inteiro:** vem primeiro, com "dia todo", e não conta para "agora" e
  "próxima".

## Jira

Com o Jira configurado ([configuracao.md](configuracao.md#jira)), uma quarta
caixa mostra o resultado da `JIRA_JQL`. Por padrão, os cards atribuídos a você
e não concluídos:

```
10 Jira
7 em andamento · 3 a fazer

APP-24 Em revisão Corrige cálculo do frete
APP-31 Fazendo Cache na cotação
APP-40 Tela de rastreio
```

- Os **em andamento** vêm primeiro, com a chave em amarelo e o nome do status.
  Os **a fazer** vêm depois, com a chave em cinza. Dentro de cada grupo, vale a
  ordem da JQL.
- O painel busca até 50 cards. Se houver mais, o número vira `50+`.
- A tecla `j` abre a mesma busca no Jira, no navegador.

## Quando cada coisa atualiza

| O quê | Quando |
|---|---|
| Relógio, "agora/próxima", destaques da agenda, barra do expediente | A cada segundo, sincronizado com o relógio do sistema |
| PRs | Ao abrir e depois a cada `PR_INTERVAL` (padrão 2 min) |
| Agenda | Ao abrir e depois a cada `CAL_INTERVAL` (padrão 5 min) |
| Jira | Ao abrir e depois a cada `JIRA_INTERVAL` (padrão 5 min) |
| Tudo | Na hora, ao apertar `r` |

O próximo ciclo só começa quando a busca anterior termina. Por isso, uma rede
lenta não empilha requisições. Apertar `r` cancela o ciclo agendado e começa um
novo a partir dali.

Durante uma busca, um pequeno indicador animado (`⠋`) aparece ao lado do título
da caixa. O rodapé mostra a hora da última atualização bem-sucedida de cada fonte.

## Erros

Se uma busca falha (rede fora, `gh` deslogado, API fora do ar), o painel
**mantém os últimos números na tela** e mostra o erro em vermelho dentro da
caixa. No ciclo seguinte ele tenta de novo, e o erro some quando der certo.
Antes da primeira busca com sucesso, a contagem aparece como `…`.

## Barra do expediente

Vai de 0% às 9h até 100% às 18h, de segunda a sexta, no horário local. Antes
das 9h fica em 0%, depois das 18h em 100%. No sábado e no domingo, a barra dá
lugar a "fim de semana 🎉".

## Layout

- Cada caixa precisa de pelo menos 34 colunas. O painel coloca lado a lado
  quantas couberem e equilibra as linhas. Com o Jira, por exemplo, 4 caixas
  que não cabem numa linha viram 2×2 em vez de 3+1.
- Sem o Jira (3 caixas): uma linha a partir de ~110 colunas e duas linhas a
  partir de ~74. Com o Jira (4 caixas): uma linha a partir de ~146 colunas.
- Num terminal mais estreito que isso, as caixas ficam empilhadas.
- As caixas de uma mesma linha têm a mesma altura.
- Títulos longos são cortados com `…` para não quebrar o layout.
- O painel usa a tela alternativa do terminal: ao sair, seu terminal volta
  exatamente como estava.

## Limites

- **API do GitHub:** cada ciclo custa 2 pontos do limite da API GraphQL, que é
  de 5.000 pontos por hora. Com o padrão de 2 minutos, isso dá 60 pontos por hora.
- **Até 50 PRs por lista.** Os números (total, "+N" dos times e o resumo) são
  exatos, mas o resumo dos seus PRs considera só os 50 atualizados mais
  recentemente.
- **Agenda:** até 100 eventos por dia na API do Google. Na tela, até 8.
