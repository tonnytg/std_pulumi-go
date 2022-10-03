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
	PROJECT           = os.Getenv("GOOGLE_CLOUD_PROJECT")
	BUCKET_NAME       = os.Getenv("BUCKET_NAME")
	BUCKET_FILE       = "CloudSQL-001/data.sql"
	INSTANCE_NAME     = "my-instance"
	DATABASE_NAME     = "guestbook"
	DATABASE_TIER     = "db-n1-standard-1"
	DATABASE_USER     = "tonnytg"
	DATABASE_PASSWORD = os.Getenv("DATABASE_PASSWORD")
	DATABASE_HOST     = os.Getenv("DATABASE_HOST")
	DATABASE_TYPE     = "POSTGRES_14"
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
			Source: pulumi.NewFileAsset(BUCKET_FILE),
		})

		instance, err := psql.NewDatabaseInstance(ctx, INSTANCE_NAME, &psql.DatabaseInstanceArgs{
			Region:          pulumi.String("us-central1"),
			DatabaseVersion: pulumi.String(DATABASE_TYPE),
			Settings: &psql.DatabaseInstanceSettingsArgs{
				Tier: pulumi.String(DATABASE_TIER),
			},
			DeletionProtection: pulumi.Bool(false),
			Name:               pulumi.String(INSTANCE_NAME),
			RootPassword:       pulumi.String(DATABASE_PASSWORD),
		})
		if err != nil {
			return err
		}

		database, err := psql.NewDatabase(ctx, DATABASE_NAME, &psql.DatabaseArgs{
			Instance: instance.Name,
			Name:     pulumi.String(DATABASE_NAME),
			Project:  pulumi.String(PROJECT),
		})
		if err != nil {
			return err
		}

		fmt.Println("Database Name:", database.SelfLink)
		fmt.Println("Database URN:", database.URN())

		return nil
	})

	connStr := fmt.Sprintf("postgres://%s:%s@%s", DATABASE_USER, DATABASE_PASSWORD, DATABASE_HOST)
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
