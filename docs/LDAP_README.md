# LDAP Documentation - Master Index

This directory contains complete documentation for LDAP authentication in Txlog Server.

## 📚 Available Documentation

### 🚀 Getting Started

1. **[LDAP_AUTHENTICATION.md](LDAP_AUTHENTICATION.md)**  
   📖 Complete LDAP authentication guide  
   - Feature overview
   - Step-by-step configuration
   - Environment variables
   - Practical examples
   - Troubleshooting

### 🔍 Discovering Your LDAP Filters

**Each LDAP server is different!** Use these resources to discover the correct values:

1. **[LDAP_FILTERS_QUICK.md](LDAP_FILTERS_QUICK.md)** ⭐ **START HERE**  
   ⚡ Quick and practical guide  
   - Reference table by server type
   - Ready-to-use commands
   - Common values (OpenLDAP, AD, FreeIPA)
   - How to test your filters

2. **[LDAP_FILTER_DISCOVERY.md](LDAP_FILTER_DISCOVERY.md)**  
   📘 Complete and detailed guide  
   - Step-by-step LDAP exploration
   - Explanation of each filter type
   - Using ldapsearch and Apache Directory Studio
   - Advanced troubleshooting

3. **[ldap-discovery.sh](ldap-discovery.sh)** 🛠️ **Interactive Script**  

   ```bash
   chmod +x ldap-discovery.sh
   ./ldap-discovery.sh
   ```

   - Interactive menu to explore your LDAP
   - Automatically tests connection
   - Discovers users and groups
   - Tests filters in real-time
   - Generates recommended configuration

### 📋 Quick References

4. **[LDAP_QUICK_REFERENCE.md](LDAP_QUICK_REFERENCE.md)**  
   📄 Configuration cheatsheet  
   - Summary of environment variables
   - .env examples by scenario
   - Useful commands

### 🔐 Service Account

5. **[LDAP_SERVICE_ACCOUNT_FAQ.md](LDAP_SERVICE_ACCOUNT_FAQ.md)**  
   ❓ Service account FAQ  
   - When to use service account vs anonymous bind
   - How to create service account
   - Required permissions
   - Security best practices

6. **[LDAP_SEM_SERVICE_ACCOUNT.md](LDAP_SEM_SERVICE_ACCOUNT.md)**  
   🔓 How to use without service account (anonymous bind)  
   - Anonymous bind configuration
   - Limitations and considerations
   - When it's possible to use

### 🏗️ Technical Information

7. **[LDAP_IMPLEMENTATION_SUMMARY.md](LDAP_IMPLEMENTATION_SUMMARY.md)**  
   🔧 Implementation details  
   - Code architecture
   - Authentication flow
   - Database structure
   - API endpoints

### 🚨 Troubleshooting

8. **[LDAP_ERROR_CODES.md](LDAP_ERROR_CODES.md)**  
   🔍 LDAP error codes explained  
   - **LDAP Result Code 32: No Such Object** (most common)
   - LDAP Result Code 49: Invalid Credentials
   - LDAP Result Code 50: Insufficient Access
   - How to diagnose each error
   - Practical solutions

---

## 🎯 Fluxo de Configuração Recomendado

```text
┌─────────────────────────────────────────────────────────────┐
│ 1. Leia LDAP_AUTHENTICATION.md                              │
│    └─ Entenda como funciona e o que é necessário            │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Use LDAP_FILTERS_QUICK.md ou ldap-discovery.sh          │
│    └─ Descubra os valores corretos para seu LDAP            │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Configure o .env com os valores descobertos              │
│    └─ Siga os exemplos do LDAP_QUICK_REFERENCE.md          │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Inicie o servidor e teste o login                        │
│    └─ Use o troubleshooting se necessário                   │
└─────────────────────────────────────────────────────────────┘
```text

---

## 🔍 Qual Documento Usar?

### Preciso descobrir LDAP_USER_FILTER e LDAP_GROUP_FILTER?

➡️ **Use:** `LDAP_FILTERS_QUICK.md` ou `./ldap-discovery.sh`

### Primeira vez configurando LDAP?

➡️ **Use:** `LDAP_AUTHENTICATION.md` → depois `LDAP_FILTERS_QUICK.md`

