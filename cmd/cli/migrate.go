package cli

import (
	"fmt"
	"log"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/spf13/cobra"
	"github.com/glebarez/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// MigrateCmd représente la commande 'migrate'
var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Exécute les migrations de la base de données.",
	Long:  `Cette commande initialise la connexion à la base de données et exécute les migrations GORM pour créer ou mettre à jour les tables nécessaires.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Récupérer la configuration chargée globalement via cmd.Cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: la configuration globale n'a pas été chargée")
		}

		// Initialiser la connexion à la base de données SQLite.
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: impossible de se connecter à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}
		// Assurez-vous que la connexion est fermée après la migration grâce à defer.
		defer sqlDB.Close()

		// Exécuter les migrations automatiques de GORM.
		// On passe les pointeurs vers tous les modèles.
		if err := db.AutoMigrate(&models.Link{}, &models.Click{}); err != nil {
			log.Fatalf("FATAL: échec lors de l'exécution des migrations: %v", err)
		}

		// Pas touche au log
		fmt.Println("Migrations de la base de données exécutées avec succès.")
	},
}

func init() {
	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(MigrateCmd)
}
