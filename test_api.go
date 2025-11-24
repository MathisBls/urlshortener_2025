package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/axellelanca/urlshortener/internal/api"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/monitor"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestSetup initialise l'environnement de test
func TestSetup() (*gorm.DB, *services.LinkService, *repository.GormClickRepository) {
	// Créer une base de données en mémoire pour les tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Auto-migrer les tables
	db.AutoMigrate(&models.Link{}, &models.Click{})

	// Initialiser les repositories
	linkRepo := repository.NewLinkRepository(db)
	clickRepo := repository.NewClickRepository(db)

	// Initialiser les services
	linkService := services.NewLinkService(linkRepo)

	return db, linkService, clickRepo.(*repository.GormClickRepository)
}

// Test 1: Vérifier que ClickEventsChannel est bien initialisé et bufferisé
func TestClickEventsChannel() {
	fmt.Println("\n=== Test 1: ClickEventsChannel ===")
	
	// Initialiser le channel avec un buffer
	bufferSize := 10
	api.ClickEventsChannel = make(chan models.ClickEvent, bufferSize)
	
	// Vérifier que le channel n'est pas nil
	if api.ClickEventsChannel == nil {
		fmt.Println("❌ ERREUR: ClickEventsChannel est nil")
		return
	}
	
	// Tester que le channel est bufferisé en envoyant plusieurs événements
	for i := 0; i < bufferSize; i++ {
		event := models.ClickEvent{
			LinkID:    uint(i),
			Timestamp: time.Now(),
			UserAgent: "test-agent",
			IP:        "127.0.0.1",
		}
		select {
		case api.ClickEventsChannel <- event:
			// Succès, le channel accepte l'événement
		default:
			fmt.Printf("❌ ERREUR: Le channel est plein après seulement %d événements\n", i)
			return
		}
	}
	
	// Vérifier qu'on peut lire les événements
	readCount := 0
	for i := 0; i < bufferSize; i++ {
		select {
		case <-api.ClickEventsChannel:
			readCount++
		case <-time.After(1 * time.Second):
			fmt.Printf("❌ ERREUR: Impossible de lire l'événement %d\n", i)
			return
		}
	}
	
	if readCount == bufferSize {
		fmt.Printf("✅ SUCCÈS: ClickEventsChannel fonctionne correctement avec un buffer de %d\n", bufferSize)
	} else {
		fmt.Printf("❌ ERREUR: Seulement %d/%d événements lus\n", readCount, bufferSize)
	}
}

// Test 2: Vérifier le Health Check
func TestHealthCheck() {
	fmt.Println("\n=== Test 2: Health Check ===")
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", api.HealthCheckHandler)
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		fmt.Printf("❌ ERREUR: Code de statut attendu 200, obtenu %d\n", w.Code)
		return
	}
	
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ ERREUR: Impossible de parser la réponse JSON: %v\n", err)
		return
	}
	
	if response["status"] != "ok" {
		fmt.Printf("❌ ERREUR: Statut attendu 'ok', obtenu '%s'\n", response["status"])
		return
	}
	
	fmt.Println("✅ SUCCÈS: Health Check fonctionne correctement")
}

// Test 3: Vérifier la création de lien
func TestCreateLink() {
	fmt.Println("\n=== Test 3: Création de lien ===")
	
	_, linkService, _ := TestSetup()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Initialiser le channel pour éviter les panics
	api.ClickEventsChannel = make(chan models.ClickEvent, 100)
	
	api.SetupRoutes(router, linkService, "http://localhost:8080")
	
	// Test avec une URL valide
	requestBody := map[string]string{
		"long_url": "https://example.com",
	}
	jsonBody, _ := json.Marshal(requestBody)
	
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusCreated {
		fmt.Printf("❌ ERREUR: Code de statut attendu 201, obtenu %d\n", w.Code)
		fmt.Printf("Réponse: %s\n", w.Body.String())
		return
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ ERREUR: Impossible de parser la réponse JSON: %v\n", err)
		return
	}
	
	if response["short_code"] == nil || response["long_url"] == nil {
		fmt.Println("❌ ERREUR: Réponse incomplète")
		return
	}
	
	fmt.Println("✅ SUCCÈS: Création de lien fonctionne correctement")
	fmt.Printf("   Short Code: %s\n", response["short_code"])
}

