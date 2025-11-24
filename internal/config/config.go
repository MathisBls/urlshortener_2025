package config

import (
	"fmt"
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)

// Config est la structure principale qui mappe l'intégralité de la configuration de l'application.
// Les tags `mapstructure` sont utilisés par Viper pour mapper les clés du fichier de config
// (ou des variables d'environnement) aux champs de la structure Go.
type Config struct {
	Server struct {
		Port    int    `mapstructure:"port"`
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"server"`

	Database struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`

	Analytics struct {
		BufferSize  int `mapstructure:"buffer_size"`
		WorkerCount int `mapstructure:"worker_count"`
	} `mapstructure:"analytics"`

	Monitor struct {
		IntervalMinutes int `mapstructure:"interval_minutes"`
	} `mapstructure:"monitor"`
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {
	// Spécifie le chemin où Viper doit chercher les fichiers de config.
	// on cherche dans le dossier 'configs' relatif au répertoire d'exécution.
	viper.AddConfigPath("configs")

	// Spécifie le nom du fichier de config (sans l'extension).
	viper.SetConfigName("config")

	// Spécifie le type de fichier de config.
	viper.SetConfigType("yaml")

	// Définir les valeurs par défaut pour toutes les options de configuration.
	// Ces valeurs seront utilisées si les clés correspondantes ne sont pas trouvées dans le fichier de config
	// ou si le fichier n'existe pas.
	// server.port, server.base_url etc.
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")

	viper.SetDefault("database.name", "url_shortener.db")

	viper.SetDefault("analytics.buffer_size", 1000)
	viper.SetDefault("analytics.worker_count", 5)

	viper.SetDefault("monitor.interval_minutes", 5)

	// Lire le fichier de configuration.
	if err := viper.ReadInConfig(); err != nil {
		// Si le fichier de config n'existe pas, on reste sur les valeurs par défaut.
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Aucun fichier de configuration trouvé, utilisation des valeurs par défaut : %v", err)
		} else {
			// Erreur réelle de lecture/parsing
			return nil, fmt.Errorf("erreur lors de la lecture du fichier de configuration: %w", err)
		}
	}

	// 4: Démapper (unmarshal) la configuration lue (ou les valeurs par défaut) dans la structure Config.
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("erreur lors du mapping de la configuration: %w", err)
	}

	// Log pour vérifier la config chargée
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