### Meu LDAP não tem conta de serviço?

➡️ **Use:** `LDAP_SEM_SERVICE_ACCOUNT.md`

### Preciso de um exemplo rápido de .env?

➡️ **Use:** `LDAP_QUICK_REFERENCE.md`

### Tenho dúvidas sobre segurança/permissões?

➡️ **Use:** `LDAP_SERVICE_ACCOUNT_FAQ.md`

### Estou tendo problemas de autenticação?

➡️ **Use:** Seção Troubleshooting do `LDAP_AUTHENTICATION.md` ou `LDAP_ERROR_CODES.md`

### Recebendo "LDAP Result Code 32: No Such Object"?

➡️ **Use:** `LDAP_ERROR_CODES.md` - Explica este erro em detalhes

### Quero entender como funciona por dentro?

➡️ **Use:** `LDAP_IMPLEMENTATION_SUMMARY.md`

### Preciso de todos os detalhes sobre filtros?

➡️ **Use:** `LDAP_FILTER_DISCOVERY.md`

---

## 🌟 Exemplos de Configuração por Servidor

### OpenLDAP

```bash
LDAP_HOST=ldap.empresa.com
LDAP_PORT=389
LDAP_BASE_DN=dc=empresa,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=empresa,dc=com
```text

### Active Directory

```bash
LDAP_HOST=ad.empresa.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=DC=empresa,DC=com
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Groups,DC=empresa,DC=com
```text

### FreeIPA

```bash
LDAP_HOST=ipa.empresa.com
LDAP_PORT=389
LDAP_BASE_DN=dc=empresa,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=cn=admins,cn=groups,cn=accounts,dc=empresa,dc=com
```text

---

## 🛠️ Ferramentas Úteis

### Script de Descoberta (Recomendado)

```bash
./ldap-discovery.sh
```text

### Comandos ldapsearch

```bash
# Testar conexão
ldapsearch -H ldap://servidor:389 -x -b "dc=exemplo,dc=com" -D "cn=admin,dc=exemplo,dc=com" -W "(objectClass=*)"

# Buscar usuários
ldapsearch -H ldap://servidor:389 -x -b "dc=exemplo,dc=com" -D "cn=admin,dc=exemplo,dc=com" -W "(uid=usuario)"

# Buscar grupos
ldapsearch -H ldap://servidor:389 -x -b "dc=exemplo,dc=com" -D "cn=admin,dc=exemplo,dc=com" -W "(cn=admins)"
```text

### Ferramentas GUI

- **Apache Directory Studio**: <https://directory.apache.org/studio/>
- **JXplorer**: <http://jxplorer.org/>
- **ldp.exe** (Windows Server - built-in)

---

## 📞 Suporte

Se você está tendo problemas:

1. ✅ Verifique a seção **Troubleshooting** em `LDAP_AUTHENTICATION.md`
2. ✅ Use o script `ldap-discovery.sh` para validar sua configuração
3. ✅ Verifique os logs do servidor (nível DEBUG mostra mais detalhes)
4. ✅ Teste os filtros manualmente com `ldapsearch`

---

## 🔐 Segurança

**Boas práticas:**

- ✅ Use TLS/LDAPS em produção (`LDAP_USE_TLS=true`)
- ✅ Use conta de serviço com permissões mínimas (apenas leitura)
- ✅ Nunca use `LDAP_SKIP_TLS_VERIFY=true` em produção
- ✅ Mantenha senhas no `.env` e adicione `.env` ao `.gitignore`
- ✅ Configure grupos separados para admins e viewers

---

## 📊 Status da Implementação

✅ Autenticação LDAP funcional  
✅ Suporte a múltiplos tipos de servidores LDAP  
✅ Integração com sessões web  
✅ Controle de acesso por grupos (admin/viewer)  
✅ Sincronização de usuários com banco de dados local  
✅ Interface de login unificada (OIDC + LDAP)  
✅ Configuração via variáveis de ambiente  
✅ Documentação completa  
✅ Script de descoberta interativo  

---

## 🚀 Versão

Documentação atualizada para Txlog Server v1.14.0+

---

## Boa configuração! 🎉

Se você tiver dúvidas ou sugestões, abra uma issue no repositório.
