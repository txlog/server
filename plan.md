# Plano de Remoção: Endpoints de Analytics

Este documento descreve as etapas necessárias para remover completamente todo o código, templates e documentação referentes às funcionalidades e endpoints `/analytics/compare`, `/analytics/freshness` e `/analytics/adoption`, bem como suas respectivas APIs em `/v1/reports/`.

---

## [x] 1. Atualizar o Roteamento Principal (`main.go`)

Devemos remover o registro das rotas das páginas HTML e também os endpoints da API associados no arquivo `main.go`.

**Trechos a serem removidos em `main.go`:**

Nas rotas das páginas de Analytics:
```go
	r.GET("/analytics/compare", controllers.GetAnalyticsCompare(database.Db))
	r.GET("/analytics/freshness", controllers.GetAnalyticsFreshness(database.Db))
	r.GET("/analytics/adoption", controllers.GetAnalyticsAdoption(database.Db))
```

Nos endpoints da API (`v1Group`):
```go
		v1Group.GET("/reports/compare-packages", v1API.ComparePackages(database.Db))
		v1Group.GET("/reports/package-freshness", v1API.GetPackageFreshness(database.Db))
		v1Group.GET("/reports/package-adoption", v1API.GetPackageAdoption(database.Db))
```

---

## [x] 2. Remover Lógica dos Controllers (`controllers/` e `controllers/api/v1/`)

### Em `controllers/analytics_controller.go`
Remover os handlers que servem as páginas HTML e a função de busca auxiliar de assets para a página de Compare.

**Remover:**
- `func GetAnalyticsCompare(database *sql.DB) gin.HandlerFunc`
- `func GetAnalyticsFreshness(database *sql.DB) gin.HandlerFunc`
- `func GetAnalyticsAdoption(database *sql.DB) gin.HandlerFunc`
- `type AssetForComparison struct { ... }`
- `func getActiveAssetsForComparison(database *sql.DB) ([]AssetForComparison, error)`

### Em `controllers/api/v1/analytics_controller.go`
Remover a lógica pesada de banco de dados que constrói os relatórios e formata via API.

**Remover:**
- `func ComparePackages(database *sql.DB) gin.HandlerFunc`
- `func comparePackageSets(database *sql.DB, machineIDs []string) (*models.PackageComparisonResult, error)`
- `func GetPackageFreshness(database *sql.DB) gin.HandlerFunc`
- `func getPackageFreshnessReport(database *sql.DB, machineID string, limit int) (*models.PackageFreshnessReport, error)`
- `func GetPackageAdoption(database *sql.DB) gin.HandlerFunc`
- `func getPackageAdoptionReport(database *sql.DB, limit, minAssets int) (*models.PackageAdoptionReport, error)`

### Em arquivos de Teste
Remover blocos correspondentes aos handlers acima nos arquivos `controllers/api/v1/analytics_controller_test.go`.

---

## [x] 3. Limpar Menus e Componentes de UI (`templates/`)

As referências aos relatórios que estão sendo removidos devem sumir da barra de navegação e das views conectadas.

### Em `templates/header.html`
Remover os links dentro do dropdown "Reports" (`#reports-menu`) e no menu lateral para versão celular (`#mobile-menu`):

**Trechos a remover:**
```html
<a class="block px-4 py-2 text-sm text-txlog-indigo hover:bg-txlog-lavender/30 transition-colors"
  href="/analytics/compare">Package Comparison</a>
<a class="block px-4 py-2 text-sm text-txlog-indigo hover:bg-txlog-lavender/30 transition-colors"
  href="/analytics/freshness">Package Freshness</a>
<a class="block px-4 py-2 text-sm text-txlog-indigo hover:bg-txlog-lavender/30 transition-colors"
  href="/analytics/adoption">Package Adoption</a>
```

### Em `templates/machine_id.html`
Remover o botão que direcionava para a comparação do asset atual.

**Trecho a remover:**
```html
<a href="/analytics/compare?preselect={{ .machine_id }}"
  class="inline-flex items-center gap-2 border-2 border-white/30 text-white text-sm font-medium px-4 py-2 rounded-xl hover:bg-white/10 transition-all">
  <i data-lucide="columns-2" class="w-4 h-4"></i> Compare with...
</a>
```

---

## [x] 4. Remoção de Arquivos (Templates)

Os arquivos fontes para estas views não têm mais utilidade e devem ser apagados por completo.

**Apagar os arquivos:**
- `templates/analytics_compare.html`
- `templates/analytics_freshness.html`
- `templates/analytics_adoption.html`

---

## [x] 5. Limpar Estruturas em Models

Encontrar e deletar em `models/` (provalvemente `models/reports.go` ou similar) as declarações referentes:
- `PackageComparisonResult`, `AssetPackageSet`, `PackageVersionDiff`
- `PackageFreshnessReport`, `PackageFreshnessInfo`
- `PackageAdoptionReport`, `PackageAdoptionInfo`

---

## [x] 6. Documentação (`docs/`)

Remover a menção destas rotas e das APIs nos tutoriais e referências do sistema e das anotações do Swagger para manter a fonte de verdade limpa.

### Em `docs/how-to/use-reports.md`
- Apagar da tabela Overview as linhas `Package Comparison`, `Package Freshness` e `Package Adoption`.
- Apagar os blocos com exemplos de comando `curl` (H3 `### Compare Packages`, `### Package Freshness`, `### Package Adoption`).

### Em `docs/reference/api-reference.md`
- Apagar do catálogo as entradas e payloads detalhando `.GET("/v1/reports/compare-packages", ...)`, `package-freshness` e `package-adoption`.

### Regerar Swagger
Por fim, rodar a suíte/comando para atualizar o Swagger, o que removerá do `docs/swagger.json`, `docs/swagger.yaml` e `docs/docs.go` a definição dos endpoints, agora que seus comentários formatados e funções foram retirados de `/v1/analytics_controller.go`:
```bash
make doc
```
