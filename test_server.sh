#!/bin/bash

# Script de test pour l'API HTTP + Serveur + Moniteur
# Ce script teste les fonctionnalités en lançant le serveur et en faisant des requêtes HTTP

echo "🧪 TESTS DE L'API HTTP + SERVEUR + MONITEUR"
echo "============================================"
echo ""

# Couleurs pour les messages
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour tester une requête HTTP
test_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo "Test: $description"
    echo "  $method $url"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data" 2>&1)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "$expected_status" ]; then
        echo -e "  ${GREEN}✅ SUCCÈS${NC}: Code HTTP $http_code (attendu: $expected_status)"
        if [ ! -z "$body" ] && [ "$body" != "null" ]; then
            echo "  Réponse: $body" | head -c 200
            echo ""
        fi
        return 0
    else
        echo -e "  ${RED}❌ ERREUR${NC}: Code HTTP $http_code (attendu: $expected_status)"
        echo "  Réponse: $body"
        return 1
    fi
}

# Vérifier si le serveur est déjà en cours d'exécution
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Le serveur semble déjà être en cours d'exécution${NC}"
    echo ""
else
    echo -e "${RED}❌ Le serveur n'est pas en cours d'exécution${NC}"
    echo "Veuillez démarrer le serveur avec: go run main.go run-server"
    echo "Puis relancez ce script dans un autre terminal"
    exit 1
fi

echo ""
echo "=== Test 1: Health Check ==="
test_request "GET" "http://localhost:8080/health" "" "200" "Health Check"

echo ""
echo "=== Test 2: Création de lien ==="
create_response=$(curl -s -X POST "http://localhost:8080/api/v1/links" \
    -H "Content-Type: application/json" \
    -d '{"long_url":"https://example.com"}')

echo "Réponse: $create_response"
short_code=$(echo "$create_response" | grep -o '"short_code":"[^"]*"' | cut -d'"' -f4)

if [ -z "$short_code" ]; then
    echo -e "${RED}❌ ERREUR: Impossible d'extraire le short_code${NC}"
else
    echo -e "${GREEN}✅ SUCCÈS${NC}: Lien créé avec short_code: $short_code"
    
    echo ""
    echo "=== Test 3: Redirection ==="
    redirect_response=$(curl -s -w "\n%{http_code}" -L "http://localhost:8080/$short_code" 2>&1)
    redirect_code=$(echo "$redirect_response" | tail -n1)
    
    if [ "$redirect_code" == "200" ] || [ "$redirect_code" == "302" ]; then
        echo -e "${GREEN}✅ SUCCÈS${NC}: Redirection fonctionne (code: $redirect_code)"
    else
        echo -e "${RED}❌ ERREUR${NC}: Redirection échouée (code: $redirect_code)"
    fi
    
    echo ""
    echo "=== Test 4: Statistiques ==="
    test_request "GET" "http://localhost:8080/api/v1/links/$short_code/stats" "" "200" "Récupération des statistiques"
    
    echo ""
    echo "=== Test 5: Gestion des erreurs ==="
    test_request "GET" "http://localhost:8080/nonexistent" "" "404" "Lien inexistant (404)"
    test_request "POST" "http://localhost:8080/api/v1/links" '{"long_url":"not-a-url"}' "400" "URL invalide (400)"
    test_request "GET" "http://localhost:8080/api/v1/links/nonexistent/stats" "" "404" "Stats pour lien inexistant (404)"
fi

echo ""
echo "============================================"
echo -e "${GREEN}✅ TOUS LES TESTS SONT TERMINÉS${NC}"
echo ""
echo "💡 Pour vérifier le moniteur d'URLs, consultez les logs du serveur"
echo "   Le moniteur vérifie les URLs toutes les 5 minutes (configurable)"

