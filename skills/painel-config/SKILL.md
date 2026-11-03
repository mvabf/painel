---
name: painel-config
description: Assistente de configuração do painel (dashboard de terminal com PRs, agenda, Jira e lembretes). Use quando a pessoa pedir para instalar, configurar, atualizar ou consertar o painel, ligar a agenda ou o Jira, escolher tema ou notificações, ou quando o painel mostrar uma caixa pedindo configuração.
---

# Configurar o painel

Você vai conduzir a pessoa pela configuração do `painel`, um passo por vez.
Pergunte só o que precisa, explique em uma frase para que serve cada opção e
aceite "pular". Tudo é opcional: sem configuração nenhuma, os PRs já aparecem.

Referência completa: https://github.com/mvabf/painel/tree/main/docs
(`configuracao.md`, `agenda.md`, `notificacoes.md`, `temas.md`, `lembretes.md`,
`problemas.md`).

## Regras

- **Segredos:** nunca imprima, repita ou registre em log o valor de um token,
  do link secreto iCal ou do `google-client.json`. Para conferir, mostre só se
  a chave existe (`grep -c '^JIRA_TOKEN=' arquivo`), nunca o valor.
- **Não sobrescreva o config.** Leia `~/.config/painel/config` antes, mude só
  as chaves combinadas e preserve o resto, inclusive os comentários.
- **Permissões:** a pasta `~/.config/painel` fica com `700` e os arquivos dela
  com `600`.
- Mostre o que vai escrever (com os segredos mascarados) antes de escrever.

## 1. Conferir o ambiente

```sh
painel version
gh auth status
```

- **`painel: command not found`:** instale a última versão. No Linux e no
  macOS, escolha o arquivo pelo sistema (`uname -sm`): `painel_linux_amd64`,
  `painel_linux_arm64`, `painel_darwin_arm64` (Apple Silicon) ou
  `painel_darwin_amd64` (Intel).

  ```sh
  gh release download --repo mvabf/painel --pattern 'painel_linux_amd64.tar.gz' --dir /tmp/painel --clobber
  mkdir -p ~/.local/bin
  tar xzf /tmp/painel/painel_linux_amd64.tar.gz -C ~/.local/bin painel
  ```

  Se ainda não achar o comando, adicione `export PATH="$HOME/.local/bin:$PATH"`
  ao `~/.zshrc` ou `~/.bashrc`. No Windows nativo, o `.zip` da página de
  releases traz o `painel.exe`.
- **Atualizar** é rodar os mesmos comandos de instalação de novo.
- **`gh` sem login:** os PRs dependem dele. Peça para a pessoa rodar
  `gh auth login` (é interativo; ela mesma roda).
- **WSL** (`grep -qi microsoft /proc/sys/kernel/osrelease`): o `enter` abre os
  links no navegador do Windows. Se não abrir, sugira instalar o `wslu`
  (`sudo apt install wslu`) ou definir `export BROWSER=wslview`.

Depois leia o config atual, se existir:

```sh
cat ~/.config/painel/config 2>/dev/null
```

Se houver token direto no arquivo, não ecoe a linha: mascare antes de mostrar.

## 2. Nome

A saudação usa o primeiro nome do `git config --global user.name`. Pergunte só
se a pessoa quiser outro: `NAME=Ana`.

## 3. Agenda

Pergunte qual calendário a pessoa usa e ofereça, nesta ordem:

1. **Endereço secreto iCal** (o mais simples, atualiza sozinho). No Google
   Calendar pelo navegador: ⚙️ > Configurações > clique na sua agenda >
   Integrar agenda > **Endereço secreto no formato iCal**. Peça para ela colar
   o link e grave:

   ```sh
   ICS_URL=<link>
   EMAIL=<e-mail dela no calendário>
   ```

   O `EMAIL` esconde os eventos recusados. Para validar sem expor o link:
   `curl -s "$(grep '^ICS_URL=' ~/.config/painel/config | cut -d= -f2-)" | grep -c BEGIN:VEVENT`.
   Se a empresa bloqueou o recurso, a opção não aparece ou o link responde
   erro: passe para a opção 2 ou 3.
