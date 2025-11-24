package repository

import (
	"fmt"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// LinkRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations CRUD sur les liens.
type LinkRepository interface {
	// CreateLink crée un nouveau lien dans la base.
	CreateLink(link *models.Link) error
	// GetLinkByShortCode récupère un lien à partir de son code court.
	GetLinkByShortCode(shortCode string) (*models.Link, error)
	// GetAllLinks retourne tous les liens stockés.
	GetAllLinks() ([]models.Link, error)
	// CountClicksByLinkID retourne le nombre total de clics pour un lien donné.
	CountClicksByLinkID(linkID uint) (int, error)
}

// GormLinkRepository est l'implémentation de LinkRepository utilisant GORM.
type GormLinkRepository struct {
	db *gorm.DB
}

// NewLinkRepository crée une nouvelle instance de GormLinkRepository.
func NewLinkRepository(db *gorm.DB) LinkRepository {
	if db == nil {
		panic("nil *gorm.DB passed to NewLinkRepository")
	}
	return &GormLinkRepository{db: db}
}

// CreateLink insère un nouveau lien dans la base de données.
func (r *GormLinkRepository) CreateLink(link *models.Link) error {
	if err := r.db.Create(link).Error; err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}
	return nil
}

// GetLinkByShortCode récupère un lien en fonction de son code court.
func (r *GormLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	var link models.Link
	if err := r.db.Where("short_code = ?", shortCode).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// GetAllLinks retourne tous les liens présents dans la base.
func (r *GormLinkRepository) GetAllLinks() ([]models.Link, error) {
	var links []models.Link
	if err := r.db.Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}
	return links, nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
func (r *GormLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les comptes
	if err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count clicks for link %d: %w", linkID, err)
	}
	return int(count), nil
}
