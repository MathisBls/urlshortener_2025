package cli

import (
	"fmt"
	"log"
	"net/url" // Pour valider le format de l'URL
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// variable longURLFlag qui stockera la valeur du flag --url
var longURLFlag string

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1: Valider que le flag --url a été fourni.
		if longURLFlag == "" {
			fmt.Println("Erreur : le flag --url est obligatoire")
			_ = cmd.Usage()
			os.Exit(1)
		}

		// Validation basique du format de l'URL avec le package url et la fonction ParseRequestURI
		// si erreur, os.Exit(1)
		parsed, err := url.ParseRequestURI(longURLFlag)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			fmt.Printf("Erreur : URL invalide '%s'\n", longURLFlag)
			os.Exit(1)
		}

		// Charger la configuration chargée globalement via cmd.cfg
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
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		// S'assurer que la connexion est fermée à la fin de l'exécution de la commande
		defer sqlDB.Close()

		// Initialiser les repositories et services nécessaires NewLinkRepository & NewLinkService
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Appeler le LinkService et la fonction CreateLink pour créer le lien court.
		// os.Exit(1) si erreur
		link, err := linkService.CreateLink(longURLFlag)
		if err != nil {
			log.Fatalf("FATAL: échec de la création du lien court : %v", err)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.ShortCode)
		fmt.Printf("URL courte créée avec succès:\n")
		fmt.Printf("Code: %s\n", link.ShortCode)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	// Définir le flag --url pour la commande create.
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")

	// Marquer le flag comme requis
	if err := CreateCmd.MarkFlagRequired("url"); err != nil {
		log.Printf("WARN: impossible de marquer --url comme requis: %v", err)
	}

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(CreateCmd)
}
