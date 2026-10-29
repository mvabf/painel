# Notificações

O painel pode avisar no desktop quando algo muda, para você não precisar ficar
olhando a tela. Vem **desligado**: cada pessoa escolhe o que quer receber.

## Ligando

No `~/.config/painel/config`:

```sh
NOTIFY=on                 # todos os avisos
NOTIFY_SOUND=ci,meeting   # só estes tocam som
NOTIFY_BEFORE=5m          # antecedência do aviso de reunião
```

`NOTIFY` e `NOTIFY_SOUND` aceitam `on` (tudo), `off` (nada) ou uma lista
separada por vírgula com os nomes abaixo.

## Eventos

| Nome | Quando avisa |
|---|---|
| `review` | Chegou um pedido de revisão novo para você |
| `approval` | Um PR seu foi **aprovado** ou recebeu **pedido de mudanças** |
| `ci` | O CI de um PR seu começou a falhar (avisa uma vez, não a cada busca) |
| `jira` | Um card novo apareceu na sua busca do Jira |
| `meeting` | Uma reunião começa em até `NOTIFY_BEFORE` (padrão 5 min) |

Algumas regras para não virar ruído:

- **Nada avisa ao abrir o painel.** A primeira busca de cada fonte só serve de
  referência; os avisos são sobre o que muda depois dela. A exceção é uma
  reunião que já está a menos de `NOTIFY_BEFORE` de começar.
- Cada reunião avisa uma vez só. Eventos de dia inteiro não avisam.
- Uma busca que falhou não gera aviso, nem a seguinte que der certo.
- Os pedidos aos seus times não avisam, só os feitos diretamente a você.

Uma sugestão: som só para o que pede ação rápida (`ci,meeting`) e balão
silencioso para o resto.

## Por sistema

| Sistema | Como o painel avisa | Som |
|---|---|---|
| Linux | `notify-send` (pacote `libnotify-bin` no Debian/Ubuntu, `libnotify` no Fedora/Arch) | Som `message-new-instant` do tema do sistema. O GNOME e o KDE tocam; alguns ambientes mais simples ignoram |
| macOS | `osascript` (já vem no sistema) | Som "Glass". Na primeira vez, o macOS pode pedir para permitir notificações do Script Editor |
| Windows | Ainda não avisa | — |

Se o modo **Não perturbe** do sistema estiver ligado, nem o balão nem o som
aparecem.

## Privacidade

O balão mostra o título do PR, do card ou da reunião. Se você compartilha a
tela com frequência, deixe só `meeting` ligado ou ative o Não perturbe durante
as apresentações (muitos apps de chamada fazem isso sozinhos).

## Limites

- O painel só avisa **enquanto está aberto**. Feche o terminal e os avisos param.
- Os avisos dependem das buscas: um pedido de revisão chega em até
  `PR_INTERVAL` (2 min por padrão), e um card do Jira em até `JIRA_INTERVAL`.
- Um PR que some e volta para a lista (por exemplo, quando pedem sua revisão
  de novo) avisa outra vez.
