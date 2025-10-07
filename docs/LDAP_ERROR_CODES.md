# Códigos de Erro LDAP - Guia de Troubleshooting

## LDAP Result Code 32: No Such Object

### 🔍 O que significa?

O erro **"LDAP Result Code 32: No Such Object"** significa que o servidor LDAP **não conseguiu encontrar o objeto**
(usuário, grupo ou DN) que você está tentando acessar. É como procurar por um arquivo que não existe em um diretório.

### 📍 Onde Pode Ocorrer?

Este erro pode acontecer em **4 situações** no Txlog Server:

#### 1. **Base DN Incorreto** (Mais Comum)

```bash
# ❌ ERRADO
LDAP_BASE_DN=ou=users,dc=exemplo,dc=com

# ✅ CORRETO
LDAP_BASE_DN=dc=exemplo,dc=com
```text

**Problema:** O `LDAP_BASE_DN` está apontando para uma OU que não existe ou está incorreta.

**Como Verificar:**

```bash
# Teste se o Base DN existe
ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
  -b "dc=exemplo,dc=com" -s base "(objectClass=*)"

# Se retornar erro 32, o Base DN está errado
```text

**Solução:**

1. Descubra o Base DN correto explorando o servidor:

   ```bash
   ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
     -b "" -s base namingContexts
   ```

2. Atualize no `.env`:

   ```bash
   LDAP_BASE_DN=dc=exemplo,dc=com  # Use o valor correto
   ```

---

#### 2. **Bind DN Incorreto**

```bash
# ❌ ERRADO
LDAP_BIND_DN=cn=readonly,dc=exemplo,dc=com

# ✅ CORRETO
LDAP_BIND_DN=cn=readonly,ou=service-accounts,dc=exemplo,dc=com
```text

**Problema:** A conta de serviço (Bind DN) não existe no caminho especificado.

**Como Verificar:**

```bash
# Teste o Bind DN
ldapsearch -H ldap://servidor:389 -x \
  -D "cn=readonly,ou=service-accounts,dc=exemplo,dc=com" \
  -W -b "dc=exemplo,dc=com" "(objectClass=*)"

# Se retornar erro 32, o Bind DN não existe
```text

**Solução:**

1. Busque a conta de serviço:

   ```bash
   # Busque por CN
   ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
     -b "dc=exemplo,dc=com" "(cn=readonly)" dn
   ```

2. Use o DN completo retornado no `.env`

---

#### 3. **Admin Group ou Viewer Group Incorreto**

```bash
# ❌ ERRADO
LDAP_ADMIN_GROUP=cn=admins,ou=grupos,dc=exemplo,dc=com

# ✅ CORRETO
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=exemplo,dc=com
```text

**Problema:** O DN do grupo não existe.

**Como Verificar:**

```bash
# Teste se o grupo existe
ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
  -b "cn=admins,ou=groups,dc=exemplo,dc=com" -s base "(objectClass=*)"

# Se retornar erro 32, o grupo não existe nesse caminho
```text

**Solução:**

1. Busque o grupo correto:

   ```bash
   ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
     -b "dc=exemplo,dc=com" "(cn=admins)" dn
   ```

2. Use o DN completo do grupo no `.env`:

   ```bash
   LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=exemplo,dc=com
   ```

---

#### 4. **Usuário Não Encontrado no Base DN**

```bash
# Base DN muito restrito
LDAP_BASE_DN=ou=employees,dc=exemplo,dc=com

# Mas o usuário está em: uid=joao,ou=contractors,dc=exemplo,dc=com
```text

**Problema:** O usuário existe no LDAP, mas **fora** do Base DN configurado.

**Como Verificar:**

```bash
# Busque o usuário em todo o diretório
ldapsearch -H ldap://servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W \
  -b "dc=exemplo,dc=com" "(uid=joao)" dn

# Se encontrar o usuário em uma OU diferente, amplie o Base DN
```text

**Solução:**

- Use um Base DN mais amplo que inclua todos os usuários:

  ```bash
  # Em vez de:
  LDAP_BASE_DN=ou=employees,dc=exemplo,dc=com
  
  # Use:
  LDAP_BASE_DN=dc=exemplo,dc=com
  ```

---

## 🔧 Como Diagnosticar Erro 32 no Txlog Server

### Passo 1: Ativar Logs de DEBUG

No `.env`:

```bash
LOG_LEVEL=DEBUG
```text

Reinicie o servidor e tente fazer login. Você verá logs detalhados:

```text
time=... level=DEBUG msg="LDAP user search: baseDN=ou=users,dc=exemplo,dc=com, filter=(uid=joao)"
time=... level=ERROR msg="LDAP search failed: LDAP Result Code 32 \"No Such Object\""
```text

