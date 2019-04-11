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
	for _, j := range rels {

		// Exclude the "continuous" release, as it's a moving point in time which has it's counter reset with every
		// commit to our main repo
		if *j.Name == "continuous" {
			continue
		}
		for _, l := range j.Assets {
			// TODO: Check if this Asset is already known

			// TODO: Store the download count for the asset in the database
			fmt.Printf("Asset: %s, downloads: %d\n", *l.Name, *l.DownloadCount)
		}
	}

	// Commit PostgreSQL transaction and close the connection
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	pg.Close()
	fmt.Println("Download counts updated")
}
