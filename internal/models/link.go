package models

import "time"

// Link représente un lien raccourci dans la base de données.
// Les tags `gorm:"..."` définissent comment GORM doit mapper cette structure à une table SQL.
type Link struct {
	ID        uint      `gorm:"primaryKey"`                   // ID est la clé primaire
	ShortCode string    `gorm:"size:10;uniqueIndex;not null"` // ShortCode doit être unique, indexé, taille max 10 caractères
	LongURL   string    `gorm:"type:text;not null"`           // LongURL ne doit pas être null
	CreatedAt time.Time // Horodatage de la création du lien

	// Relation avec les clics : un lien peut avoir plusieurs clics
	Clicks []Click `gorm:"foreignKey:LinkID"`
}
