# Script de test PowerShell pour l'API HTTP + Serveur + Moniteur
# Ce script teste les fonctionnalités en lançant le serveur et en faisant des requêtes HTTP

Write-Host "🧪 TESTS DE L'API HTTP + SERVEUR + MONITEUR" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Fonction pour tester une requête HTTP
function Test-Request {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Data = $null,
        [int]$ExpectedStatus,
        [string]$Description
    )
    
    Write-Host "Test: $Description" -ForegroundColor Yellow
    Write-Host "  $Method $Url"
    
    try {
        if ($Data) {
            $response = Invoke-WebRequest -Uri $Url -Method $Method -Body $Data -ContentType "application/json" -UseBasicParsing -ErrorAction Stop
        } else {
            $response = Invoke-WebRequest -Uri $Url -Method $Method -UseBasicParsing -ErrorAction Stop
        }
        
        $statusCode = $response.StatusCode
        
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "  ✅ SUCCÈS: Code HTTP $statusCode (attendu: $ExpectedStatus)" -ForegroundColor Green
            if ($response.Content) {
                $contentPreview = $response.Content.Substring(0, [Math]::Min(200, $response.Content.Length))
                Write-Host "  Réponse: $contentPreview"
            }
            return $true
        } else {
            Write-Host "  ❌ ERREUR: Code HTTP $statusCode (attendu: $ExpectedStatus)" -ForegroundColor Red
            return $false
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "  ✅ SUCCÈS: Code HTTP $statusCode (attendu: $ExpectedStatus)" -ForegroundColor Green
            return $true
        } else {
            Write-Host "  ❌ ERREUR: Code HTTP $statusCode (attendu: $ExpectedStatus)" -ForegroundColor Red
            Write-Host "  Erreur: $($_.Exception.Message)"
            return $false
        }
    }
}

# Vérifier si le serveur est déjà en cours d'exécution
try {
    $healthCheck = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ Le serveur est en cours d'exécution" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Le serveur n'est pas en cours d'exécution" -ForegroundColor Red
    Write-Host "Veuillez démarrer le serveur avec: go run main.go run-server" -ForegroundColor Yellow
    Write-Host "Puis relancez ce script dans un autre terminal" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "=== Test 1: Health Check ===" -ForegroundColor Cyan
Test-Request -Method "GET" -Url "http://localhost:8080/health" -ExpectedStatus 200 -Description "Health Check"

Write-Host ""
Write-Host "=== Test 2: Création de lien ===" -ForegroundColor Cyan
$createBody = '{"long_url":"https://example.com"}' | ConvertTo-Json
try {
    $createResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/links" -Method POST -Body $createBody -ContentType "application/json"
    Write-Host "Réponse: $($createResponse | ConvertTo-Json)" -ForegroundColor Gray
    $shortCode = $createResponse.short_code
    
    if ($shortCode) {
        Write-Host "✅ SUCCÈS: Lien créé avec short_code: $shortCode" -ForegroundColor Green
        
        Write-Host ""
        Write-Host "=== Test 3: Redirection ===" -ForegroundColor Cyan
        try {
            $redirectResponse = Invoke-WebRequest -Uri "http://localhost:8080/$shortCode" -MaximumRedirection 0 -UseBasicParsing -ErrorAction Stop
            Write-Host "  ⚠️  Redirection non suivie (code: $($redirectResponse.StatusCode))" -ForegroundColor Yellow
        } catch {
            $redirectCode = $_.Exception.Response.StatusCode.value__
            if ($redirectCode -eq 302) {
                Write-Host "  ✅ SUCCÈS: Redirection fonctionne (code: $redirectCode)" -ForegroundColor Green
            } else {
                Write-Host "  ❌ ERREUR: Code de redirection inattendu: $redirectCode" -ForegroundColor Red
            }
        }
        
        Write-Host ""
        Write-Host "=== Test 4: Statistiques ===" -ForegroundColor Cyan
        Test-Request -Method "GET" -Url "http://localhost:8080/api/v1/links/$shortCode/stats" -ExpectedStatus 200 -Description "Récupération des statistiques"
        
        Write-Host ""
        Write-Host "=== Test 5: Gestion des erreurs ===" -ForegroundColor Cyan
        try {
            Invoke-WebRequest -Uri "http://localhost:8080/nonexistent" -UseBasicParsing -ErrorAction Stop
            Write-Host "  ❌ ERREUR: Devrait retourner 404" -ForegroundColor Red
        } catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq 404) {
                Write-Host "  ✅ SUCCÈS: Lien inexistant retourne 404" -ForegroundColor Green
            }
        }
        
        $invalidBody = '{"long_url":"not-a-url"}' | ConvertTo-Json
        try {
            Invoke-RestMethod -Uri "http://localhost:8080/api/v1/links" -Method POST -Body $invalidBody -ContentType "application/json" -ErrorAction Stop
            Write-Host "  ❌ ERREUR: Devrait retourner 400 pour URL invalide" -ForegroundColor Red
        } catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq 400) {
                Write-Host "  ✅ SUCCÈS: URL invalide retourne 400" -ForegroundColor Green
            }
        }
        
        try {
            Invoke-WebRequest -Uri "http://localhost:8080/api/v1/links/nonexistent/stats" -UseBasicParsing -ErrorAction Stop
            Write-Host "  ❌ ERREUR: Devrait retourner 404" -ForegroundColor Red
        } catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq 404) {
                Write-Host "  ✅ SUCCÈS: Stats pour lien inexistant retourne 404" -ForegroundColor Green
            }
        }
    } else {
        Write-Host "❌ ERREUR: Impossible d'extraire le short_code" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ ERREUR lors de la création du lien: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "✅ TOUS LES TESTS SONT TERMINÉS" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Pour vérifier le moniteur d'URLs, consultez les logs du serveur" -ForegroundColor Yellow
Write-Host "   Le moniteur vérifie les URLs toutes les 5 minutes (configurable)" -ForegroundColor Yellow

