package cli

import (
	"errors"
	"fmt"
	"log"
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"

	"github.com/glebarez/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// variable shortCodeFlag qui stockera la valeur du flag --code
var shortCodeFlag string

// StatsCmd représente la commande 'stats'
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Valider que le flag --code a été fourni.
		// os.Exit(1) si erreur
		if shortCodeFlag == "" {
			fmt.Println("Erreur : le flag --code est obligatoire")
			_ = cmd.Usage()
			os.Exit(1)
		}

		// Charger la configuration chargée globalement via cmd.cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("FATAL: la configuration globale n'a pas été chargée")
		}

		// 3: Initialiser la connexion à la BDD.
		// log.Fatalf si erreur
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: impossible de se connecter à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		// S'assurer que la connexion est fermée à la fin de l'exécution de la commande grâce à defer
		defer sqlDB.Close()

		// Initialiser les repositories et services nécessaires NewLinkRepository & NewLinkService
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// 5: Appeler GetLinkStats pour récupérer le lien et ses statistiques.
		// Attention, la fonction retourne 3 valeurs
		// Pour l'erreur, utilisez gorm.ErrRecordNotFound
		// Si erreur, os.Exit(1)
		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Printf("Aucun lien trouvé pour le code court : %s\n", shortCodeFlag)
				os.Exit(1)
			}
			log.Fatalf("FATAL: échec de la récupération des statistiques : %v", err)
		}

		fmt.Printf("Statistiques pour le code court: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	// Définir le flag --code pour la commande stats.
	StatsCmd.Flags().StringVar(&shortCodeFlag, "code", "", "Code court du lien à inspecter")

	// Marquer le flag comme requis
	if err := StatsCmd.MarkFlagRequired("code"); err != nil {
		log.Printf("WARN: impossible de marquer --code comme requis: %v", err)
	}

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(StatsCmd)
}
