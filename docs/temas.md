# Temas

Um tema muda três coisas: as **cores**, a **borda** das caixas e a **fonte do
relógio**.

## Escolhendo

- **Na hora:** aperte `t` no painel. Ele passa pelos temas em ordem, e o
  rodapé mostra qual está ativo.
- **Para sempre:** coloque `THEME=<nome>` no `~/.config/painel/config`. A tecla
  `t` não grava a escolha, então use o config para manter o tema.
- **Para comparar:** `painel themes` mostra uma amostra de cada tema no
  terminal.

## Temas prontos

| Nome | Borda | Relógio | Observação |
|---|---|---|---|
| `default` | arredondada | blocos | Lilás, rosa e amarelo. Tem versão para fundo claro e para fundo escuro |
| `dracula` | arredondada | sombreado | Paleta [Dracula](https://draculatheme.com) |
| `nord` | fina | fino | Paleta [Nord](https://www.nordtheme.com), tons frios |
| `gruvbox` | grossa | blocos | Paleta [Gruvbox](https://github.com/morhetz/gruvbox), tons quentes |
| `tokyo-night` | arredondada | fino | Paleta Tokyo Night |
| `solarized` | dupla | blocos | Paleta [Solarized](https://ethanschoonover.com/solarized/), com cinzas para fundo claro e escuro |
| `mono` | ASCII (`+-|`) | blocos | Preto e branco, para terminais sem cor ou para gravar a tela |

Os temas com paleta fixa (`dracula`, `nord`, `gruvbox`, `tokyo-night`) foram
feitos para fundo escuro. Num terminal de fundo claro, prefira `default`,
`solarized` ou `mono`.

## Criando um tema

Crie um arquivo em `~/.config/painel/themes/`. O nome do arquivo é o nome do
tema. O formato é o mesmo do `config`:

```sh
# ~/.config/painel/themes/meu
BASE=nord                 # parte de um tema pronto (opcional; o padrão é default)
ACCENT=#FF9E64            # relógio, linha do tempo e seleção
PINK=#F7768E              # número das revisões e o "agora" da linha do tempo
GREEN=#9ECE6A             # meus PRs, CI ok, aprovado
YELLOW=#E0AF68            # agenda, reuniões, CI rodando, cards em andamento
BLUE=#7DCFFF              # Jira e status dos cards
SUBTLE=#565F89            # textos secundários e bordas
ERROR=#DB4B4B             # erros, CI falhando, mudanças pedidas
SELECTED_TEXT=#1A1B26     # texto da linha selecionada (o fundo é o ACCENT)
BORDER=rounded            # rounded, normal, thick, double, ascii ou hidden
CLOCK=slim                # blocks, slim ou shade
```

- Tudo é opcional. O que o arquivo não define vem do `BASE`.
- Uma cor pode ter duas versões, para fundo claro e escuro: `GREEN=#2B8A3E/#69DB7C`.
- Depois de criar, use `THEME=meu` ou aperte `t` até chegar nele. Os temas
  dos arquivos vêm depois dos prontos, em ordem alfabética.
- Não precisa recompilar nem reiniciar: o arquivo é lido toda vez que o tema é
  escolhido.

Fez um tema bonito? Mande um PR adicionando em `builtinThemes`, no `theme.go`.

## Fontes do relógio

```
blocks          slim            shade
 █  ███ ███    ╻ ╺━┓ ┏━┓     ▓  ▓▓▓ ▓▓▓
██    █ █ █    ┃   ┃ ┃ ┃    ▓▓    ▓ ▓ ▓
 █  ███ █ █    ┃ ┏━┛ ┃ ┃     ▓  ▓▓▓ ▓ ▓
 █  █   █ █    ┃ ┃   ┃ ┃     ▓  ▓   ▓ ▓
███ ███ ███    ╹ ┗━╸ ┗━┛    ▓▓▓ ▓▓▓ ▓▓▓
```

As três têm a mesma largura e a mesma altura, então trocar de fonte não muda o
layout.
