# Conectando a agenda

O painel mostra os eventos de **hoje** (da meia-noite à meia-noite, no fuso do
seu computador). Há três fontes possíveis:

| Fonte | Atualiza sozinha | Esforço | Quando usar |
|---|---|---|---|
| [Endereço secreto iCal](#1-endereço-secreto-ical) | sim | 1 minuto | Primeira opção, se a sua conta tiver o recurso |
| [Login no Google Calendar](#2-login-no-google-calendar-painel-auth) | sim | ~5 minutos, uma vez | Contas Google Workspace sem o endereço secreto |
| [Arquivo `.ics` local](#3-arquivo-ics-local) | não | 1 minuto | Testar, ou calendários de outros serviços |

O painel escolhe a fonte assim:

1. Se a chave `AGENDA` estiver definida (`google` ou `ics`), usa essa.
2. Senão, se você já fez `painel auth`, usa o Google.
3. Senão, se `ICS_URL` estiver definida, usa o iCal.
4. Senão, a caixa da agenda mostra como configurar.

---

## 1. Endereço secreto iCal

Todo calendário do Google tem um link privado que entrega a agenda no formato
iCal. O painel baixa esse link periodicamente.

1. Abra [calendar.google.com](https://calendar.google.com) **no navegador do
   computador** (o app do celular não mostra essa opção).
2. Clique na engrenagem ⚙️ no canto superior direito e depois em **Configurações**.
3. Na coluna da esquerda, em **Configurações das minhas agendas**, clique no seu
   nome.
4. Desça até **Integrar agenda**.
5. No fim dessa seção, em **Endereço secreto no formato iCal**, clique no ícone
   de copiar.
6. Cole no arquivo `~/.config/painel/config`:

   ```sh
   ICS_URL=https://calendar.google.com/calendar/ical/voce%40empresa.com/private-1a2b3c.../basic.ics
   EMAIL=voce@empresa.com
   ```

   O `EMAIL` é opcional. Com ele, o painel esconde as reuniões que você recusou.

> **Cuidado com o endereço "público".** A mesma seção mostra um link com
> `/public/basic.ics`. Ele só funciona se o calendário estiver aberto para
> qualquer pessoa. Do contrário, o painel mostra `agenda: HTTP 404`. Use o que
> tem `/private-.../`.

> **Trate o endereço secreto como senha.** Quem tiver o link vê toda a sua
> agenda. Se ele vazar, use **Redefinir** na mesma tela para gerar um novo.

**Não achou o campo "Endereço secreto"?** Em muitas contas Google Workspace, o
administrador desativa esse recurso. Nesse caso, use a opção 2.

O mesmo funciona com outros serviços que oferecem um link iCal (Outlook.com,
iCloud, Fastmail…): cole o link `https://...` em `ICS_URL`.

---

## 2. Login no Google Calendar (`painel auth`)

Aqui o painel acessa a API do Google Calendar com a sua permissão, como
qualquer app que pede "Fazer login com o Google". O Google exige que todo app
desse tipo tenha uma identificação própria (um "client OAuth"). Como o painel
não é um app publicado, **cada pessoa cria a sua**, uma vez só.

### Passo a passo

1. **Crie um projeto no Google Cloud.**
   Acesse [console.cloud.google.com](https://console.cloud.google.com), clique
   no seletor de projeto no topo, depois em **Novo projeto**, e dê um nome como
   `painel`. Não precisa de faturamento.

2. **Ative a API do Calendar.**
   Menu ☰ > **APIs e serviços** > **Biblioteca**. Procure **Google Calendar
   API** e clique em **Ativar**.

3. **Configure a tela de permissão.**
   **APIs e serviços** > **Tela de permissão OAuth** (às vezes aparece como
   **Google Auth Platform**).
   - Se o projeto foi criado com a conta da empresa, escolha o tipo **Interno**.
     Basta preencher o nome do app (`painel`) e o seu e-mail.
   - Se foi criado com uma conta pessoal (Gmail), escolha **Externo**. Em
     **Usuários de teste** (ou **Público-alvo**), adicione o e-mail da conta cuja
     agenda você quer ver.

4. **Crie a credencial.**
   **APIs e serviços** > **Credenciais** > **Criar credenciais** > **ID do
   cliente OAuth**. Em tipo de aplicativo, escolha **App para computador**. Dê
   qualquer nome e clique em **Criar**.

5. **Baixe o JSON** na janela que aparece (botão **Fazer o download do JSON**)
   e salve como:

   ```
   ~/.config/painel/google-client.json
   ```

   Por exemplo: `mv ~/Downloads/client_secret_*.json ~/.config/painel/google-client.json`

6. **Faça o login:**

   ```sh
   painel auth
   ```

   O navegador abre na tela de login do Google. Escolha a conta e autorize.
   Quando aparecer "Pronto, pode voltar ao terminal.", o painel já está usando
   o Google Calendar.

### O que o painel pode fazer com esse acesso

O login pede só o escopo `calendar.events.readonly`: **ler eventos**. O painel
não consegue criar, mudar ou apagar nada na sua agenda. Você pode revogar o
acesso a qualquer momento em
[myaccount.google.com/permissions](https://myaccount.google.com/permissions).

Uma vantagem sobre o iCal: o Google informa a sua resposta a cada convite, então
as reuniões recusadas somem sem precisar configurar `EMAIL`.

### Quando dá errado

| Mensagem | O que fazer |
|---|---|
| `não achei .../google-client.json` | O passo 5 não foi feito, ou o arquivo está com outro nome |
| `google-client.json inválido` | A credencial não é do tipo **App para computador**. Crie outra (passo 4) |
| No navegador: **"Acesso bloqueado"** ou **"Este app está bloqueado"** | O administrador do Google Workspace bloqueia apps não aprovados. Só ele libera, adicionando o client ID em *Controles de API* do console de administração |
| No navegador: **"O app não concluiu o processo de verificação"** | Projeto **Externo**: falta adicionar o seu e-mail como usuário de teste (passo 3) |
| No navegador: **"Este app não foi verificado"** com link *Avançado* | Normal em projeto pessoal. Clique em *Avançado* > *Acessar painel* |
| `tempo esgotado esperando o login` | O login não foi concluído em 5 minutos. Rode `painel auth` de novo |
| `agenda: Google Calendar respondeu HTTP 403` | A Google Calendar API não foi ativada no projeto (passo 2) |
| `agenda: ... invalid_grant` | O acesso foi revogado, ou o projeto Externo está em modo de teste (nesse modo o token expira em 7 dias). Rode `painel auth` de novo |

> **Projeto Externo em modo de teste:** o Google invalida o login a cada 7 dias.
> Para evitar isso, publique o app em **Tela de permissão OAuth** > **Publicar
> app**. Para uso pessoal, não é preciso passar pela verificação.

### Voltar para o iCal

Defina `AGENDA=ics` no config ou apague `~/.config/painel/google-token.json`.

---

## 3. Arquivo `.ics` local

`ICS_URL` também aceita o caminho de um arquivo no seu computador:

```sh
ICS_URL=~/.config/painel/calendario.ics
```

Serve para qualquer calendário exportado: no Google, **Configurações** >
**Importar e exportar** > **Exportar** (vem um `.zip` com um `.ics` dentro). No
Outlook e no Apple Calendar, use a opção de exportar.

O painel relê o arquivo a cada ciclo, mas **o arquivo não muda sozinho**: é uma
foto da agenda no momento da exportação. Reuniões marcadas depois não aparecem.

---

## Detalhes que valem para todas as fontes

- **Recorrências:** reuniões recorrentes aparecem certas no dia, inclusive
  exceções. Uma ocorrência remarcada aparece só no novo horário, mesmo que
  tenha ido para outro dia, e uma ocorrência cancelada some.
- **Cancelados e recusados:** eventos cancelados nunca aparecem. Recusados
  somem no Google sempre, e no iCal quando `EMAIL` está configurado.
- **Repetidos:** eventos idênticos (mesmo título e mesmo horário) aparecem uma
  vez só. O feed do Google às vezes traz duas séries iguais que a interface
  dele mostra como uma.
- **Local de trabalho:** o "Trabalhando de casa/escritório" do Google não entra
  na lista quando a fonte é o Google, porque a API marca esse tipo de evento. No
  iCal não há essa marcação, então ele aparece como evento de dia inteiro.
- **Dia inteiro:** aparecem primeiro, com "dia todo" no lugar da hora.
- **Limites:** a caixa mostra até 8 eventos. A API do Google é consultada com
  até 100 eventos por dia.
