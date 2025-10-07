# LDAP Service Account - Perguntas Frequentes

## É obrigatório usar service account?

**NÃO!** O service account é **opcional**. O Txlog Server funciona perfeitamente
sem ele em muitos cenários.

## Quando eu NÃO preciso de service account?

Você pode autenticar **SEM service account** quando:

1. **OpenLDAP com anonymous bind habilitado**
   - É a configuração padrão do OpenLDAP
   - Permite buscas sem autenticação
   - Usuários autenticados podem ler seus próprios grupos

2. **Ambiente de desenvolvimento/teste**
   - Configuração mais rápida
   - Menos credenciais para gerenciar
   - Facilita testes

3. **LDAP com ACLs permissivas**
   - Permite anonymous read em users e groups
   - Permite usuários lerem seus próprios grupos

## Quando eu PRECISO de service account?

Você **precisa** de service account quando:

1. **Active Directory**
   - AD geralmente bloqueia anonymous bind
   - Requer autenticação para buscas
   - Política de segurança padrão da Microsoft

2. **OpenLDAP com ACLs restritas**
   - Anonymous bind desabilitado
   - Usuários não podem ler grupos
   - Políticas de segurança corporativas

3. **Requisitos de compliance**
   - Auditoria de acessos
   - Rastreamento de quem faz buscas
   - Políticas de segurança da empresa

## Como funciona SEM service account?

```text
Fluxo de autenticação:

1. Usuário digita username + senha no Txlog
   ↓
2. Txlog conecta ao LDAP (sem autenticação)
   ↓
3. Busca usuário via anonymous bind
   ↓
4. Autentica usuário com bind usando suas credenciais
   ↓
5. Verifica grupos usando a sessão autenticada do usuário
   ↓
6. Cria sessão no Txlog
```

## Como funciona COM service account?

```text
Fluxo de autenticação:

1. Usuário digita username + senha no Txlog
   ↓
2. Txlog conecta ao LDAP
   ↓
3. Txlog faz bind com service account
   ↓
4. Busca usuário usando service account
   ↓
5. Autentica usuário com bind usando credenciais do usuário
   ↓
6. Re-bind com service account
   ↓
7. Verifica grupos usando service account
   ↓
8. Cria sessão no Txlog
```

## Qual é mais seguro?

**Depende do seu ambiente:**

### COM Service Account é mais seguro quando

- ✅ Você precisa rastrear todos os acessos LDAP
- ✅ Você quer limitar exatamente quais objetos podem ser lidos
- ✅ Você quer desabilitar anonymous bind (boa prática)
- ✅ Você tem compliance/auditoria

### SEM Service Account pode ser igualmente seguro quando

- ✅ Anonymous bind só permite leitura (não escrita)
- ✅ ACLs LDAP estão bem configuradas
- ✅ Você está em rede privada/confiável
- ✅ Você tem outros controles de segurança

## Qual é mais fácil de configurar?

**SEM service account** é muito mais simples:

```bash
# Apenas 4 variáveis!
LDAP_HOST=ldap.exemplo.com
LDAP_BASE_DN=ou=users,dc=exemplo,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=exemplo,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=exemplo,dc=com
```

vs

```bash
# COM service account: 6 variáveis
LDAP_HOST=ldap.exemplo.com
LDAP_BIND_DN=cn=svc-txlog,dc=exemplo,dc=com      # +1
LDAP_BIND_PASSWORD=senha_secreta                  # +2
LDAP_BASE_DN=ou=users,dc=exemplo,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=exemplo,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=exemplo,dc=com
```

## Como testar qual opção funciona para mim?

### Teste 1: Anonymous bind funciona?

```bash
ldapsearch -H ldap://seu-ldap:389 -x \
  -b "ou=users,dc=exemplo,dc=com" \
  "(uid=seuusuario)"
```

- **Funciona?** → Você pode usar SEM service account
- **Erro de acesso?** → Você precisa de service account