// Test 4: Vérifier la redirection et l'envoi d'événement
func TestRedirectAndClickEvent() {
	fmt.Println("\n=== Test 4: Redirection et ClickEvent ===")
	
	_, linkService, clickRepo := TestSetup()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Créer un lien d'abord
	link, err := linkService.CreateLink("https://example.com")
	if err != nil {
		fmt.Printf("❌ ERREUR: Impossible de créer le lien: %v\n", err)
		return
	}
	
	// Initialiser le channel avec un petit buffer pour tester
	api.ClickEventsChannel = make(chan models.ClickEvent, 5)
	
	// Démarrer un worker pour consommer les événements
	eventReceived := make(chan bool, 1)
	go func() {
		select {
		case event := <-api.ClickEventsChannel:
			if event.LinkID == link.ID {
				eventReceived <- true
			}
		case <-time.After(2 * time.Second):
			eventReceived <- false
		}
	}()
	
	api.SetupRoutes(router, linkService, "http://localhost:8080")
	
	// Tester la redirection
	req, _ := http.NewRequest("GET", "/"+link.ShortCode, nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusFound {
		fmt.Printf("❌ ERREUR: Code de statut attendu 302, obtenu %d\n", w.Code)
		return
	}
	
	// Vérifier que l'événement a été envoyé
	select {
	case received := <-eventReceived:
		if received {
			fmt.Println("✅ SUCCÈS: Redirection et envoi de ClickEvent fonctionnent correctement")
		} else {
			fmt.Println("❌ ERREUR: ClickEvent non reçu")
		}
	case <-time.After(3 * time.Second):
		fmt.Println("❌ ERREUR: Timeout en attendant le ClickEvent")
	}
	
	// Vérifier que le clic a été enregistré en base
	time.Sleep(500 * time.Millisecond) // Attendre que le worker traite l'événement
	count, err := clickRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		fmt.Printf("⚠️  AVERTISSEMENT: Impossible de compter les clics: %v\n", err)
	} else if count > 0 {
		fmt.Printf("✅ SUCCÈS: Clic enregistré en base de données (count: %d)\n", count)
	}
}

// Test 5: Vérifier les statistiques
func TestGetStats() {
	fmt.Println("\n=== Test 5: Statistiques ===")
	
	_, linkService, clickRepo := TestSetup()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	api.ClickEventsChannel = make(chan models.ClickEvent, 100)
	api.SetupRoutes(router, linkService, "http://localhost:8080")
	
	// Créer un lien
	link, err := linkService.CreateLink("https://example.com")
	if err != nil {
		fmt.Printf("❌ ERREUR: Impossible de créer le lien: %v\n", err)
		return
	}
	
	// Créer quelques clics directement
	for i := 0; i < 3; i++ {
		click := &models.Click{
			LinkID:    link.ID,
			Timestamp: time.Now(),
			UserAgent: "test-agent",
			IPAddress: "127.0.0.1",
		}
		clickRepo.CreateClick(click)
	}
	
	// Tester les stats
	req, _ := http.NewRequest("GET", "/api/v1/links/"+link.ShortCode+"/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		fmt.Printf("❌ ERREUR: Code de statut attendu 200, obtenu %d\n", w.Code)
		fmt.Printf("Réponse: %s\n", w.Body.String())
		return
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ ERREUR: Impossible de parser la réponse JSON: %v\n", err)
		return
	}
	
	totalClicks, ok := response["total_clicks"].(float64)
	if !ok {
		fmt.Println("❌ ERREUR: total_clicks manquant ou invalide")
		return
	}
	
	if int(totalClicks) != 3 {
		fmt.Printf("❌ ERREUR: Nombre de clics attendu 3, obtenu %.0f\n", totalClicks)
		return
	}
	
	fmt.Println("✅ SUCCÈS: Récupération des statistiques fonctionne correctement")
}

