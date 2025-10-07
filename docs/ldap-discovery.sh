#!/bin/bash
#
# LDAP Discovery Script
# Helper script to discover LDAP filters
#
# Usage: ./ldap-discovery.sh
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     LDAP Discovery Script - Txlog Server          ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""

# Check if ldapsearch is installed
if ! command -v ldapsearch &> /dev/null; then
    echo -e "${RED}❌ ldapsearch not found!${NC}"
    echo ""
    echo "Install the ldap-utils package:"
    echo "  - Debian/Ubuntu: sudo apt-get install ldap-utils"
    echo "  - Red Hat/CentOS: sudo yum install openldap-clients"
    echo "  - Mac: brew install openldap"
    echo ""
    exit 1
fi

# Collect LDAP server information
echo -e "${YELLOW}📝 LDAP Server Configuration${NC}"
echo ""

read -p "LDAP server host (e.g., ldap.example.com): " LDAP_HOST
read -p "Port (389 for LDAP, 636 for LDAPS) [389]: " LDAP_PORT
LDAP_PORT=${LDAP_PORT:-389}

read -p "Use TLS/LDAPS? (y/n) [n]: " USE_TLS
USE_TLS=${USE_TLS:-n}

if [[ $USE_TLS == "y" || $USE_TLS == "Y" ]]; then
    LDAP_URL="ldaps://${LDAP_HOST}:${LDAP_PORT}"
else
    LDAP_URL="ldap://${LDAP_HOST}:${LDAP_PORT}"
fi

read -p "Base DN (e.g., dc=example,dc=com): " BASE_DN
read -p "Bind DN (e.g., cn=admin,dc=example,dc=com): " BIND_DN
read -s -p "Bind DN password: " BIND_PASSWORD
echo ""
echo ""

# Test connection
echo -e "${YELLOW}🔍 Testing connection...${NC}"
if ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -s base "(objectClass=*)" dn &> /dev/null; then
    echo -e "${GREEN}✅ Connection successful!${NC}"
else
    echo -e "${RED}❌ Connection failed. Check credentials.${NC}"
    exit 1
fi
echo ""

