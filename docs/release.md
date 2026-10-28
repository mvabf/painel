# Publicando uma versão

As versões são publicadas pelo GitHub Actions. Ninguém compila nada na própria
máquina.

## Passo a passo

1. Confira que a `main` está verde no [CI](https://github.com/mvabf/painel/actions).
2. Crie e envie a tag da versão:

   ```sh
   git checkout main && git pull
   git tag v0.2.0
   git push origin v0.2.0
   ```

3. O workflow **Release** roda os testes, compila para todos os sistemas e
   cria a release com os arquivos anexados e um changelog gerado a partir dos
   PRs e commits. Leva uns 2 minutos.
4. Confira em [Releases](https://github.com/mvabf/painel/releases).

## Numeração

O projeto segue [versionamento semântico](https://semver.org/lang/pt-BR/):

| Mudança | Exemplo | Versão |
|---|---|---|
| Correção de bug | evento some da agenda | `v0.1.0` → `v0.1.1` |
| Funcionalidade nova, sem quebrar nada | caixa do Jira | `v0.1.1` → `v0.2.0` |
| Algo que exige ação de quem já usa | chave de config renomeada | `v0.x` → `v1.0.0` (ou próxima major) |

Enquanto a versão começar com `v0.`, mudanças incompatíveis podem acontecer em
qualquer versão minor. Descreva-as nas notas da release.

## O que é gerado

| Arquivo | Conteúdo |
|---|---|
| `painel_<sistema>_<arquitetura>.tar.gz` / `.zip` | Executável, README e `docs/` |
| `checksums.txt` | SHA-256 de cada arquivo |

Os nomes não levam o número da versão. Assim, o link
`.../releases/latest/download/painel_linux_amd64.tar.gz` sempre aponta para a
mais recente, e os comandos do README servem para instalar e para atualizar.

O executável é compilado sem CGO, então não depende de nenhuma biblioteca do
sistema. A versão é gravada nele (`painel version`) pelo `-ldflags` definido
em `.goreleaser.yaml`.

## Se der errado

- **A release não apareceu:** veja o log do workflow **Release** em Actions.
- **Tag errada:** apague a release na página do GitHub e a tag com
  `git push --delete origin v0.2.0 && git tag -d v0.2.0`. Depois crie de novo.
- **Testar a configuração sem publicar:** com o
  [GoReleaser](https://goreleaser.com/install/) instalado, rode
  `goreleaser release --snapshot --clean`. Os arquivos ficam em `dist/`.
