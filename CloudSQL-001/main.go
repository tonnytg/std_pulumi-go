package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	psql "github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"log"
	"os"
)

type Config struct {
	Project string `json:"project"`
	Region  string `json:"region"`
	Bucket  struct {
		Name        string `json:"name"`
		MultiRegion string `json:"multi_regions"`
		File        string `json:"file"`
		Path        string `json:"file_path"`
	} `json:"bucket"`
	Instance struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Version      string `json:"version"`
		Tier         string `json:"tier"`
		RootPassword string `json:"root_password"`
	} `json:"instance"`
	Database struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"database"`
}

func ReadConf() Config {
	dat, err := os.ReadFile("conf.json")
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func main() {

	c := ReadConf()

	pulumi.Run(func(ctx *pulumi.Context) error {

		// Create a GCP resource (Storage Bucket)
		bucket, err := storage.NewBucket(ctx, c.Bucket.Name, &storage.BucketArgs{
			Location: pulumi.String(c.Bucket.MultiRegion),
		})
		if err != nil {
			return err
		}

		// Create Bucket Object with SQL script
		_, err = storage.NewBucketObject(ctx, c.Bucket.Name, &storage.BucketObjectArgs{
			Bucket: bucket.Name,
			Source: pulumi.NewFileAsset(c.Bucket.Path),
		})

		instance, err := psql.NewDatabaseInstance(ctx, c.Instance.Name, &psql.DatabaseInstanceArgs{
			Region:          pulumi.String(c.Region),
			DatabaseVersion: pulumi.String(c.Instance.Type),
			Settings: &psql.DatabaseInstanceSettingsArgs{
				Tier: pulumi.String(c.Instance.Tier),
			},
			DeletionProtection: pulumi.Bool(false),
			Name:               pulumi.String(c.Instance.Name),
			RootPassword:       pulumi.String(c.Instance.RootPassword),
		})
		if err != nil {
			return err
		}

		database, err := psql.NewDatabase(ctx, c.Database.Name, &psql.DatabaseArgs{
			Instance: instance.Name,
			Name:     pulumi.String(c.Database.Name),
			Project:  pulumi.String(c.Project),
		})
		if err != nil {
			return err
		}

		fmt.Println("Database Name:", database.SelfLink)
		fmt.Println("Database URN:", database.URN())

		return nil
	})

	connStr := fmt.Sprintf("postgres://%s:%s@%s", c.Database.Username, c.Database.Password, c.Database.Host)
	db, err := sql.Open(
		"postgres", connStr,
	)

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS test")
	if err != nil {
		panic(err)
	}
	fmt.Println("Database test created!")

	// Create a table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS test (id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY, valor TEXT NOT NULL)")
	if err != nil {
		panic(err)
	}

	fmt.Println(db, err)

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		log.Fatal("Build Query:", err)
	}

	for rows.Next() {
		var id int
		var valor string
		err = rows.Scan(&id, &valor)
		if err != nil {
			log.Fatal("Scan:", err)
		}
		fmt.Println(id, valor)
	}
	defer db.Close()
}