2. **Login no Google (`painel auth`)**: para Google Workspace sem o endereço
   secreto. Exige criar um client OAuth próprio, uma vez: siga
   `docs/agenda.md`, seção 2, com a pessoa, um passo por vez. O JSON baixado
   vai para `~/.config/painel/google-client.json` (`chmod 600`). Depois ela
   roda `painel auth`, que abre o navegador.
3. **Arquivo `.ics` exportado** (não atualiza sozinho): Google Calendar >
   Configurações > Importar e exportar > Exportar. Vem um `.zip`. Extraia numa
   pasta temporária vazia, escolha o `.ics` da agenda principal (o que tem o
   e-mail da pessoa no nome) e copie:

   ```sh
   install -m 600 <arquivo>.ics ~/.config/painel/calendar.ics
   ```

   e grave `ICS_URL=~/.config/painel/calendar.ics`. Avise que reuniões novas só
   entram exportando de novo.

## 4. Jira (opcional)

Só para Jira Cloud. Precisa de três valores: URL base
(`https://empresa.atlassian.net`, sem `/browse`), e-mail da conta Atlassian e
token de API, criado em
https://id.atlassian.com/manage-profile/security/api-tokens.

- **Prefira não copiar o token para o config.** Se a pessoa já tem um arquivo
  de segredos com `JIRA_BASE_URL`, `JIRA_EMAIL` e `JIRA_API_TOKEN`, grave só
  `ENV_FILE=<caminho>`. Se não tem, crie `~/.config/painel/jira.env` com
  `600` e peça para ela colar o token direto no arquivo (ou num comando que
  ela mesma roda), sem passar o valor pela conversa.
- Validar sem imprimir o token:

  ```sh
  set -a; . ~/.config/painel/jira.env; set +a
  curl -s -o /dev/null -w '%{http_code}\n' -u "$JIRA_EMAIL:$JIRA_API_TOKEN" "$JIRA_BASE_URL/rest/api/3/myself"
  ```

  `200` está certo; `401` é e-mail ou token errado; `404` é URL errada.
- Busca da caixa: o padrão são os cards atribuídos a ela e não concluídos. Para
  outra, `JIRA_JQL=...` (exemplos em `docs/configuracao.md`).

## 5. Ajustes

Ofereça em uma pergunta só, com os padrões:

- `REVIEW_MAX_AGE=7d`: esconde pedidos de revisão mais velhos que isso.
- `NOTIFY=on` (ou uma lista como `review,ci,meeting`): balões no desktop.
  `NOTIFY_SOUND=ci,meeting`: quais tocam som. `NOTIFY_BEFORE=5m`: antecedência
  das reuniões. No Linux precisa do `notify-send`; com o **Não perturbe**
  ligado o balão não aparece (os lembretes tocam mesmo assim).
- `THEME=...`: rode `painel themes` para ver as opções; a tecla `t` troca com
  o painel aberto e já salva a escolha no config.
- Intervalos: `PR_INTERVAL=2m`, `CAL_INTERVAL=5m`, `JIRA_INTERVAL=5m`.

## 6. Gravar e conferir

```sh
mkdir -p ~/.config/painel && chmod 700 ~/.config/painel
chmod 600 ~/.config/painel/config
```

Mostre o config final com os segredos mascarados e peça para a pessoa abrir
`painel` (ou `painel demo`, para ver como é sem dados reais). Apresente os
atalhos principais: `↑↓` navega, `enter` abre, `tab` troca de caixa, `n` cria
um lembrete, `r` atualiza, `q` sai. Se algo não aparecer, consulte
`docs/problemas.md`.
