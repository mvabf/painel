# Segurança e privacidade

## Resumo

- O painel **só lê** dados. Ele não escreve nada no GitHub nem na agenda.
- Não há servidor do painel, telemetria nem coleta de dados. As únicas
  conexões são com o GitHub (via `gh`) e com o seu calendário.
- Nenhuma credencial fica dentro do repositório.

## O que fica guardado e onde

Tudo fica em `~/.config/painel/`. O painel cria essa pasta com permissão `700`
(só o seu usuário entra) e os tokens com `600` (só o seu usuário lê).

| Arquivo | Conteúdo | Sensível? |
|---|---|---|
| `config` | Suas preferências. Pode conter `ICS_URL` | **Sim, se `ICS_URL` for um endereço secreto**: ele dá acesso à sua agenda inteira |
| `google-client.json` | Identificação do client OAuth que você criou no Google Cloud | Moderado: sozinho não dá acesso a nada, mas não deve ser publicado |
| `google-token.json` | Token de acesso e de renovação do Google | **Sim**: permite ler os eventos da sua agenda |
| `*.ics` (se você usar um arquivo local) | Cópia da sua agenda | **Sim**: títulos, participantes e descrições das reuniões |

Se você criou algum desses arquivos à mão, garanta as permissões:

```sh
chmod 700 ~/.config/painel
chmod 600 ~/.config/painel/*
```

O **GitHub** não tem arquivo do painel: o acesso é o login que o `gh` já guarda
no chaveiro do sistema (ou em `~/.config/gh/hosts.yml`).

## Acessos

| Serviço | Acesso | Como revogar |
|---|---|---|
| GitHub | Os mesmos do seu `gh`. O painel só usa a API de busca | `gh auth logout`, ou em github.com > Settings > Applications |
| Google Calendar (`painel auth`) | `calendar.events.readonly`: só leitura de eventos | [myaccount.google.com/permissions](https://myaccount.google.com/permissions) e apague `google-token.json` |
| iCal secreto | Leitura da agenda para quem tiver o link | Google Calendar > Configurações > seu calendário > **Redefinir** endereço secreto |

## Conexões de rede

| Destino | Quando |
|---|---|
| `api.github.com` (pelo `gh`) | A cada `PR_INTERVAL` |
| O host de `ICS_URL` | A cada `CAL_INTERVAL`, se a fonte for iCal |
| `oauth2.googleapis.com` e `www.googleapis.com` | A cada `CAL_INTERVAL`, se a fonte for Google |
| `accounts.google.com` e `127.0.0.1:<porta aleatória>` | Só durante o `painel auth` |

O `painel auth` sobe um servidor temporário em `127.0.0.1` (acessível só na sua
máquina) para receber a resposta do login. Ele usa PKCE e confere o parâmetro
`state`, como recomenda o Google para apps de computador, e é encerrado assim
que o login termina.

## Ao compartilhar

- **Não faça commit** de nada que está em `~/.config/painel/`. Os arquivos ficam
  fora do repositório justamente para evitar isso.
- Cada pessoa deve criar o **próprio** client OAuth. Não compartilhe o
  `google-client.json`.
- Se um endereço secreto iCal vazar (em print, chat ou commit), redefina-o no
  Google Calendar. O antigo para de funcionar na hora.
