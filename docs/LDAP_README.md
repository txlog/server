# Documentação LDAP - Índice Mestre

Este diretório contém documentação completa sobre autenticação LDAP no Txlog Server.

## 📚 Documentação Disponível

### 🚀 Para Começar

1. **[LDAP_AUTHENTICATION.md](LDAP_AUTHENTICATION.md)**  
   📖 Guia completo de autenticação LDAP  
   - Visão geral da funcionalidade
   - Configuração passo a passo
   - Variáveis de ambiente
   - Exemplos práticos
   - Troubleshooting

### 🔍 Descobrindo seus Filtros LDAP

**Cada servidor LDAP é diferente!** Use estes recursos para descobrir os valores corretos:

1. **[LDAP_FILTERS_QUICK.md](LDAP_FILTERS_QUICK.md)** ⭐ **COMECE AQUI**  
   ⚡ Guia rápido e prático  
   - Tabela de referência por tipo de servidor
   - Comandos prontos para usar
   - Valores comuns (OpenLDAP, AD, FreeIPA)
   - Como testar seus filtros

2. **[LDAP_FILTER_DISCOVERY.md](LDAP_FILTER_DISCOVERY.md)**  
   📘 Guia completo e detalhado  
   - Passo a passo para explorar seu LDAP
   - Explicação de cada tipo de filtro
   - Uso de ldapsearch e Apache Directory Studio
   - Troubleshooting avançado

3. **[ldap-discovery.sh](ldap-discovery.sh)** 🛠️ **Script Interativo**  

   ```bash
   chmod +x ldap-discovery.sh
   ./ldap-discovery.sh
   ```

   - Menu interativo para explorar seu LDAP
   - Testa conexão automaticamente
   - Descobre usuários e grupos
   - Testa filtros em tempo real
   - Gera configuração recomendada

### 📋 Referências Rápidas

4. **[LDAP_QUICK_REFERENCE.md](LDAP_QUICK_REFERENCE.md)**  
   📄 Cheatsheet de configuração  
   - Variáveis de ambiente resumidas
   - Exemplos de .env por cenário
   - Comandos úteis

### 🔐 Conta de Serviço

5. **[LDAP_SERVICE_ACCOUNT_FAQ.md](LDAP_SERVICE_ACCOUNT_FAQ.md)**  
   ❓ Perguntas frequentes sobre conta de serviço  
   - Quando usar conta de serviço vs anonymous bind
   - Como criar conta de serviço
   - Permissões necessárias
   - Melhores práticas de segurança

6. **[LDAP_SEM_SERVICE_ACCOUNT.md](LDAP_SEM_SERVICE_ACCOUNT.md)**  
   🔓 Como usar sem conta de serviço (anonymous bind)  
   - Configuração para anonymous bind
   - Limitações e considerações
   - Quando é possível usar

### 🏗️ Informações Técnicas

7. **[LDAP_IMPLEMENTATION_SUMMARY.md](LDAP_IMPLEMENTATION_SUMMARY.md)**  
   🔧 Detalhes de implementação  
   - Arquitetura do código
   - Fluxo de autenticação
   - Estrutura de banco de dados
   - Endpoints da API

### 🚨 Troubleshooting

8. **[LDAP_ERROR_CODES.md](LDAP_ERROR_CODES.md)**  
   🔍 Códigos de erro LDAP explicados  
   - **LDAP Result Code 32: No Such Object** (mais comum)
   - LDAP Result Code 49: Invalid Credentials
   - LDAP Result Code 50: Insufficient Access
   - Como diagnosticar cada erro
   - Soluções práticas

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
