# URL Shortener

Service de raccourcissement d'URLs haute performance développé en Go, avec analytics asynchrones, surveillance des URLs, et interfaces REST API et CLI.

## Fonctionnalités

- **Raccourcissement d'URLs** : Génération de codes courts alphanumériques uniques de 6 caractères
- **Redirection rapide** : Redirections HTTP 302 instantanées avec analytics sans latence
- **Analytics asynchrones** : Suivi des clics non-bloquant utilisant des goroutines et des channels bufferisés
- **Surveillance des URLs** : Vérifications périodiques de santé pour toutes les URLs raccourcies avec notifications de changement d'état
- **API REST** : API HTTP complète pour l'accès programmatique
- **Interface CLI** : Outils en ligne de commande pour la gestion des liens et les statistiques
- **Configurable** : Configuration basée sur YAML avec valeurs par défaut sensées

## Architecture

L'application suit une architecture en couches propre avec une séparation claire des responsabilités.

### Composants principaux

- **Models** : Entités du domaine (`Link`, `Click`, `ClickEvent`)
- **Repositories** : Couche d'accès aux données avec implémentations GORM
- **Services** : Couche de logique métier (génération de codes, validation, statistiques)
- **Workers** : Traitement asynchrone des clics utilisant des goroutines
- **Monitor** : Vérification de santé des URLs en arrière-plan avec suivi d'état
- **API Handlers** : Gestionnaires de requêtes HTTP utilisant le framework Gin
- **CLI Commands** : Interface en ligne de commande basée sur Cobra

## Structure du projet

```
urlshortener/
├── cmd/
│   ├── root.go              # Initialisation de la commande racine Cobra
│   ├── server/
│   │   └── server.go        # Démarrage et orchestration du serveur HTTP
│   └── cli/
│       ├── create.go        # Commande de création de lien court
│       ├── stats.go         # Commande d'affichage des statistiques
│       └── migrate.go       # Commande de migration de base de données
├── internal/
│   ├── api/
│   │   └── handlers.go      # Gestionnaires de requêtes HTTP (Gin)
│   ├── config/
│   │   └── config.go        # Chargement de la configuration (Viper)
│   ├── models/
│   │   ├── link.go         # Modèle de domaine Link
│   │   └── click.go        # Modèle de domaine Click
│   ├── repository/
│   │   ├── link_repository.go    # Accès aux données des liens
│   │   └── click_repository.go   # Accès aux données des clics
│   ├── services/
│   │   ├── link_service.go       # Logique métier des liens
│   │   └── click_service.go      # Logique métier des clics
│   ├── workers/
│   │   └── click_workers.go      # Traitement asynchrone des clics
│   └── monitor/
│       └── url_monitor.go        # Surveillance de santé des URLs
├── configs/
│   └── config.yaml          # Configuration de l'application
├── main.go                  # Point d'entrée de l'application
└── README.md
```

## Installation

### Prérequis

- Go 1.24+
- SQLite (inclus via le driver GORM)

### Compilation

```bash
git clone [https://github.com/axellelanca/urlshortener.git](https://github.com/MathisBls/urlshortener_2025.git)
cd urlshortener
go mod tidy
go build -o url-shortener
```

## Démarrage rapide

### 1. Initialiser la base de données

```bash
./url-shortener migrate
```

Cela crée le fichier de base de données SQLite (`url_shortener.db`) et configure les tables nécessaires.

### 2. Démarrer le serveur

```bash
./url-shortener run-server
```

Le serveur démarre sur le port 8080 (configurable) et lance :
- Le serveur API HTTP
- Les workers d'analytics de clics (5 workers par défaut)
- Le service de surveillance des URLs (vérifie toutes les 5 minutes par défaut)

### 3. Créer un lien court

**Via CLI :**
```bash
./url-shortener create --url="https://www.example.com"
```

**Via API :**
```bash
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://www.example.com"}'
```

### 4. Accéder au lien court

```bash
curl -L http://localhost:8080/VOTRE_CODE
```

La redirection se fait instantanément et les analytics de clics sont enregistrés de manière asynchrone.

### 5. Voir les statistiques

**Via CLI :**
```bash
./url-shortener stats --code="VOTRE_CODE"
```

**Via API :**
```bash
curl http://localhost:8080/api/v1/links/VOTRE_CODE/stats
```

## Configuration

Modifiez `configs/config.yaml` pour personnaliser l'application :

```yaml
server:
  port: 8080
  base_url: "http://localhost:8080"

database:
  name: "url_shortener.db"

analytics:
  buffer_size: 1000      # Taille du buffer du channel d'événements de clic
  worker_count: 5       # Nombre de workers asynchrones de clics

monitor:
  interval_minutes: 5    # Intervalle de vérification de santé des URLs
```

L'application utilise des valeurs par défaut sensées si le fichier de configuration est absent.

## Référence API

### Health Check

```http
GET /health
```

**Réponse :**
```json
{
  "status": "ok"
}
```