### Passo 2: Identificar Qual DN Está Incorreto

Os logs mostram qual operação falhou:

| Mensagem de Log | DN Incorreto | Variável .env |
|-----------------|--------------|---------------|
| "LDAP user search: baseDN=..." | Base DN | `LDAP_BASE_DN` |
| "Binding with service account: ..." | Bind DN | `LDAP_BIND_DN` |
| "LDAP search failed: ... filter=(uid=...)" | Base DN | `LDAP_BASE_DN` |
| "Failed to check admin group membership" | Admin Group | `LDAP_ADMIN_GROUP` |
| "Failed to check viewer group membership" | Viewer Group | `LDAP_VIEWER_GROUP` |

### Passo 3: Validar o DN Correto

Use `ldapsearch` ou o script `ldap-discovery.sh`:

```bash
./ldap-discovery.sh
# Opção 1: Explorar estrutura do diretório
# Opção 2: Buscar usuários
# Opção 3: Buscar grupos
```text

### Passo 4: Corrigir e Testar

1. Atualize o `.env` com o DN correto
2. Reinicie o servidor
3. Tente fazer login novamente

---

## 📋 Checklist de Verificação para Erro 32

Quando encontrar **"LDAP Result Code 32"**, verifique:

- [ ] **LDAP_BASE_DN** existe e está acessível?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "dc=exemplo,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_BIND_DN** existe (se configurado)?

  ```bash
  ldapsearch -H ldap://... -x -D "cn=readonly,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_ADMIN_GROUP** existe?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "cn=admins,ou=groups,dc=exemplo,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_VIEWER_GROUP** existe (se configurado)?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "cn=viewers,ou=groups,dc=exemplo,dc=com" -s base "(objectClass=*)"
  ```

- [ ] Usuários estão dentro do **LDAP_BASE_DN**?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "dc=exemplo,dc=com" "(uid=usuario)"
  ```

---

## 🌟 Exemplos de Configuração Correta

### OpenLDAP Típico

```bash
LDAP_BASE_DN=dc=empresa,dc=com
LDAP_BIND_DN=cn=readonly,ou=service-accounts,dc=empresa,dc=com
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=empresa,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,ou=groups,dc=empresa,dc=com
```text

### Active Directory

```bash
LDAP_BASE_DN=DC=empresa,DC=com
LDAP_BIND_DN=CN=LDAP Service,OU=Service Accounts,DC=empresa,DC=com
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Security Groups,DC=empresa,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Users,OU=Security Groups,DC=empresa,DC=com
```text

### FreeIPA

```bash
LDAP_BASE_DN=dc=empresa,dc=com
LDAP_BIND_DN=uid=readonly,cn=sysaccounts,cn=etc,dc=empresa,dc=com
LDAP_ADMIN_GROUP=cn=txlog-admins,cn=groups,cn=accounts,dc=empresa,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,cn=groups,cn=accounts,dc=empresa,dc=com
```text

---

## 🔍 Outros Códigos de Erro LDAP Comuns

### Code 34: Invalid DN Syntax

**O que significa:** O formato do DN está incorreto.

**Exemplo:**

```bash
# ❌ ERRADO (falta vírgula)
LDAP_BASE_DN=ou=usersdc=exemplo,dc=com

# ✅ CORRETO
LDAP_BASE_DN=ou=users,dc=exemplo,dc=com
```text

### Code 49: Invalid Credentials

**O que significa:** Usuário/senha incorretos.

**Causas comuns:**

1. Senha errada em `LDAP_BIND_PASSWORD`
2. Senha do usuário que está tentando fazer login está incorreta
3. Conta de serviço expirada ou desabilitada

**Como verificar:**

```bash
# Teste o Bind DN
ldapsearch -H ldap://servidor:389 -x \
  -D "cn=readonly,dc=exemplo,dc=com" \
  -w "sua_senha" \
  -b "dc=exemplo,dc=com" "(objectClass=*)"
```text

### Code 50: Insufficient Access Rights

**O que significa:** A conta não tem permissão para executar a operação.

**Solução:** A conta de serviço precisa de:

- Permissão de leitura no Base DN
- Permissão de leitura nos grupos configurados

### Code 52: Unavailable

**O que significa:** Servidor LDAP não está disponível.

**Causas:**

1. Servidor LDAP parado
2. Porta bloqueada por firewall
3. Problemas de rede

**Como verificar:**

```bash
# Teste conectividade
telnet ldap.servidor.com 389

