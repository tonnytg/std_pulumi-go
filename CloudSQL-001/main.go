package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	psql "github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"log"
	"os"
)

var (
	PROJECT     = os.Getenv("GOOGLE_CLOUD_PROJECT")
	BUCKET_NAME = os.Getenv("BUCKET_NAME")
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Create a GCP resource (Storage Bucket)
		bucket, err := storage.NewBucket(ctx, BUCKET_NAME, &storage.BucketArgs{
			Location: pulumi.String("US"),
		})
		if err != nil {
			return err
		}

		// Create Bucket Object with SQL script
		_, err = storage.NewBucketObject(ctx, "sql-object", &storage.BucketObjectArgs{
			Bucket: bucket.Name,
			Source: pulumi.NewFileAsset("CloudSQL-001/data.sql"),
		})

		instance, err := psql.NewDatabaseInstance(ctx, "myinstance1", &psql.DatabaseInstanceArgs{
			Region:          pulumi.String("us-central1"),
			DatabaseVersion: pulumi.String("POSTGRES_14"),
			Settings: &psql.DatabaseInstanceSettingsArgs{
				Tier: pulumi.String("db-f1-micro"),
			},
			DeletionProtection: pulumi.Bool(false),
			Name:               pulumi.String("myinstance1"),
			RootPassword:       pulumi.String("mysecretpassword"),
		})
		if err != nil {
			return err
		}

		database, err := psql.NewDatabase(ctx, "mydatabase1", &psql.DatabaseArgs{
			Instance: instance.Name,
			Name:     pulumi.String("guestbook"),
			Project:  pulumi.String(PROJECT),
		})
		if err != nil {
			return err
		}

		fmt.Println("Database Name:", database.SelfLink)
		fmt.Println("Database URN:", database.URN())

		return nil
	})

	connStr := "postgres://tonnytg:mysecretpassword@34.71.191.38"
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