### Teste 2: Usuário pode ler grupos?

```bash
ldapsearch -H ldap://seu-ldap:389 \
  -D "uid=seuusuario,ou=users,dc=exemplo,dc=com" \
  -w "suasenha" \
  -b "cn=admins,ou=groups,dc=exemplo,dc=com"
```

- **Retorna grupos?** → Verificação funcionará
- **Erro?** → Precisa de service account com permissões

## Recomendações por tipo de servidor

| Servidor LDAP | Recomendação | Motivo |
|---------------|--------------|---------|
| **OpenLDAP** (padrão) | ✅ SEM service account | Anonymous bind habilitado por padrão |
| **OpenLDAP** (hardened) | ⚠️ COM service account | Anonymous bind desabilitado |
| **Active Directory** | ⚠️ COM service account | Requer autenticação para buscas |
| **FreeIPA** | ⚠️ COM service account | Políticas mais restritivas |
| **389 Directory** | ✅ SEM service account | Geralmente permite anonymous |

## Posso mudar depois?

**SIM!** Você pode:

1. **Começar SEM service account**
   - Testar se funciona
   - Se funcionar, deixar assim
   - Se não funcionar, adicionar service account

2. **Começar COM service account**
   - Funciona em qualquer cenário
   - Remover depois se quiser simplificar

**Não há impacto nos usuários** - é apenas configuração do servidor.

## Exemplo prático: Minha primeira configuração

### Passo 1: Comece simples (SEM service account)

```bash
# .env
LDAP_HOST=ldap.minhaempresa.com
LDAP_BASE_DN=ou=users,dc=minhaempresa,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=minhaempresa,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=minhaempresa,dc=com
```

### Passo 2: Teste o login

- ✅ **Funcionou?** Pronto! Deixe assim.
- ❌ **Erro "user not found"?** Adicione service account (Passo 3)

### Passo 3: Se necessário, adicione service account

```bash
# .env (adicione estas 2 linhas)
LDAP_BIND_DN=cn=readonly,dc=minhaempresa,dc=com
LDAP_BIND_PASSWORD=senha_da_conta_servico
```

## Segurança: Service Account vs Anonymous Bind

### Service Account

**Vantagens de segurança:**

- ✅ Logs mostram qual conta fez cada busca
- ✅ Pode auditar acessos específicos
- ✅ Pode revogar acesso facilmente
- ✅ Pode limitar exatamente o que é acessível

**Desvantagens:**

- ❌ Mais uma credencial para proteger
- ❌ Senha pode vazar
- ❌ Precisa gerenciar rotação de senha

### Anonymous Bind

**Vantagens de segurança:**

- ✅ Nenhuma credencial para vazar
- ✅ Nenhuma senha para gerenciar
- ✅ Mais simples = menos chance de erro

**Desvantagens:**

- ❌ Mais difícil de auditar acessos
- ❌ Qualquer um pode fazer buscas
- ❌ Pode não atender políticas corporativas

## Conclusão

| Critério | SEM Service Account | COM Service Account |
|----------|---------------------|---------------------|
| **Simplicidade** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Segurança** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Compatibilidade** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Auditoria** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Manutenção** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

**Resposta rápida:**

- 🏠 **Homelab/Desenvolvimento?** → SEM service account
- 🏢 **Produção/Empresa?** → COM service account
- 💼 **Active Directory?** → COM service account (obrigatório)
- 🐧 **OpenLDAP simples?** → SEM service account
- 📋 **Tem compliance?** → COM service account

## Precisa de ajuda?

1. Consulte [LDAP_SEM_SERVICE_ACCOUNT.md](LDAP_SEM_SERVICE_ACCOUNT.md) para guia completo
2. Veja [LDAP_QUICK_REFERENCE.md](LDAP_QUICK_REFERENCE.md) para exemplos rápidos
3. Leia [LDAP_AUTHENTICATION.md](LDAP_AUTHENTICATION.md) para documentação completa