# Ou com LDAPS
openssl s_client -connect ldap.servidor.com:636
```text

### Code 53: Unwilling to Perform

**O que significa:** Servidor recusou executar a operação.

**Causas comuns:**

1. Tentativa de modificar dados em modo read-only
2. Violação de política do servidor
3. Operação não permitida (ex: anonymous bind desabilitado)

---

## 🛠️ Ferramentas de Diagnóstico

### 1. Script ldap-discovery.sh

```bash
./ldap-discovery.sh
# Use as opções do menu para testar cada componente
```text

### 2. ldapsearch Manual

```bash
# Template completo de teste
ldapsearch -H ldap://SERVIDOR:PORTA \
  -x \
  -D "BIND_DN" \
  -W \
  -b "BASE_DN" \
  "FILTER" \
  atributos

# Exemplo real
ldapsearch -H ldap://ldap.empresa.com:389 \
  -x \
  -D "cn=readonly,dc=empresa,dc=com" \
  -W \
  -b "dc=empresa,dc=com" \
  "(uid=joao)" \
  dn uid cn mail
```text

### 3. Apache Directory Studio (GUI)

- Download: <https://directory.apache.org/studio/>
- Permite navegar visualmente pela árvore LDAP
- Mostra erros de forma mais amigável

### 4. Logs do Txlog Server

```bash
# Ative DEBUG no .env
LOG_LEVEL=DEBUG

# Execute o servidor
make run

# Logs mostrarão:
# - Base DN usado nas buscas
# - Filtros aplicados
# - Resultados de cada operação
# - Erros detalhados
```text

---

## 📊 Tabela Resumida de Códigos LDAP

| Código | Nome | Significado | Solução Comum |
|--------|------|-------------|---------------|
| 0 | Success | Operação bem-sucedida | N/A |
| 32 | No Such Object | DN não existe | Verificar DNs no .env |
| 34 | Invalid DN Syntax | Formato de DN incorreto | Verificar vírgulas e formato |
| 49 | Invalid Credentials | Usuário/senha incorretos | Verificar credenciais |
| 50 | Insufficient Access | Sem permissão | Ajustar ACLs da conta |
| 52 | Unavailable | Servidor indisponível | Verificar conectividade |
| 53 | Unwilling to Perform | Operação não permitida | Verificar políticas do servidor |
| 65 | Object Class Violation | Problema com objectClass | Verificar schema |

---

## 🚨 Troubleshooting Passo a Passo

### Quando receber "LDAP Result Code 32"

```bash
# 1. Ative logs detalhados
echo "LOG_LEVEL=DEBUG" >> .env

# 2. Reinicie o servidor e tente fazer login
make run

# 3. Nos logs, identifique qual DN falhou:
#    - "LDAP user search: baseDN=..." → problema no LDAP_BASE_DN
#    - "failed to bind with service account" → problema no LDAP_BIND_DN
#    - "Failed to check ... group membership" → problema no grupo

# 4. Teste o DN manualmente:
ldapsearch -H ldap://servidor:389 \
  -x -D "cn=admin,dc=exemplo,dc=com" -W \
  -b "DN_SUSPEITO" -s base "(objectClass=*)"

# 5. Se retornar erro 32, o DN está errado
#    Se retornar sucesso, o DN existe (problema em outro lugar)

# 6. Use o script para descobrir o DN correto:
./ldap-discovery.sh
# Opção 1: Explorar estrutura
# Opção 2 ou 3: Buscar o objeto correto

# 7. Atualize o .env com o DN correto

# 8. Reinicie e teste novamente
```text

---

## 📞 Precisa de Ajuda?

1. ✅ Use `./ldap-discovery.sh` para explorar seu LDAP
2. ✅ Ative `LOG_LEVEL=DEBUG` para ver detalhes
3. ✅ Teste cada DN manualmente com `ldapsearch`
4. ✅ Consulte `LDAP_FILTER_DISCOVERY.md` para guia completo

---

## ✨ Resumo

**"LDAP Result Code 32: No Such Object"** = **DN não existe**

**Verifique sempre:**

1. ✅ `LDAP_BASE_DN` - O ponto de partida das buscas
2. ✅ `LDAP_BIND_DN` - A conta de serviço (se usada)
3. ✅ `LDAP_ADMIN_GROUP` - O grupo de administradores
4. ✅ `LDAP_VIEWER_GROUP` - O grupo de visualizadores

**Use ferramentas:**

- `./ldap-discovery.sh` - Descoberta interativa
- `ldapsearch` - Testes manuais
- `LOG_LEVEL=DEBUG` - Logs detalhados

🎯 Na maioria dos casos, o erro 32 é causado por um **Base DN incorreto**!
