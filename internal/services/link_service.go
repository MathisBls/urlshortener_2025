package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm" // Nécessaire pour la gestion spécifique de gorm.ErrRecordNotFound

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository" // Importe le package repository
)

// Définition du jeu de caractères pour la génération des codes courts.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// LinkService est une structure qui fournit des méthodes pour la logique métier des liens.
// Elle détient linkRepo qui est une référence vers une interface LinkRepository.
// IMPORTANT : Le champ doit être du type de l'interface (non-pointeur).
type LinkService struct {
	linkRepo repository.LinkRepository
}

// NewLinkService crée et retourne une nouvelle instance de LinkService.
func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

// GenerateShortCode génère un code court aléatoire d'une longueur spécifiée.
// Elle utilise le package 'crypto/rand' pour éviter la prévisibilité.
func (s *LinkService) GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		length = 6 // Longueur par défaut
	}

	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// CreateLink crée un nouveau lien raccourci.
// Il génère un code court unique, puis persiste le lien dans la base de données.
func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {
	var shortCode string
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		// Génère un code de 6 caractères
		code, err := s.GenerateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		// Vérifie si le code généré existe déjà en base de données
		_, err = s.linkRepo.GetLinkByShortCode(code)

		if err != nil {
			// Si l'erreur est 'record not found' de GORM, cela signifie que le code est unique.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shortCode = code // Le code est unique, on peut l'utiliser
				break            // Sort de la boucle de retry
			}
			// Si c'est une autre erreur de base de données, retourne l'erreur.
			return nil, fmt.Errorf("database error checking short code uniqueness: %w", err)
		}

		// Si aucune erreur (le code a été trouvé), cela signifie une collision.
		log.Printf("Short code '%s' already exists, retrying generation (%d/%d)...", code, i+1, maxRetries)
		// La boucle continuera pour générer un nouveau code.
	}

	// Si après toutes les tentatives, aucun code unique n'a été trouvé
	if shortCode == "" {
		return nil, errors.New("unable to generate unique shortcode after maximum retries")
	}

	// Crée une nouvelle instance du modèle Link.
	link := &models.Link{
		LongURL:   longURL,
		ShortCode: shortCode,
		CreatedAt: time.Now(),
	}

	// Persiste le nouveau lien dans la base de données via le repository
	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	// Retourne le lien créé
	return link, nil
}

// GetLinkByShortCode récupère un lien via son code court.
// Il délègue l'opération de recherche au repository.
func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	return s.linkRepo.GetLinkByShortCode(shortCode)
}

// GetLinkStats récupère les statistiques pour un lien donné (nombre total de clics).
// Il interagit avec le LinkRepository pour obtenir le lien, puis compte les clics.
func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	// Récupérer le lien par son shortCode
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, 0, err
	}

	// Compter le nombre de clics pour ce LinkID
	totalClicks, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clicks: %w", err)
	}

	// Retourner les 3 valeurs
	return link, totalClicks, nil
}
