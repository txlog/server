# Configuração LDAP Sem Service Account - Guia Prático

## Quando Usar

A autenticação **SEM service account** funciona quando:

✅ Seu servidor LDAP permite anonymous bind para buscas
✅ Usuários autenticados podem ler seus próprios grupos
✅ Você está usando OpenLDAP com configuração padrão
✅ Você quer uma configuração mais simples e com menos credenciais

## Quando NÃO Usar (Precisa de Service Account)

❌ Active Directory (geralmente requer autenticação para buscas)
❌ LDAP com ACLs restritas que bloqueiam anonymous bind
❌ Ambientes de produção com políticas de segurança rígidas
❌ LDAP que não permite usuários lerem seus próprios grupos

## Exemplo 1: OpenLDAP Básico (Sem Service Account)

```bash
# Configuração mínima - apenas 4 variáveis!
LDAP_HOST=ldap.minhaempresa.com
LDAP_BASE_DN=ou=users,dc=minhaempresa,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=minhaempresa,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=minhaempresa,dc=com
```

### Como Funciona

1. **Busca de Usuário**: Anonymous bind → busca usuário por `uid`
2. **Autenticação**: Bind com as credenciais do próprio usuário
3. **Verificação de Grupos**: Usa a sessão autenticada do usuário para ler grupos

## Exemplo 2: OpenLDAP com TLS (Sem Service Account)

```bash
LDAP_HOST=ldap.minhaempresa.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=ou=people,dc=minhaempresa,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=minhaempresa,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,ou=groups,dc=minhaempresa,dc=com
```

## Exemplo 3: Testando Sem Service Account

### Teste 1: Verificar se anonymous bind funciona

```bash
# Tenta buscar sem autenticação
ldapsearch -H ldap://ldap.minhaempresa.com:389 \
  -x \
  -b "ou=users,dc=minhaempresa,dc=com" \
  "(uid=meuusuario)"
```

**Se funcionar**: ✅ Pode usar sem service account
**Se falhar com "No such object" ou "Insufficient access"**: ❌ Precisa de service account

### Teste 2: Verificar leitura de grupos

```bash
# Autentica como usuário e tenta ler grupo
ldapsearch -H ldap://ldap.minhaempresa.com:389 \
  -D "uid=meuusuario,ou=users,dc=minhaempresa,dc=com" \
  -w "minhasenha" \
  -b "cn=admins,ou=groups,dc=minhaempresa,dc=com" \
  "(member=uid=meuusuario,ou=users,dc=minhaempresa,dc=com)"
```

**Se retornar o grupo**: ✅ Verificação de grupos funcionará
**Se falhar**: ❌ Precisa de service account com permissões de leitura

## Configuração OpenLDAP para Permitir Anonymous Bind

Se você administra o servidor OpenLDAP, configure para permitir anonymous reads:

```ldif
# /etc/ldap/slapd.conf ou via olcAccess

# Permitir anonymous read para usuários e grupos
olcAccess: {0}to dn.subtree="ou=users,dc=minhaempresa,dc=com"
  by anonymous read
  by * read

olcAccess: {1}to dn.subtree="ou=groups,dc=minhaempresa,dc=com"
  by anonymous read
  by * read
```

## Comparação: Com vs Sem Service Account

### SEM Service Account (Mais Simples)

**Prós:**

- ✅ Configuração mais simples (menos variáveis)
- ✅ Não precisa criar conta de serviço
- ✅ Menos credenciais para gerenciar
- ✅ Funciona bem com OpenLDAP padrão

**Contras:**

- ❌ Não funciona com Active Directory (geralmente)
- ❌ Requer anonymous bind habilitado
- ❌ Pode não atender políticas de segurança
- ❌ Usuário precisa ter permissão para ler grupos

**Configuração:**

```bash
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

### COM Service Account (Mais Robusto)

**Prós:**

- ✅ Funciona com Active Directory
- ✅ Funciona com LDAP restritivo
- ✅ Mais controle sobre permissões
- ✅ Melhor para produção

**Contras:**

- ❌ Mais variáveis de configuração
- ❌ Precisa criar e gerenciar conta de serviço
- ❌ Mais uma senha para guardar com segurança

**Configuração:**

```bash
LDAP_HOST=ldap.example.com
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=senha_da_conta_servico
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Fluxo de Autenticação Detalhado

### Sem Service Account

