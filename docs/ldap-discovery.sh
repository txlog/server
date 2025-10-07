#!/bin/bash
#
# LDAP Discovery Script
# Script auxiliar para descobrir filtros LDAP
#
# Uso: ./ldap-discovery.sh
#

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     LDAP Discovery Script - Txlog Server          ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""

# Verificar se ldapsearch está instalado
if ! command -v ldapsearch &> /dev/null; then
    echo -e "${RED}❌ ldapsearch não encontrado!${NC}"
    echo ""
    echo "Instale o pacote ldap-utils:"
    echo "  - Debian/Ubuntu: sudo apt-get install ldap-utils"
    echo "  - Red Hat/CentOS: sudo yum install openldap-clients"
    echo "  - Mac: brew install openldap"
    echo ""
    exit 1
fi

# Coletar informações do servidor LDAP
echo -e "${YELLOW}📝 Configuração do Servidor LDAP${NC}"
echo ""

read -p "Host do servidor LDAP (ex: ldap.exemplo.com): " LDAP_HOST
read -p "Porta (389 para LDAP, 636 para LDAPS) [389]: " LDAP_PORT
LDAP_PORT=${LDAP_PORT:-389}

read -p "Usar TLS/LDAPS? (s/n) [n]: " USE_TLS
USE_TLS=${USE_TLS:-n}

if [[ $USE_TLS == "s" || $USE_TLS == "S" ]]; then
    LDAP_URL="ldaps://${LDAP_HOST}:${LDAP_PORT}"
else
    LDAP_URL="ldap://${LDAP_HOST}:${LDAP_PORT}"
fi

read -p "Base DN (ex: dc=exemplo,dc=com): " BASE_DN
read -p "Bind DN (ex: cn=admin,dc=exemplo,dc=com): " BIND_DN
read -s -p "Senha do Bind DN: " BIND_PASSWORD
echo ""
echo ""

# Testar conexão
echo -e "${YELLOW}🔍 Testando conexão...${NC}"
if ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -s base "(objectClass=*)" dn &> /dev/null; then
    echo -e "${GREEN}✅ Conexão bem-sucedida!${NC}"
else
    echo -e "${RED}❌ Falha na conexão. Verifique as credenciais.${NC}"
    exit 1
fi
echo ""

