# Arquitetura

O painel é um único binário Go, sem servidor e sem banco. A interface usa o
[Bubble Tea](https://github.com/charmbracelet/bubbletea), que segue a
arquitetura Elm: um **modelo** com o estado, uma função **Update** que recebe
mensagens e devolve o novo estado, e uma função **View** que transforma o
estado em texto.

## Mapa dos arquivos

```
main.go         entrada: subcomandos (auth, version, help) e início do programa
config.go       leitura de ~/.config/painel/config e das variáveis PAINEL_*
model.go        estado, mensagens e Update: relógio, ciclos de busca, teclas
view.go         View: layout, caixas, cores e textos
bigclock.go     fonte de 5 linhas do relógio grande
prs.go          busca de PRs via `gh search prs`
agenda.go       agenda via iCal (URL ou arquivo) e ordenação dos eventos
google.go       login OAuth (`painel auth`) e agenda via Google Calendar API
view_test.go    testes de renderização com dados falsos
agenda_test.go  testes do iCal: recorrências, remarcações, recusados, repetidos
demo_test.go    gera os quadros do GIF do README (só com PAINEL_DEMO_DIR)
scripts/demo_gif.py       desenha os quadros em docs/demo.gif
.goreleaser.yaml          como os executáveis de release são gerados
.github/workflows/ci.yml  testes e compilação em cada push e PR
.github/workflows/release.yml  publica a release quando sobe uma tag v*
```

## Fluxo

```
            Init
             │
   ┌─────────┼──────────────┐
   ▼         ▼              ▼
 tick()   loadPRs()     loadAgenda()        ← comandos (rodam fora da UI)
   │         │              │
tickMsg    prsMsg       agendaMsg           ← mensagens de volta
   │         │              │
   ▼         ▼              ▼
           Update ──────────────► View
   │         │              │
 tick()   tea.Tick(PR_INTERVAL)  tea.Tick(CAL_INTERVAL)
             │              │
       refreshPRsMsg   refreshAgendaMsg ──► loadPRs()/loadAgenda() de novo
```

Pontos importantes:

- **Nada de rede dentro de `Update` ou `View`.** As buscas são `tea.Cmd`, que o
  Bubble Tea roda em goroutines. O resultado volta como mensagem (`prsMsg`,
  `agendaMsg`). Assim a tela nunca trava.
- **O próximo ciclo só é agendado quando o anterior termina.** Isso evita
  acumular requisições com a rede lenta.
- **Gerações (`prsGen`, `agendaGen`)** evitam buscas duplicadas: ao apertar `r`,
  o contador sobe, e os `refresh*Msg` agendados com o número antigo são
  ignorados.
- **Erro não apaga dado.** Se uma busca falha, o modelo guarda o erro mas
  mantém as listas anteriores.
- **O relógio** agenda cada tick para o início do próximo segundo, para virar
  junto com o relógio do sistema.

## Fontes de dados

Cada fonte é uma função que devolve um resultado completo, com dados, erro e
horário da busca:

```go
func fetchPRs() prsResult
func fetchAgenda(url, email string) agendaResult   // iCal
func fetchGoogleAgenda() agendaResult              // Google Calendar API
```

Todas são síncronas e não sabem nada da interface, o que facilita testar e
trocar. As duas fontes de agenda devolvem o mesmo `agendaResult` com
`[]event`. A escolha entre elas fica em `model.loadAgenda`, a partir de
`config.Agenda`.

## Como adicionar uma caixa nova

Exemplo: issues do Jira atribuídas a você.

1. **Fonte**: crie `jira.go` com um tipo de resultado e uma função de busca:

   ```go
   type jiraResult struct {
       Issues []issue
       Err    error
       At     time.Time
   }

   func fetchJira(cfg config) jiraResult { ... }
   ```

2. **Configuração**: adicione as chaves necessárias em `config` e em
   `loadConfig` (ex.: `JIRA_URL`, lido com `get("JIRA_URL")`).

3. **Modelo** (`model.go`):
   - campos `jira jiraResult`, `jiraLoading bool`, `jiraGen int`;
   - mensagens `jiraMsg` e `refreshJiraMsg struct{ gen int }`;
   - `loadJira()` retornando um `tea.Cmd`;
   - inclua `m.loadJira()` em `Init` e no atalho `r`;
   - trate `jiraMsg` (guardar o resultado e agendar o próximo ciclo) e
     `refreshJiraMsg` (conferir a geração), igual aos PRs.

4. **Tela** (`view.go`): crie `viewJiraBox(width int) string`, que devolve só o
   **conteúdo**. A borda e a altura comum são aplicadas em `View`. Adicione a
   caixa ao `JoinHorizontal`/`JoinVertical` e ajuste o cálculo de `cols` se
   passar de três colunas.

5. **Teste**: inclua dados falsos em `fakeModel` (`view_test.go`) e confira que
   o texto aparece na tela.

## Estilo

- Textos da interface, comentários e mensagens de erro em **português**.
- Cores são `lipgloss.AdaptiveColor` (valor claro e escuro) e ficam declaradas
  no topo de `view.go`.
- Mensagens de erro começam com a fonte (`gh:`, `agenda:`), porque aparecem
  sozinhas dentro da caixa.
- Credenciais e dados pessoais nunca entram no repositório: tudo fica em
  `~/.config/painel/`.

## Testes

```sh
make test
go test -run TestViewRenders -v ./...   # mostra a tela renderizada no log
```

- `view_test.go` monta um `model` com dados falsos e horário fixo e confere a
  tela em duas larguras (lado a lado e empilhada).
- `agenda_test.go` gera um `.ics` com horários relativos a hoje e confere quais
  eventos sobram. Cobre ocorrências remarcadas para outro dia, eventos à
  meia-noite, cancelados, recusados e repetidos.

Os testes não acessam a rede: `fetchPRs` e `fetchGoogleAgenda` não são chamadas
neles, e `fetchAgenda` lê um arquivo temporário.

> **Atenção com o gocal:** ele descarta eventos sem `DTSTAMP` (obrigatório pela
> RFC 5545) e só esconde uma ocorrência remarcada se a versão remarcada estiver
> na janela pedida. `fetchAgenda` contorna o segundo ponto com `SkipBounds` e
> faz o recorte do dia por conta própria. O comentário no código explica.

## GIF do README

O `docs/demo.gif` é gerado pelo próprio código, com dados fictícios. Assim ele
acompanha a tela de verdade e nunca mostra dados de ninguém:

```sh
make demo-gif
```

`TestDemoFrames` renderiza a `View` segundo a segundo (14:59:53 a 15:00:06,
para mostrar a virada da reunião) com cores truecolor e grava cada quadro em
ANSI. `scripts/demo_gif.py` desenha esses quadros com Pillow e a fonte DejaVu
Sans Mono. Os dados ficam em `demoModel`, em `demo_test.go`.

## Dependências

| Pacote | Uso |
|---|---|
| `github.com/charmbracelet/bubbletea` | Laço da interface, teclado, tela alternativa |
| `github.com/charmbracelet/lipgloss` | Cores, bordas e layout |
| `github.com/charmbracelet/x/ansi` | Cortar e quebrar texto que tem cores |
| `github.com/apognu/gocal` | Ler iCal e expandir recorrências |
| `golang.org/x/oauth2` | Login no Google com PKCE e renovação do token |

Programa externo: `gh` (GitHub CLI), chamado com `os/exec`.
