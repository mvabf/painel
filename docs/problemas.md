# Problemas comuns

## Instalação

**`painel: command not found`**
O README instala o binário em `~/.local/bin`, que pode não estar no seu
`PATH`. Adicione ao `~/.zshrc` ou `~/.bashrc`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Depois abra um terminal novo.

**`gh: release not found` ou `no assets match the file pattern`**
Confira o nome do arquivo na tabela do README. Se o repositório for privado, a
sua conta do `gh` precisa ter acesso a ele.

**macOS: "não é possível abrir porque o desenvolvedor não pode ser verificado"**
Acontece quando o arquivo foi baixado pelo navegador. Libere com
`xattr -d com.apple.quarantine ~/.local/bin/painel`. Baixando pelo `gh`, como
no README, isso não acontece.

**`go: command not found` ou erro de versão no `make`**
O projeto precisa do Go 1.24 ou mais novo. Confira com `go version` e instale
em [go.dev/doc/install](https://go.dev/doc/install).

**A tela aparece sem cores ou com caracteres estranhos**
Use um terminal com suporte a cores e UTF-8 (GNOME Terminal, iTerm2, Windows
Terminal, Alacritty, kitty, WezTerm…) e uma fonte com os caracteres de desenho
de caixa. Confira se `echo $LANG` termina em `UTF-8`.

## PRs

**`gh: ... not logged in` ou `gh: To get started with GitHub CLI, please run: gh auth login`**
Rode `gh auth login` e escolha GitHub.com.

**`gh: exec: "gh": executable file not found`**
O GitHub CLI não está instalado. Veja [cli.github.com](https://cli.github.com/).

**Os números ficam em `…` por muito tempo**
A primeira busca ainda não terminou. As três buscas levam poucos segundos, mas
uma rede lenta ou uma VPN podem demorar mais. Rode
`gh search prs --author=@me --state=open` direto no terminal para ver quanto demora.

**`gh: API rate limit exceeded`**
O seu login passou do limite da API do GitHub, que é compartilhado com outras
ferramentas (scripts, CI local, extensões). Aumente `PR_INTERVAL` ou espere o
limite renovar. `gh api rate_limit` mostra quanto falta.

**O número não bate com o que vejo no GitHub**
O número principal conta só os pedidos de revisão feitos **a você pelo nome**.
Os pedidos a times aparecem na linha "+N pedidos aos seus times". Veja
[funcionamento.md](funcionamento.md#de-onde-vêm-os-prs).

**PRs de uma organização não aparecem**
A organização pode exigir SSO. Rode `gh auth refresh` e autorize a organização
quando o navegador pedir.

## Agenda

**A caixa diz "Rode `painel auth` para usar o Google Calendar ou configure ICS_URL"**
Nenhuma fonte de agenda foi configurada. Veja [agenda.md](agenda.md).

**`agenda: HTTP 404`**
Você usou o endereço **público** (`/public/basic.ics`) em vez do **secreto**
(`/private-.../basic.ics`). Veja [agenda.md](agenda.md#1-endereço-secreto-ical).

**A agenda funcionava e passou a dar `agenda: HTTP 4xx` (iCal)**
O endereço secreto provavelmente foi redefinido. Copie o novo no Google Calendar.

**`agenda: open ...: no such file or directory`**
`ICS_URL` aponta para um arquivo que não existe. Confira o caminho.

**Não existe a opção "Endereço secreto no formato iCal"**
O administrador do Google Workspace desativou esse recurso. Use
[`painel auth`](agenda.md#2-login-no-google-calendar-painel-auth).

**Erros do `painel auth`**
Há uma tabela dedicada em
[agenda.md](agenda.md#quando-dá-errado).

**Uma reunião que recusei continua aparecendo**
No iCal, configure `EMAIL` com o e-mail exato do convite. No Google, isso é
automático.

**A agenda está desatualizada**
Se `ICS_URL` aponta para um arquivo exportado, ele é uma foto do momento da
exportação. Use o endereço secreto ou o `painel auth`. Se for um link, o Google
pode demorar (às vezes horas) para refletir mudanças no feed iCal. A API (`painel
auth`) reflete as mudanças na hora.

**Os horários estão deslocados**
O painel usa o fuso do sistema. Confira com `date`. Se estiver errado, ajuste o
fuso do sistema ou rode com `TZ=America/Sao_Paulo painel`.

## Navegador

**A tecla `p` ou o `painel auth` não abrem o navegador**
O painel usa `xdg-open` no Linux, `open` no macOS e o handler padrão no Windows.
No `painel auth`, o link também é impresso no terminal: copie e abra à mão.