# Menu principal
while true; do
    echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                 Menu Principal                    ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "1) Explorar estrutura do diretório"
    echo "2) Buscar usuários"
    echo "3) Buscar grupos"
    echo "4) Testar filtro de usuário"
    echo "5) Testar filtro de grupo"
    echo "6) Ver configuração recomendada"
    echo "0) Sair"
    echo ""
    read -p "Escolha uma opção: " option
    echo ""

    case $option in
        1)
            echo -e "${YELLOW}📂 Estrutura do Diretório${NC}"
            echo ""
            echo "OUs (Organizational Units):"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(objectClass=organizationalUnit)" dn | grep "^dn:" | sed 's/^dn: /  - /'
            echo ""
            ;;
        
        2)
            echo -e "${YELLOW}👥 Buscar Usuários${NC}"
            echo ""
            echo "Escolha o tipo de busca:"
            echo "1) Por objectClass=person"
            echo "2) Por objectClass=inetOrgPerson"
            echo "3) Por objectClass=posixAccount"
            echo "4) Por objectClass=user (Active Directory)"
            echo ""
            read -p "Opção: " search_type
            
            case $search_type in
                1) filter="(objectClass=person)" ;;
                2) filter="(objectClass=inetOrgPerson)" ;;
                3) filter="(objectClass=posixAccount)" ;;
                4) filter="(objectClass=user)" ;;
                *) echo -e "${RED}Opção inválida${NC}"; continue ;;
            esac
            
            echo ""
            echo "Primeiros 5 usuários encontrados:"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$filter" dn uid cn sAMAccountName mail | head -n 30
            echo ""
            
            read -p "Ver detalhes de um usuário específico? (s/n): " view_user
            if [[ $view_user == "s" || $view_user == "S" ]]; then
                read -p "Digite o username/uid: " username
                echo ""
                echo "Tentando buscar com diferentes atributos..."
                
                for attr in uid cn sAMAccountName mail; do
                    echo -e "${BLUE}Buscando com $attr=$username:${NC}"
                    result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "($attr=$username)" 2>/dev/null)
                    if [ ! -z "$result" ]; then
                        echo "$result"
                        echo ""
                        echo -e "${GREEN}✅ Encontrado com filtro: ($attr=%s)${NC}"
                        echo ""
                        break
                    fi
                done
            fi
            ;;
        
        3)
            echo -e "${YELLOW}👥 Buscar Grupos${NC}"
            echo ""
            echo "Escolha o tipo de busca:"
            echo "1) Por objectClass=groupOfNames"
            echo "2) Por objectClass=groupOfUniqueNames"
            echo "3) Por objectClass=posixGroup"
            echo "4) Por objectClass=group (Active Directory)"
            echo ""
            read -p "Opção: " search_type
            
            case $search_type in
                1) filter="(objectClass=groupOfNames)" ;;
                2) filter="(objectClass=groupOfUniqueNames)" ;;
                3) filter="(objectClass=posixGroup)" ;;
                4) filter="(objectClass=group)" ;;
                *) echo -e "${RED}Opção inválida${NC}"; continue ;;
            esac
            
            echo ""
            echo "Grupos encontrados:"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$filter" dn cn
            echo ""
            
            read -p "Ver detalhes de um grupo específico? (s/n): " view_group
            if [[ $view_group == "s" || $view_group == "S" ]]; then
                read -p "Digite o CN do grupo: " groupname
                echo ""
                ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(cn=$groupname)" dn cn member uniqueMember memberUid
                echo ""
                
                echo -e "${BLUE}Identificando tipo de atributo de membro:${NC}"
                result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(cn=$groupname)" member uniqueMember memberUid 2>/dev/null)
                
                if echo "$result" | grep -q "^member:"; then
                    echo -e "${GREEN}✅ Usa 'member' - Filtro recomendado: (member=%s)${NC}"
                elif echo "$result" | grep -q "^uniqueMember:"; then
                    echo -e "${GREEN}✅ Usa 'uniqueMember' - Filtro recomendado: (uniqueMember=%s)${NC}"
                elif echo "$result" | grep -q "^memberUid:"; then
                    echo -e "${GREEN}✅ Usa 'memberUid' - Filtro recomendado: (memberUid=%s)${NC}"
                    echo -e "${YELLOW}⚠️  Atenção: memberUid usa apenas o uid, não o DN completo${NC}"
                fi
                echo ""
            fi
            ;;
        
        4)
            echo -e "${YELLOW}🧪 Testar Filtro de Usuário${NC}"
            echo ""
            echo "Filtros comuns:"
            echo "  (uid=%s)                - OpenLDAP, FreeIPA"
            echo "  (sAMAccountName=%s)     - Active Directory"
            echo "  (cn=%s)                 - Sistemas antigos"
            echo "  (mail=%s)               - Login por email"
            echo ""
            read -p "Digite o filtro (ex: (uid=%s)): " user_filter
            read -p "Digite um username para testar: " test_user
            
            test_filter=$(echo "$user_filter" | sed "s/%s/$test_user/")
            echo ""
            echo -e "${BLUE}Testando filtro: $test_filter${NC}"
            
            error_output=$(mktemp)
            result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$test_filter" 2>"$error_output")
            exit_code=$?
            
            if [ $exit_code -ne 0 ]; then
                echo -e "${RED}❌ Erro ao executar ldapsearch:${NC}"
                cat "$error_output"
                rm -f "$error_output"
            elif [ ! -z "$result" ]; then
                echo "$result"
                echo ""
                count=$(echo "$result" | grep -c "^dn:" || true)
                if [ "$count" -eq 1 ]; then
                    echo -e "${GREEN}✅ Filtro OK! Encontrou exatamente 1 usuário.${NC}"
                elif [ "$count" -gt 1 ]; then
                    echo -e "${YELLOW}⚠️  Filtro encontrou $count usuários. Deve retornar apenas 1!${NC}"
                else
                    echo -e "${RED}❌ Nenhum usuário encontrado.${NC}"
                fi
                rm -f "$error_output"
            else
                echo -e "${RED}❌ Nenhum usuário encontrado.${NC}"
                rm -f "$error_output"
            fi
            echo ""
            ;;
        
        5)
            echo -e "${YELLOW}🧪 Testar Filtro de Grupo${NC}"
            echo ""
            echo "Filtros comuns:"
            echo "  (member=%s)        - groupOfNames, Active Directory"
            echo "  (uniqueMember=%s)  - groupOfUniqueNames"
            echo "  (memberUid=%s)     - posixGroup (usa só uid, não DN)"
            echo ""
            read -p "Digite o filtro (ex: (member=%s)): " group_filter
            read -p "Digite o DN do grupo para testar: " test_group_dn
            read -p "Digite o DN do usuário: " test_user_dn
            
            test_filter=$(echo "$group_filter" | sed "s/%s/$test_user_dn/")
            echo ""
            echo -e "${BLUE}Testando se usuário pertence ao grupo...${NC}"
            echo -e "${BLUE}Grupo: $test_group_dn${NC}"
            echo -e "${BLUE}Filtro: $test_filter${NC}"
            
            error_output=$(mktemp)
            result=$(ldapsearch -H "$LDAP_URL" -x -b "$test_group_dn" -D "$BIND_DN" -w "$BIND_PASSWORD" -s base -LLL "$test_filter" dn 2>"$error_output")
            exit_code=$?
            
            if [ $exit_code -ne 0 ]; then
                echo -e "${RED}❌ Erro ao executar ldapsearch:${NC}"
                cat "$error_output"
                rm -f "$error_output"
            elif [ ! -z "$result" ]; then
                echo "$result"
                echo ""
                echo -e "${GREEN}✅ Usuário é membro do grupo!${NC}"
                rm -f "$error_output"
            else
                echo -e "${RED}❌ Usuário NÃO é membro do grupo, ou filtro incorreto.${NC}"
                rm -f "$error_output"
            fi
            echo ""
            ;;
        
        6)
            echo -e "${YELLOW}📋 Configuração Recomendada${NC}"
            echo ""
            echo "Com base nas informações coletadas, adicione ao seu .env:"
            echo ""
            echo "LDAP_HOST=$LDAP_HOST"
            echo "LDAP_PORT=$LDAP_PORT"
            if [[ $USE_TLS == "s" || $USE_TLS == "S" ]]; then
                echo "LDAP_USE_TLS=true"
            else
                echo "LDAP_USE_TLS=false"
            fi
            echo "LDAP_BASE_DN=$BASE_DN"
            echo "LDAP_BIND_DN=$BIND_DN"
            echo "LDAP_BIND_PASSWORD=sua_senha_aqui"
            echo ""
            echo "# Descubra estes valores usando as opções 2, 3, 4 e 5 do menu:"
            echo "LDAP_USER_FILTER=(uid=%s)           # ou (sAMAccountName=%s) para AD"
            echo "LDAP_GROUP_FILTER=(member=%s)       # ou (uniqueMember=%s) ou (memberUid=%s)"
            echo "LDAP_ADMIN_GROUP=cn=admins,ou=groups,$BASE_DN"
            echo "LDAP_VIEWER_GROUP=cn=viewers,ou=groups,$BASE_DN"
            echo ""
            ;;
        
        0)
            echo -e "${GREEN}👋 Até logo!${NC}"
            exit 0
            ;;
        
        *)
            echo -e "${RED}Opção inválida!${NC}"
            ;;
    esac
    
    read -p "Pressione ENTER para continuar..."
    echo ""
done
