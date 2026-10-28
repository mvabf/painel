# Como funciona

## De onde vêm os PRs

O painel não guarda token do GitHub. Ele chama o **GitHub CLI** (`gh`), que já
está autenticado na sua máquina, e roda três buscas em paralelo:

| Busca | Equivalente no site | Vira |
|---|---|---|
| `user-review-requested:@me` | PRs que pediram revisão **para você** | Número principal de **Revisões pendentes** e a lista |
| `review-requested:@me` | O mesmo, **incluindo** pedidos aos seus times | Usado para calcular a linha "+N pedidos aos seus times" |
| `author:@me` | Seus PRs | **Meus PRs abertos** |

Todas filtram só PRs abertos, em qualquer repositório que você consiga ver.

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

### Ordem e rascunhos

As listas vêm ordenadas pela última atualização (a mais recente no topo).
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

## Quando cada coisa atualiza

| O quê | Quando |
|---|---|
| Relógio, "agora/próxima", destaques da agenda, barra do expediente | A cada segundo, sincronizado com o relógio do sistema |
| PRs | Ao abrir e depois a cada `PR_INTERVAL` (padrão 2 min) |
| Agenda | Ao abrir e depois a cada `CAL_INTERVAL` (padrão 5 min) |
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

- Com 110 colunas ou mais, as três caixas ficam lado a lado e com a mesma altura.
- Abaixo disso, elas ficam empilhadas e ocupam a largura toda.
- Títulos longos são cortados com `…` para não quebrar o layout.
- O painel usa a tela alternativa do terminal: ao sair, seu terminal volta
  exatamente como estava.

## Limites

- **Busca do GitHub:** a API permite 30 buscas por minuto. Cada ciclo do painel
  faz 3. Com o padrão de 2 minutos, isso dá 1,5 por minuto.
- **Até 100 PRs por busca.** Se você tiver mais de 100 pedidos de revisão
  somando os times, o "+N" fica subestimado.
- **Agenda:** até 100 eventos por dia na API do Google. Na tela, até 8.
