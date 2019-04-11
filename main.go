package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/github"
	"github.com/jackc/pgx"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

// Configuration file
type TomlConfig struct {
	GitHub GitHubInfo
	Pg     PGInfo
}
type GitHubInfo struct {
	Token string
}
type PGInfo struct {
	Database       string
	NumConnections int `toml:"num_connections"`
	Port           int
	Password       string
	Server         string
	SSL            bool
	Username       string
}

var (
	// Application config
	Conf TomlConfig

	// Show debugging info?
	debug = true

	// PostgreSQL connection pool handle
	pg *pgx.ConnPool
)

func main() {
	// Override config file location via environment variables
	var err error
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		userHome, err := homedir.Dir()
		if err != nil {
			log.Fatalf("User home directory couldn't be determined: %s", "\n")
		}
		configFile = filepath.Join(userHome, ".db4s", "github_stats.toml")
	}

	// Read our configuration settings
	if _, err = toml.DecodeFile(configFile, &Conf); err != nil {
		log.Fatal(err)
	}

	// Make sure we have a GitHub access token
	if Conf.GitHub.Token == "" {
		log.Fatal("GitHub access token required.  Obtain one from https://github.com/settings/tokens and add " +
			"it to the configuration file.")
	}

	// Setup the PostgreSQL config
	pgConfig := new(pgx.ConnConfig)
	pgConfig.Host = Conf.Pg.Server
	pgConfig.Port = uint16(Conf.Pg.Port)
	pgConfig.User = Conf.Pg.Username
	pgConfig.Password = Conf.Pg.Password
	pgConfig.Database = Conf.Pg.Database
	clientTLSConfig := tls.Config{InsecureSkipVerify: true}
	if Conf.Pg.SSL {
		// TODO: Likely need to add the PG TLS cert file info here
		pgConfig.TLSConfig = &clientTLSConfig
	} else {
		pgConfig.TLSConfig = nil
	}

	// Connect to PostgreSQL
	pgPoolConfig := pgx.ConnPoolConfig{*pgConfig, 10, nil, 2 * time.Second}
	pg, err = pgx.NewConnPool(pgPoolConfig)
	if err != nil {
		log.Fatalf("Couldn't connect to PostgreSQL server: %v\n", err)
	}

	// Authenticate to the GitHub API server
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: Conf.GitHub.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Grab release info
	rels, resp, err := client.Repositories.ListReleases(ctx, "sqlitebrowser", "sqlitebrowser",
		&github.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Rate.Remaining == 0 {
		log.Fatal("Exceeded rate limit")
	}
	if len(rels) == 0 {
		log.Fatal("No releases count")
	}

	// Begin PostgreSQL transaction
	tx, err := pg.Begin()
	if err != nil {
		log.Fatal(err)
	}
	// Set up an automatic transaction roll back if the function exits without committing
	defer tx.Rollback()

	// Extract the asset download counts
	dateStamp := time.Now().UTC()
	if debug { fmt.Printf("Datestamp: %s\n", dateStamp.Format(time.RFC1123))}
	for _, rel := range rels {

		// Exclude the "continuous" release, as it's a moving point in time which has it's counter reset with every
		// commit to our main repo
		if *rel.TagName == "continuous" {
			continue
		}
		for _, asset := range rel.Assets {
			// Check if this asset is in the database
			dbQuery := `
				SELECT count(*)
				FROM github_release_assets
				WHERE asset_name = $1`
			var numResults int
			err = tx.QueryRow(dbQuery, *asset.Name).Scan(&numResults)
			if err != nil {
				log.Fatal(err)
			}
			if numResults == 0 {
				// Asset isn't yet in the database, so add it
				dbQuery := `
					INSERT INTO github_release_assets (asset_name) VALUES ($1)`
				commandTag, err := tx.Exec(dbQuery, *asset.Name)
				if err != nil {
					log.Fatal(err)
				}
				if numRows := commandTag.RowsAffected(); numRows != 1 {
					log.Fatalf("Wrong number of rows affected (%d) when adding asset '%s'\n", numRows, *asset.Name)
				}
			}

			// Store the current download count of the asset in the database
			if debug {  fmt.Printf("  * %s, downloads: %d\n", *asset.Name, *asset.DownloadCount) }
			dbQuery = `
					INSERT INTO github_download_counts (asset, count_timestamp, download_count)
					VALUES ((SELECT asset_id FROM github_release_assets WHERE asset_name = $1), $2, $3)`
			commandTag, err := tx.Exec(dbQuery, *asset.Name, dateStamp, *asset.DownloadCount)
			if err != nil {
				log.Fatal(err)
			}
			if numRows := commandTag.RowsAffected(); numRows != 1 {
				log.Fatalf("Wrong number of rows affected (%d) when adding asset '%s'\n", numRows, *asset.Name)
			}

		}
	}

	// Commit PostgreSQL transaction and close the connection
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	pg.Close()

	if debug { fmt.Println("Download counts updated") }
}
