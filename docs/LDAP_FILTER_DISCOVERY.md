# Descobrindo Filtros LDAP para seu Servidor

Este guia ajuda você a descobrir os valores corretos para `LDAP_USER_FILTER` e `LDAP_GROUP_FILTER` no seu ambiente LDAP específico.

## Índice

- [Ferramentas Necessárias](#ferramentas-necessárias)
- [Passo 1: Conectar ao Servidor LDAP](#passo-1-conectar-ao-servidor-ldap)
- [Passo 2: Explorar a Estrutura](#passo-2-explorar-a-estrutura)
- [Passo 3: Encontrar Usuários](#passo-3-encontrar-usuários)
- [Passo 4: Encontrar Grupos](#passo-4-encontrar-grupos)
- [Passo 5: Determinar Filtros](#passo-5-determinar-filtros)
- [Exemplos Comuns](#exemplos-comuns)

---

## Ferramentas Necessárias

### Linux/Mac

```bash
# Instalar ldap-utils (Debian/Ubuntu)
sudo apt-get install ldap-utils

# Instalar ldap-utils (Red Hat/CentOS/AlmaLinux)
sudo yum install openldap-clients

# Instalar ldap-utils (Mac)
brew install openldap
```text

### Windows

- Baixe e instale **Apache Directory Studio** (GUI): <https://directory.apache.org/studio/>
- Ou use **ldp.exe** (já incluído no Windows Server)

---

## Passo 1: Conectar ao Servidor LDAP

### Usando ldapsearch (Linha de Comando)

```bash
# Teste de conexão básica (sem TLS)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  "(objectClass=*)" dn

# Com TLS/LDAPS
ldapsearch -H ldaps://seu-servidor-ldap.com:636 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  "(objectClass=*)" dn
```text

**Parâmetros:**

- `-H`: URL do servidor LDAP
- `-x`: Autenticação simples
- `-b`: Base DN (ponto inicial da busca)
- `-D`: Bind DN (usuário para autenticação)
- `-W`: Solicita senha interativamente
- `-w password`: Senha na linha de comando (não recomendado)

### Usando Apache Directory Studio (GUI)

1. Abra o Apache Directory Studio
2. Clique em **"New Connection"**
3. Configure:
   - **Connection name**: Nome descritivo
   - **Hostname**: Endereço do servidor LDAP
   - **Port**: 389 (LDAP) ou 636 (LDAPS)
   - **Encryption**: None/LDAPS/StartTLS
4. Na aba **"Authentication"**:
   - **Authentication Method**: Simple Authentication
   - **Bind DN**: cn=admin,dc=exemplo,dc=com
   - **Bind Password**: sua senha
5. Clique em **"Check Network Parameter"** para testar
6. Clique em **"Finish"**

---

## Passo 2: Explorar a Estrutura

### Descobrir a Estrutura Base

```bash
# Listar todas as entradas no nível raiz
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -s one \
  "(objectClass=*)" dn

# Ver a estrutura completa (use com cuidado em diretórios grandes)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=organizationalUnit)" dn
```text

**Estruturas comuns:**

```text
dc=exemplo,dc=com
├── ou=users          ← Usuários geralmente ficam aqui
├── ou=people         ← Ou aqui
├── ou=grupos         ← Grupos geralmente ficam aqui
├── ou=groups         ← Ou aqui
└── ou=departments    ← Estrutura organizacional
```text

---

## Passo 3: Encontrar Usuários

### 3.1 Buscar Todos os Usuários

```bash
# Buscar por pessoas (pessoa)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=person)"

# Buscar por inetOrgPerson (mais comum)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=inetOrgPerson)"

# Buscar por posixAccount (sistemas Unix/Linux)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=posixAccount)"
```text

### 3.2 Examinar um Usuário Específico

```bash
# Buscar usuário por uid
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(uid=joao.silva)"

# Buscar usuário por cn (common name)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(cn=João Silva)"

# Buscar usuário por sAMAccountName (Active Directory)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(sAMAccountName=joao.silva)"
```text

### 3.3 Identificar o Atributo de Login

Examine a saída e procure por:

```ldif
dn: uid=joao.silva,ou=users,dc=exemplo,dc=com
objectClass: inetOrgPerson
objectClass: posixAccount
uid: joao.silva              ← Atributo de login (comum em OpenLDAP)
cn: João Silva
mail: joao.silva@exemplo.com
```text

ou

```ldif
dn: CN=João Silva,CN=Users,DC=exemplo,DC=com
objectClass: user
sAMAccountName: joao.silva   ← Atributo de login (Active Directory)
cn: João Silva
userPrincipalName: joao.silva@exemplo.com
mail: joao.silva@exemplo.com
```text

**Atributos comuns de login:**

- `uid`: OpenLDAP, FreeIPA, 389 Directory Server
- `sAMAccountName`: Active Directory
- `cn`: Alguns sistemas mais antigos
- `mail`: Alguns sistemas usam email como login

---

## Passo 4: Encontrar Grupos

### 4.1 Buscar Todos os Grupos

```bash
# Buscar por groupOfNames (LDAP padrão)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=groupOfNames)"

# Buscar por groupOfUniqueNames
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=groupOfUniqueNames)"

# Buscar por posixGroup (sistemas Unix/Linux)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=posixGroup)"

# Buscar por group (Active Directory)
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(objectClass=group)"
```text

### 4.2 Examinar um Grupo Específico

```bash
# Buscar grupo específico
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(cn=admins)"
```text

### 4.3 Identificar o Atributo de Membros

Examine a saída e procure por:

#### Tipo 1: groupOfNames (OpenLDAP)

```ldif
dn: cn=admins,ou=groups,dc=exemplo,dc=com
objectClass: groupOfNames
cn: admins
member: uid=joao.silva,ou=users,dc=exemplo,dc=com    ← DN completo do usuário
member: uid=maria.santos,ou=users,dc=exemplo,dc=com
```

#### Tipo 2: groupOfUniqueNames

```ldif
dn: cn=admins,ou=groups,dc=exemplo,dc=com
objectClass: groupOfUniqueNames
cn: admins
uniqueMember: uid=joao.silva,ou=users,dc=exemplo,dc=com  ← DN completo
uniqueMember: uid=maria.santos,ou=users,dc=exemplo,dc=com
```

#### Tipo 3: posixGroup

```ldif
dn: cn=admins,ou=groups,dc=exemplo,dc=com
objectClass: posixGroup
cn: admins
gidNumber: 1000
memberUid: joao.silva      ← Apenas o uid, não o DN completo
memberUid: maria.santos
```

#### Tipo 4: Active Directory

```ldif
dn: CN=Admins,CN=Users,DC=exemplo,DC=com
objectClass: group
cn: Admins
member: CN=João Silva,CN=Users,DC=exemplo,DC=com     ← DN completo
member: CN=Maria Santos,CN=Users,DC=exemplo,DC=com
```

---

## Passo 5: Determinar Filtros

### LDAP_USER_FILTER

Baseado no atributo de login identificado no **Passo 3.3**:

| Atributo de Login | LDAP_USER_FILTER | Sistema |
|------------------|------------------|---------|
| `uid` | `(uid=%s)` | OpenLDAP, FreeIPA, 389 DS |
| `sAMAccountName` | `(sAMAccountName=%s)` | Active Directory |
| `cn` | `(cn=%s)` | Sistemas antigos |
| `mail` | `(mail=%s)` | Login por email |
| `userPrincipalName` | `(userPrincipalName=%s)` | AD (login com email) |

**O `%s` será substituído pelo username digitado no login.**

### LDAP_GROUP_FILTER

Baseado no atributo de membros identificado no **Passo 4.3**:

| Atributo de Membro | LDAP_GROUP_FILTER | Sistema |
|-------------------|-------------------|---------|
| `member` | `(member=%s)` | groupOfNames, AD |
| `uniqueMember` | `(uniqueMember=%s)` | groupOfUniqueNames |
| `memberUid` | `(memberUid=%s)` | posixGroup |

**O `%s` será substituído pelo DN completo do usuário** (ex: `uid=joao.silva,ou=users,dc=exemplo,dc=com`)

**EXCEÇÃO:** Para `posixGroup` com `memberUid`, o Txlog Server precisa extrair apenas o `uid` do DN do usuário.

---

## Exemplos Comuns

### OpenLDAP com groupOfNames

```bash
# Estrutura
ou=users: uid=joao.silva
ou=groups: cn=admins com member=uid=joao.silva,ou=users,dc=exemplo,dc=com

# Configuração
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```text

### OpenLDAP com posixGroup

```bash
# Estrutura
ou=users: uid=joao.silva
ou=groups: cn=admins com memberUid=joao.silva

# Configuração
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)
```text

**⚠️ IMPORTANTE:** Para posixGroup, você precisa modificar o código do Txlog Server para extrair apenas o `uid` do DN antes de fazer a busca de grupo. Atualmente, ele passa o DN completo.

### Active Directory

```bash
# Estrutura
CN=Users: sAMAccountName=joao.silva
CN=Groups: cn=Admins com member=CN=João Silva,CN=Users,DC=exemplo,DC=com

# Configuração
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
```text

### FreeIPA

```bash
# Estrutura
cn=users: uid=joao.silva
cn=groups: cn=admins com member=uid=joao.silva,cn=users,cn=accounts,dc=exemplo,dc=com

# Configuração
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```text

### 389 Directory Server

```bash
# Estrutura similar ao OpenLDAP
ou=People: uid=joao.silva
ou=Groups: cn=admins com member=uid=joao.silva,ou=People,dc=exemplo,dc=com

# Configuração
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```text

---

## Testando os Filtros

### Testar LDAP_USER_FILTER

```bash
# Substitua %s pelo username real
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "ou=users,dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -LLL \
  "(uid=joao.silva)"

# Se retornar exatamente 1 usuário, o filtro está correto
```text

### Testar LDAP_GROUP_FILTER

```bash
# Primeiro, obtenha o DN completo do usuário
USER_DN="uid=joao.silva,ou=users,dc=exemplo,dc=com"

# Substitua %s pelo DN do usuário
ldapsearch -H ldap://seu-servidor-ldap.com:389 \
  -x \
  -b "cn=admins,ou=groups,dc=exemplo,dc=com" \
  -D "cn=admin,dc=exemplo,dc=com" \
  -W \
  -s base \
  -LLL \
  "(member=uid=joao.silva,ou=users,dc=exemplo,dc=com)"

# Se retornar o grupo, o filtro está correto
```text

---

## Filtros Avançados

### Combinar Múltiplos Atributos

```bash
# Buscar usuário por uid OU email
LDAP_USER_FILTER=(|(uid=%s)(mail=%s))

# Buscar usuário por sAMAccountName OU userPrincipalName (AD)
LDAP_USER_FILTER=(|(sAMAccountName=%s)(userPrincipalName=%s))
```text

### Filtrar por ObjectClass

```bash
# Garantir que é um inetOrgPerson com uid específico
LDAP_USER_FILTER=(&(objectClass=inetOrgPerson)(uid=%s))

# Garantir que é um grupo específico com membro
LDAP_GROUP_FILTER=(&(objectClass=groupOfNames)(member=%s))
```text

---

## Troubleshooting

### Erro: "user not found"

1. Verifique se o `LDAP_BASE_DN` está correto
2. Teste o `LDAP_USER_FILTER` manualmente com `ldapsearch`
3. Verifique se o usuário realmente existe no diretório

### Erro: "not a member of any authorized group"

1. Verifique se o `LDAP_ADMIN_GROUP` ou `LDAP_VIEWER_GROUP` está correto (deve ser o DN completo do grupo)
2. Teste o `LDAP_GROUP_FILTER` manualmente com `ldapsearch`
3. Verifique se o usuário é realmente membro do grupo no LDAP

### Erro: "failed to connect to LDAP"

1. Verifique se o host e porta estão corretos
2. Teste conectividade: `telnet ldap-server 389` ou `openssl s_client -connect ldap-server:636`
3. Verifique firewall e regras de rede

### Erro: "failed to bind with service account"

1. Verifique se o `LDAP_BIND_DN` está correto (formato completo do DN)
2. Verifique se a senha em `LDAP_BIND_PASSWORD` está correta
3. Teste o bind manualmente com `ldapsearch`

---

## Recursos Adicionais

- **OpenLDAP Documentation**: <https://www.openldap.org/doc/>
- **Active Directory LDAP**: <https://docs.microsoft.com/en-us/windows/win32/adsi/search-filter-syntax>
- **FreeIPA Documentation**: <https://www.freeipa.org/page/Documentation>
- **Apache Directory Studio**: <https://directory.apache.org/studio/>
- **LDAP Filter Syntax**: <https://ldap.com/ldap-filters/>

---

## Exemplo Completo de Configuração

```bash
# OpenLDAP com groupOfNames
LDAP_HOST=ldap.exemplo.com
LDAP_PORT=389
LDAP_USE_TLS=false
LDAP_BASE_DN=dc=exemplo,dc=com
LDAP_BIND_DN=cn=admin,dc=exemplo,dc=com
LDAP_BIND_PASSWORD=senha_secreta
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=exemplo,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=exemplo,dc=com
LDAP_GROUP_FILTER=(member=%s)
```text

```bash
# Active Directory
LDAP_HOST=ad.exemplo.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=DC=exemplo,DC=com
LDAP_BIND_DN=CN=Service Account,CN=Users,DC=exemplo,DC=com
LDAP_BIND_PASSWORD=senha_secreta
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,CN=Users,DC=exemplo,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Viewers,CN=Users,DC=exemplo,DC=com
LDAP_GROUP_FILTER=(member=%s)
```text

---

## Conclusão

Cada servidor LDAP pode ter uma estrutura diferente. Use as ferramentas de exploração (`ldapsearch` ou Apache Directory Studio) para:

1. ✅ Identificar onde os usuários estão armazenados
2. ✅ Identificar qual atributo é usado para login (uid, sAMAccountName, etc.)
3. ✅ Identificar onde os grupos estão armazenados
4. ✅ Identificar qual atributo armazena membros (member, uniqueMember, memberUid)
5. ✅ Testar os filtros manualmente antes de configurar o Txlog Server

Com essas informações, você pode configurar corretamente os filtros LDAP!