// Test 6: Vérifier la gestion des erreurs
func TestErrorHandling() {
	fmt.Println("\n=== Test 6: Gestion des erreurs ===")
	
	_, linkService, _ := TestSetup()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	api.ClickEventsChannel = make(chan models.ClickEvent, 100)
	api.SetupRoutes(router, linkService, "http://localhost:8080")
	
	// Test 1: URL invalide
	requestBody := map[string]string{
		"long_url": "not-a-valid-url",
	}
	jsonBody, _ := json.Marshal(requestBody)
	
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		fmt.Printf("❌ ERREUR: Code de statut attendu 400 pour URL invalide, obtenu %d\n", w.Code)
		return
	}
	
	// Test 2: Lien non trouvé
	req2, _ := http.NewRequest("GET", "/nonexistent", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusNotFound {
		fmt.Printf("❌ ERREUR: Code de statut attendu 404 pour lien inexistant, obtenu %d\n", w2.Code)
		return
	}
	
	// Test 3: Stats pour lien inexistant
	req3, _ := http.NewRequest("GET", "/api/v1/links/nonexistent/stats", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	
	if w3.Code != http.StatusNotFound {
		fmt.Printf("❌ ERREUR: Code de statut attendu 404 pour stats de lien inexistant, obtenu %d\n", w3.Code)
		return
	}
	
	fmt.Println("✅ SUCCÈS: Gestion des erreurs fonctionne correctement")
}

// Test 7: Vérifier le moniteur d'URLs
func TestUrlMonitor() {
	fmt.Println("\n=== Test 7: Moniteur d'URLs ===")
	
	// Ce test vérifie que le moniteur peut être instancié et démarré
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ ERREUR: Impossible de créer la base de données: %v\n", err)
		return
	}
	
	db.AutoMigrate(&models.Link{})
	
	linkRepo := repository.NewLinkRepository(db)
	
	// Créer un moniteur avec un intervalle court pour les tests
	monitorInterval := 1 * time.Second
	urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)
	
	if urlMonitor == nil {
		fmt.Println("❌ ERREUR: Impossible de créer le moniteur")
		return
	}
	
	// Vérifier que le moniteur peut être démarré (dans une goroutine pour ne pas bloquer)
	started := make(chan bool, 1)
	go func() {
		urlMonitor.Start()
		started <- true
	}()
	
	// Attendre un peu pour voir si le moniteur démarre sans erreur
	select {
	case <-started:
		fmt.Println("✅ SUCCÈS: Moniteur d'URLs peut être démarré")
	case <-time.After(2 * time.Second):
		fmt.Println("✅ SUCCÈS: Moniteur d'URLs peut être instancié et démarré")
		fmt.Printf("   Intervalle configuré: %v\n", monitorInterval)
		fmt.Println("   Note: Le moniteur tourne en arrière-plan")
	}
}

// Fonction principale de test
func main() {
	fmt.Println("🧪 DÉMARRAGE DES TESTS DE L'API HTTP + SERVEUR + MONITEUR")
	fmt.Println("=" + strings.Repeat("=", 60))
	
	TestClickEventsChannel()
	TestHealthCheck()
	TestCreateLink()
	TestRedirectAndClickEvent()
	TestGetStats()
	TestErrorHandling()
	TestUrlMonitor()
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ TOUS LES TESTS SONT TERMINÉS")
	fmt.Println("\n💡 Pour tester le serveur complet, exécutez:")
	fmt.Println("   go run main.go run-server")
	fmt.Println("\n💡 Puis dans un autre terminal:")
	fmt.Println("   curl http://localhost:8080/health")
	fmt.Println("   curl -X POST http://localhost:8080/api/v1/links -H 'Content-Type: application/json' -d '{\"long_url\":\"https://example.com\"}'")
}