```text
1. Cliente envia username + password
   ↓
2. Servidor conecta ao LDAP (anonymous)
   ↓
3. Busca usuário: ldapsearch -x "(uid=username)"
   ↓
4. Encontra: uid=username,ou=users,dc=example,dc=com
   ↓
5. Autentica: bind com uid=username + password do usuário
   ↓
6. Verifica grupos usando a sessão autenticada do usuário
   ↓
7. Cria sessão no Txlog Server
   ↓
8. Usuário logado!
```

### Com Service Account

```text
1. Cliente envia username + password
   ↓
2. Servidor conecta ao LDAP
   ↓
3. Bind com service account
   ↓
4. Busca usuário: ldapsearch "(uid=username)"
   ↓
5. Encontra: uid=username,ou=users,dc=example,dc=com
   ↓
6. Autentica: bind com uid=username + password do usuário
   ↓
7. Re-bind com service account
   ↓
8. Verifica grupos usando service account
   ↓
9. Cria sessão no Txlog Server
   ↓
10. Usuário logado!
```

## Troubleshooting

### Erro: "User not found"

**Sem service account:**

```bash
# Teste anonymous search
ldapsearch -H ldap://seu-ldap:389 -x \
  -b "ou=users,dc=example,dc=com" \
  "(uid=testuser)"
```

**Solução:** Se falhar, você precisa de service account.

### Erro: "Failed to check group membership"

**Sem service account:**

```bash
# Teste se usuário pode ler grupos
ldapsearch -H ldap://seu-ldap:389 \
  -D "uid=testuser,ou=users,dc=example,dc=com" \
  -w "senha" \
  -b "cn=admins,ou=groups,dc=example,dc=com"
```

**Solução:** Se falhar, configure service account com permissão de leitura em grupos.

### Erro: "Failed to connect to LDAP"

Mesmo problema com ou sem service account - verifique:

- Host e porta corretos
- Firewall liberado
- LDAP server rodando

## Recomendações

### Desenvolvimento/Teste

✅ **Use SEM service account** se possível

- Mais rápido de configurar
- Menos complexo
- OpenLDAP local geralmente permite

### Produção

✅ **Use COM service account**

- Mais seguro
- Mais controle
- Funciona com Active Directory
- Atende políticas de segurança

### Ambientes Mistos

✅ **Comece SEM service account**

- Teste a conectividade básica
- Se funcionar, decida se vai adicionar service account
- Se não funcionar, adicione service account

## Exemplo Completo: Docker Compose

### Sem Service Account (OpenLDAP)

```yaml
version: '3.8'
services:
  txlog-server:
    image: cr.rda.run/txlog/server:main
    ports:
      - "8080:8080"
    environment:
      # Database
      - PGSQL_HOST=postgres
      - PGSQL_DB=txlog
      - PGSQL_USER=txlog
      - PGSQL_PASSWORD=txlog_password

      # LDAP - SEM service account
      - LDAP_HOST=openldap
      - LDAP_BASE_DN=ou=users,dc=example,dc=com
      - LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
      - LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

### Com Service Account (Active Directory)

```yaml
version: '3.8'
services:
  txlog-server:
    image: cr.rda.run/txlog/server:main
    ports:
      - "8080:8080"
    environment:
      # Database
      - PGSQL_HOST=postgres
      - PGSQL_DB=txlog
      - PGSQL_USER=txlog
      - PGSQL_PASSWORD=txlog_password

      # LDAP - COM service account
      - LDAP_HOST=ad.empresa.local
      - LDAP_BIND_DN=CN=SvcTxlog,OU=ServiceAccounts,DC=empresa,DC=local
      - LDAP_BIND_PASSWORD=senha_servico
      - LDAP_BASE_DN=CN=Users,DC=empresa,DC=local
      - LDAP_USER_FILTER=(sAMAccountName=%s)
      - LDAP_ADMIN_GROUP=CN=TxlogAdmins,OU=Groups,DC=empresa,DC=local
      - LDAP_VIEWER_GROUP=CN=TxlogUsers,OU=Groups,DC=empresa,DC=local
```

## Conclusão

A autenticação **SEM service account** é:

- ✅ **Mais simples** de configurar
- ✅ **Perfeitamente funcional** para OpenLDAP
- ✅ **Ideal para desenvolvimento** e ambientes menos restritivos
- ❌ **Não funciona** com Active Directory típico
- ❌ **Pode não atender** políticas de segurança corporativas

**Recomendação:** Comece sem service account. Se funcionar e atender suas
necessidades de segurança, ótimo! Se não funcionar ou se você precisa de mais
controle, adicione o service account.
