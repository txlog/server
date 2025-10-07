# Guia Rápido: Filtros LDAP

## TL;DR - Descobrindo seus Filtros

### Método 1: Script Automático (Recomendado)

```bash
./ldap-discovery.sh
```

### Método 2: Manual com ldapsearch

**1. Encontrar usuários:**

```bash
ldapsearch -H ldap://seu-servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(uid=username)"
```

**2. Ver atributos de um usuário:**

```bash
ldapsearch -H ldap://seu-servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(uid=joao)" dn uid cn sAMAccountName
```

**3. Ver grupos:**

```bash
ldapsearch -H ldap://seu-servidor:389 -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(cn=admins)" dn member uniqueMember memberUid
```

---

## Valores Comuns por Tipo de Servidor

### OpenLDAP (Padrão)

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### Active Directory

```bash
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### FreeIPA

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### OpenLDAP com posixGroup

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)
```

⚠️ **Atenção:** posixGroup usa apenas o `uid` (ex: `joao`) e não o DN completo.

---

## Tabela de Referência Rápida

### USER_FILTER - Por Sistema

| Sistema | Atributo | Filtro |
|---------|----------|--------|
| OpenLDAP | `uid` | `(uid=%s)` |
| Active Directory | `sAMAccountName` | `(sAMAccountName=%s)` |
| AD (email login) | `userPrincipalName` | `(userPrincipalName=%s)` |
| FreeIPA | `uid` | `(uid=%s)` |
| Antigos | `cn` | `(cn=%s)` |
| Email login | `mail` | `(mail=%s)` |

### GROUP_FILTER - Por Tipo de Grupo

| ObjectClass | Atributo de Membro | Filtro | Valor Esperado |
|-------------|-------------------|--------|----------------|
| `groupOfNames` | `member` | `(member=%s)` | DN completo |
| `groupOfUniqueNames` | `uniqueMember` | `(uniqueMember=%s)` | DN completo |
| `posixGroup` | `memberUid` | `(memberUid=%s)` | Apenas uid |
| `group` (AD) | `member` | `(member=%s)` | DN completo |

---

## Como Saber Qual Usar?

### Passo 1: Identificar o Atributo de Login do Usuário

Busque um usuário e veja qual campo contém o nome de login:

```bash
ldapsearch -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(objectClass=person)" uid cn sAMAccountName
```

Exemplo de saída OpenLDAP:

```text
dn: uid=joao.silva,ou=users,dc=exemplo,dc=com
uid: joao.silva          ← Este é o campo de login!
cn: João Silva
```

Exemplo de saída Active Directory:

```text
dn: CN=João Silva,CN=Users,DC=exemplo,DC=com
sAMAccountName: joao.silva    ← Este é o campo de login!
cn: João Silva
```

**Resultado:** Use o nome do atributo no filtro → `(uid=%s)` ou `(sAMAccountName=%s)`

---

### Passo 2: Identificar o Atributo de Membro do Grupo

Busque um grupo e veja qual campo lista os membros:

```bash
ldapsearch -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(cn=admins)" member uniqueMember memberUid
```

Exemplo com `member`:

```text
dn: cn=admins,ou=groups,dc=exemplo,dc=com
member: uid=joao.silva,ou=users,dc=exemplo,dc=com    ← DN completo
member: uid=maria.santos,ou=users,dc=exemplo,dc=com
```

**Resultado:** `LDAP_GROUP_FILTER=(member=%s)`

Exemplo com `memberUid`:

```text
dn: cn=admins,ou=groups,dc=exemplo,dc=com
memberUid: joao.silva    ← Apenas o uid, sem DN
memberUid: maria.santos
```

**Resultado:** `LDAP_GROUP_FILTER=(memberUid=%s)`
⚠️ **Requer modificação no código para extrair apenas o uid do DN**

---

## Testando Antes de Configurar

### Teste 1: Usuário pode ser encontrado?

```bash
# Substitua %s pelo username real
ldapsearch -x -D "cn=admin,dc=exemplo,dc=com" -W -b "dc=exemplo,dc=com" "(uid=joao.silva)"
```

✅ Deve retornar **exatamente 1 usuário**

### Teste 2: Usuário pertence ao grupo?

```bash
# Substitua o DN do grupo e do usuário
ldapsearch -x -D "cn=admin,dc=exemplo,dc=com" -W \
  -b "cn=admins,ou=groups,dc=exemplo,dc=com" \
  -s base \
  "(member=uid=joao.silva,ou=users,dc=exemplo,dc=com)"
```

✅ Deve retornar o grupo se o usuário for membro

---

## Exemplo Completo de .env

```bash
# OpenLDAP
LDAP_HOST=ldap.empresa.com
LDAP_PORT=389
LDAP_USE_TLS=false
LDAP_BASE_DN=dc=empresa,dc=com
LDAP_BIND_DN=cn=readonly,dc=empresa,dc=com
LDAP_BIND_PASSWORD=senha_readonly

LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=empresa,dc=com
LDAP_VIEWER_GROUP=cn=txlog-viewers,ou=groups,dc=empresa,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

```bash
# Active Directory
LDAP_HOST=ad.empresa.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_SKIP_TLS_VERIFY=false
LDAP_BASE_DN=DC=empresa,DC=com
LDAP_BIND_DN=CN=LDAP Service,OU=Service Accounts,DC=empresa,DC=com
LDAP_BIND_PASSWORD=senha_da_conta_servico

LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Security Groups,DC=empresa,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Users,OU=Security Groups,DC=empresa,DC=com
LDAP_GROUP_FILTER=(member=%s)
```

---

## Erros Comuns

| Erro | Causa | Solução |
|------|-------|---------|
| "user not found" | LDAP_USER_FILTER errado | Use `ldapsearch` para testar o filtro |
| "not a member of any authorized group" | LDAP_GROUP_FILTER errado ou grupo incorreto | Verifique se o usuário está no grupo e teste o filtro |
| "failed to bind" | LDAP_BIND_DN ou senha incorretos | Teste o bind manualmente |
| "connection refused" | Host/porta incorretos ou firewall | Verifique conectividade com `telnet` |

---

## Recursos

- **Documento completo:** `LDAP_FILTER_DISCOVERY.md`
- **Script interativo:** `./ldap-discovery.sh`
- **Documentação LDAP oficial:** <https://ldap.com/ldap-filters/>

---

## Dica Final

**Use o script `ldap-discovery.sh`** - ele guia você passo a passo para descobrir todos os valores necessários de forma interativa!

```bash
chmod +x ldap-discovery.sh
./ldap-discovery.sh
```