### Créer un lien court

```http
POST /api/v1/links
Content-Type: application/json

{
  "long_url": "https://www.example.com"
}
```

**Réponse (201 Created) :**
```json
{
  "short_code": "abc123",
  "long_url": "https://www.example.com",
  "full_short_url": "http://localhost:8080/abc123"
}
```

### Redirection

```http
GET /{shortCode}
```

**Réponse :** Redirection HTTP 302 vers l'URL originale

### Obtenir les statistiques

```http
GET /api/v1/links/{shortCode}/stats
```

**Réponse (200 OK) :**
```json
{
  "short_code": "abc123",
  "long_url": "https://www.example.com",
  "total_clicks": 42
}
```

**Réponses d'erreur :**
- `404 Not Found` : Le lien n'existe pas
- `500 Internal Server Error` : Erreur serveur

## Commandes CLI

### Créer un lien

```bash
./url-shortener create --url="https://www.example.com"
```

### Voir les statistiques

```bash
./url-shortener stats --code="abc123"
```

### Lancer le serveur

```bash
./url-shortener run-server
```

### Migration de base de données

```bash
./url-shortener migrate
```

## Détails techniques

### Génération de codes courts

- **Algorithme** : Génération aléatoire cryptographiquement sécurisée utilisant `crypto/rand`
- **Longueur** : 6 caractères (configurable)
- **Jeu de caractères** : `a-z`, `A-Z`, `0-9` (62 caractères)
- **Gestion des collisions** : Nouvelle tentative automatique jusqu'à 5 fois avec vérification d'unicité
- **Probabilité** : ~56 milliards de combinaisons possibles (62^6)

### Traitement asynchrone des clics

- **Architecture** : Pattern worker pool avec channels bufferisés
- **Workers** : Taille du pool configurable (par défaut : 5 goroutines)
- **Channel** : Channel bufferisé (par défaut : 1000 événements)
- **Non-bloquant** : Les redirections n'attendent jamais la persistance des clics
- **Résilience** : Protection contre le débordement du channel avec abandon d'événements

### Surveillance des URLs

- **Méthode** : Requêtes HTTP HEAD avec timeout de 5 secondes
- **Intervalle** : Configurable (par défaut : 5 minutes)
- **Suivi d'état** : Map d'état en mémoire avec protection par mutex
- **Notifications** : Logs des changements d'état (ACCESSIBLE ↔ INACCESSIBLE)
- **Codes de statut** : Les codes 2xx et 3xx sont considérés comme accessibles

### Schéma de base de données

**Table Links :**
- `id` (uint, clé primaire)
- `short_code` (string, unique, indexé, max 10 caractères)
- `long_url` (text, not null)
- `created_at` (timestamp)

**Table Clicks :**
- `id` (uint, clé primaire)
- `link_id` (uint, clé étrangère, indexé)
- `timestamp` (timestamp)
- `user_agent` (string, max 255)
- `ip_address` (string, max 50)

## Patterns de conception

- **Repository Pattern** : Couche d'abstraction pour l'accès aux données
- **Service Layer** : Séparation de la logique métier
- **Dependency Injection** : Gestion des dépendances basée sur les constructeurs
- **Worker Pool** : Traitement concurrent de tâches
- **Observer Pattern** : Notifications de changement d'état des URLs

## Concurrence et sécurité

- **Goroutines** : Utilisées pour les workers et la surveillance
- **Channels** : Channels bufferisés pour la communication asynchrone
- **Mutexes** : Protection d'état dans le moniteur d'URLs
- **Thread Safety** : Tous les états partagés sont correctement synchronisés

## Gestion des erreurs

- **Validation** : Validation des entrées aux frontières API et CLI
- **Erreurs de base de données** : Propagation appropriée des erreurs avec contexte
- **Dégradation gracieuse** : Abandon des événements de clic si le channel est plein (loggé)
- **Codes de statut HTTP** : Codes de statut appropriés pour tous les scénarios

## Considérations de performance

- **Redirections non-bloquantes** : Les analytics ne ralentissent jamais les redirections
- **Requêtes efficaces** : Recherches en base de données indexées
- **Connection Pooling** : Gestion des connexions par GORM
- **Surcharge minimale** : Requêtes HTTP HEAD légères pour la surveillance

## Développement

### Exécuter les tests

```bash
go test ./...
```

### Structure du code

Le codebase suit les meilleures pratiques Go :
- Limites de packages claires
- Design basé sur les interfaces
- Gestion d'erreurs complète
- Logging détaillé
- Gestion de configuration

## Dépendances

- **Gin** : Framework web HTTP
- **GORM** : ORM pour les opérations de base de données
- **Cobra** : Framework CLI
- **Viper** : Gestion de configuration
- **SQLite** : Base de données embarquée

## Licence

Ce projet fait partie d'un exercice éducatif.

## Contribution

Ceci est un projet éducatif. Pour des questions ou des améliorations, veuillez ouvrir une issue.