# Main menu
while true; do
    echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                   Main Menu                       ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "1) Explore directory structure"
    echo "2) Search users"
    echo "3) Search groups"
    echo "4) Test user filter"
    echo "5) Test group filter"
    echo "6) View recommended configuration"
    echo "0) Exit"
    echo ""
    read -p "Choose an option: " option
    echo ""

    case $option in
        1)
            echo -e "${YELLOW}📂 Directory Structure${NC}"
            echo ""
            echo "OUs (Organizational Units):"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(objectClass=organizationalUnit)" dn | grep "^dn:" | sed 's/^dn: /  - /'
            echo ""
            ;;
        
        2)
            echo -e "${YELLOW}👥 Search Users${NC}"
            echo ""
            echo "Choose search type:"
            echo "1) By objectClass=person"
            echo "2) By objectClass=inetOrgPerson"
            echo "3) By objectClass=posixAccount"
            echo "4) By objectClass=user (Active Directory)"
            echo ""
            read -p "Option: " search_type
            
            case $search_type in
                1) filter="(objectClass=person)" ;;
                2) filter="(objectClass=inetOrgPerson)" ;;
                3) filter="(objectClass=posixAccount)" ;;
                4) filter="(objectClass=user)" ;;
                *) echo -e "${RED}Invalid option${NC}"; continue ;;
            esac
            
            echo ""
            echo "First 5 users found:"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$filter" dn uid cn sAMAccountName mail | head -n 30
            echo ""
            
            read -p "View details of a specific user? (y/n): " view_user
            if [[ $view_user == "y" || $view_user == "Y" ]]; then
                read -p "Enter username/uid: " username
                echo ""
                echo "Trying to search with different attributes..."
                
                for attr in uid cn sAMAccountName mail; do
                    echo -e "${BLUE}Searching with $attr=$username:${NC}"
                    result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "($attr=$username)" 2>/dev/null)
                    if [ ! -z "$result" ]; then
                        echo "$result"
                        echo ""
                        echo -e "${GREEN}✅ Found with filter: ($attr=%s)${NC}"
                        echo ""
                        break
                    fi
                done
            fi
            ;;
        
        3)
            echo -e "${YELLOW}👥 Search Groups${NC}"
            echo ""
            echo "Choose search type:"
            echo "1) By objectClass=groupOfNames"
            echo "2) By objectClass=groupOfUniqueNames"
            echo "3) By objectClass=posixGroup"
            echo "4) By objectClass=group (Active Directory)"
            echo ""
            read -p "Option: " search_type
            
            case $search_type in
                1) filter="(objectClass=groupOfNames)" ;;
                2) filter="(objectClass=groupOfUniqueNames)" ;;
                3) filter="(objectClass=posixGroup)" ;;
                4) filter="(objectClass=group)" ;;
                *) echo -e "${RED}Invalid option${NC}"; continue ;;
            esac
            
            echo ""
            echo "Groups found:"
            ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$filter" dn cn
            echo ""
            
            read -p "View details of a specific group? (y/n): " view_group
            if [[ $view_group == "y" || $view_group == "Y" ]]; then
                read -p "Enter group CN: " groupname
                echo ""
                ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(cn=$groupname)" dn cn member uniqueMember memberUid
                echo ""
                
                echo -e "${BLUE}Identifying member attribute type:${NC}"
                result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "(cn=$groupname)" member uniqueMember memberUid 2>/dev/null)
                
                if echo "$result" | grep -q "^member:"; then
                    echo -e "${GREEN}✅ Uses 'member' - Recommended filter: (member=%s)${NC}"
                elif echo "$result" | grep -q "^uniqueMember:"; then
                    echo -e "${GREEN}✅ Uses 'uniqueMember' - Recommended filter: (uniqueMember=%s)${NC}"
                elif echo "$result" | grep -q "^memberUid:"; then
                    echo -e "${GREEN}✅ Uses 'memberUid' - Recommended filter: (memberUid=%s)${NC}"
                    echo -e "${YELLOW}⚠️  Note: memberUid uses only the uid, not the full DN${NC}"
                fi
                echo ""
            fi
            ;;
        
        4)
            echo -e "${YELLOW}🧪 Test User Filter${NC}"
            echo ""
            echo "Common filters:"
            echo "  (uid=%s)                - OpenLDAP, FreeIPA"
            echo "  (sAMAccountName=%s)     - Active Directory"
            echo "  (cn=%s)                 - Legacy systems"
            echo "  (mail=%s)               - Email login"
            echo ""
            read -p "Enter filter (e.g., (uid=%s)): " user_filter
            read -p "Enter a username to test: " test_user
            
            test_filter=$(echo "$user_filter" | sed "s/%s/$test_user/")
            echo ""
            echo -e "${BLUE}Testing filter: $test_filter${NC}"
            
            error_output=$(mktemp)
            result=$(ldapsearch -H "$LDAP_URL" -x -b "$BASE_DN" -D "$BIND_DN" -w "$BIND_PASSWORD" -LLL "$test_filter" 2>"$error_output")
            exit_code=$?
            
            if [ $exit_code -ne 0 ]; then
                echo -e "${RED}❌ Error executing ldapsearch:${NC}"
                cat "$error_output"
                rm -f "$error_output"
            elif [ ! -z "$result" ]; then
                echo "$result"
                echo ""
                count=$(echo "$result" | grep -c "^dn:" || true)
                if [ "$count" -eq 1 ]; then
                    echo -e "${GREEN}✅ Filter OK! Found exactly 1 user.${NC}"
                elif [ "$count" -gt 1 ]; then
                    echo -e "${YELLOW}⚠️  Filter found $count users. Should return only 1!${NC}"
                else
                    echo -e "${RED}❌ No user found.${NC}"
                fi
                rm -f "$error_output"
            else
                echo -e "${RED}❌ No user found.${NC}"
                rm -f "$error_output"
            fi
            echo ""
            ;;
        
        5)
            echo -e "${YELLOW}🧪 Test Group Filter${NC}"
            echo ""
            echo "Common filters:"
            echo "  (member=%s)        - groupOfNames, Active Directory"
            echo "  (uniqueMember=%s)  - groupOfUniqueNames"
            echo "  (memberUid=%s)     - posixGroup (uses only uid, not DN)"
            echo ""
            read -p "Enter filter (e.g., (member=%s)): " group_filter
            read -p "Enter group DN to test: " test_group_dn
            read -p "Enter user DN: " test_user_dn
            
            test_filter=$(echo "$group_filter" | sed "s/%s/$test_user_dn/")
            echo ""
            echo -e "${BLUE}Testing if user belongs to group...${NC}"
            echo -e "${BLUE}Group: $test_group_dn${NC}"
            echo -e "${BLUE}Filter: $test_filter${NC}"
            
            error_output=$(mktemp)
            result=$(ldapsearch -H "$LDAP_URL" -x -b "$test_group_dn" -D "$BIND_DN" -w "$BIND_PASSWORD" -s base -LLL "$test_filter" dn 2>"$error_output")
            exit_code=$?
            
            if [ $exit_code -ne 0 ]; then
                echo -e "${RED}❌ Error executing ldapsearch:${NC}"
                cat "$error_output"
                rm -f "$error_output"
            elif [ ! -z "$result" ]; then
                echo "$result"
                echo ""
                echo -e "${GREEN}✅ User is a member of the group!${NC}"
                rm -f "$error_output"
            else
                echo -e "${RED}❌ User is NOT a member of the group, or filter is incorrect.${NC}"
                rm -f "$error_output"
            fi
            echo ""
            ;;
        
        6)
            echo -e "${YELLOW}📋 Recommended Configuration${NC}"
            echo ""
            echo "Based on the information collected, add to your .env:"
            echo ""
            echo "LDAP_HOST=$LDAP_HOST"
            echo "LDAP_PORT=$LDAP_PORT"
            if [[ $USE_TLS == "y" || $USE_TLS == "Y" ]]; then
                echo "LDAP_USE_TLS=true"
            else
                echo "LDAP_USE_TLS=false"
            fi
            echo "LDAP_BASE_DN=$BASE_DN"
            echo "LDAP_BIND_DN=$BIND_DN"
            echo "LDAP_BIND_PASSWORD=your_password_here"
            echo ""
            echo "# Discover these values using menu options 2, 3, 4, and 5:"
            echo "LDAP_USER_FILTER=(uid=%s)           # or (sAMAccountName=%s) for AD"
            echo "LDAP_GROUP_FILTER=(member=%s)       # or (uniqueMember=%s) or (memberUid=%s)"
            echo "LDAP_ADMIN_GROUP=cn=admins,ou=groups,$BASE_DN"
            echo "LDAP_VIEWER_GROUP=cn=viewers,ou=groups,$BASE_DN"
            echo ""
            ;;
        
        0)
            echo -e "${GREEN}👋 Goodbye!${NC}"
            exit 0
            ;;
        
        *)
            echo -e "${RED}Invalid option!${NC}"
            ;;
    esac
    
    read -p "Press ENTER to continue..."
    echo ""
done
