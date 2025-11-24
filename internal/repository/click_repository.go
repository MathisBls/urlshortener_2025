package repository

import (
	"fmt"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// ClickRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations sur les clics. Cette abstraction permet à la couche service
// de rester indépendante de l'implémentation spécifique de la base de données.
type ClickRepository interface {
	// CreateClick enregistre un nouvel événement de clic.
	CreateClick(click *models.Click) error
	// CountClicksByLinkID retourne le nombre de clics pour un lien donné.
	CountClicksByLinkID(linkID uint) (int, error)
}

// GormClickRepository est l'implémentation de ClickRepository utilisant GORM.
type GormClickRepository struct {
	db *gorm.DB
}

// NewClickRepository crée une nouvelle instance de GormClickRepository.
func NewClickRepository(db *gorm.DB) ClickRepository {
	if db == nil {
		panic("nil *gorm.DB passed to NewClickRepository")
	}
	return &GormClickRepository{db: db}
}

// CreateClick utilise GORM pour créer une nouvelle entrée dans la table "clicks".
func (r *GormClickRepository) CreateClick(click *models.Click) error {
	if err := r.db.Create(click).Error; err != nil {
		return fmt.Errorf("failed to create click: %w", err)
	}
	return nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
// Cette méthode est utilisée pour fournir des statistiques pour une URL courte.
func (r *GormClickRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les décomptes
	if err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count clicks for link %d: %w", linkID, err)
	}
	return int(count), nil // Conversion de int64 vers int
}
