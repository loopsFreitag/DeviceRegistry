package cmd

import (
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/loopsFreitag/DeviceRegistry/internal/config"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
)

var allowMissingMigrations bool

func init() {
	migrateCmd.PersistentFlags().BoolVarP(&allowMissingMigrations, "allow-missing", "", false, `Allow missing migrations`)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateUpByOneCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "DB migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate up",
	Run: func(cmd *cobra.Command, args []string) {
		config.ReadConfig(model.Environment, "")

		// get the shared instance of dbx just to be able to close it when
		// the command exits.
		dbx := model.InitDB()
		defer func(dbx *sqlx.DB) {
			log.Println("Closing DB connection...")
			if err := dbx.Close(); err != nil {
				log.Error("Failed to close DB connection. err: ", err.Error())
			}
		}(dbx)

		opts := []goose.OptionsFunc{}
		if allowMissingMigrations {
			opts = append(opts, goose.WithAllowMissing())
		}

		if err := goose.Up(model.DBX().DB, viper.GetString("migration.dir"), opts...); err != nil {
			log.Fatalln(err)
		}

		log.Println("Migrations completed successfully")
	},
}

var migrateUpByOneCmd = &cobra.Command{
	Use:   "up-by-one",
	Short: "Migrate up by one version",
	Run: func(cmd *cobra.Command, args []string) {
		config.ReadConfig(model.Environment, "")

		// get the shared instance of dbx just to be able to close it when
		// the command exits.
		dbx := model.InitDB()
		defer func(dbx *sqlx.DB) {
			log.Println("Closing DB connection...")
			if err := dbx.Close(); err != nil {
				log.Error("Failed to close DB connection. err: ", err.Error())
			}
		}(dbx)

		opts := []goose.OptionsFunc{}
		if allowMissingMigrations {
			opts = append(opts, goose.WithAllowMissing())
		}

		if err := goose.UpByOne(model.DBX().DB, viper.GetString("migration.dir"), opts...); err != nil {
			log.Fatalln(err)
		}

		log.Println("Migration completed successfully")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Migrate down",
	Run: func(cmd *cobra.Command, args []string) {
		config.ReadConfig(model.Environment, "")

		// get the shared instance of dbx just to be able to close it when
		// the command exits.
		dbx := model.InitDB()
		defer func(dbx *sqlx.DB) {
			log.Println("Closing DB connection...")
			if err := dbx.Close(); err != nil {
				log.Error("Failed to close DB connection. err: ", err.Error())
			}
		}(dbx)

		opts := []goose.OptionsFunc{}
		if allowMissingMigrations {
			opts = append(opts, goose.WithAllowMissing())
		}

		if err := goose.Down(model.DBX().DB, viper.GetString("migration.dir"), opts...); err != nil {
			log.Fatalln(err)
		}

		log.Println("Rollback completed successfully")
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Migration status",
	Run: func(cmd *cobra.Command, args []string) {
		config.ReadConfig(model.Environment, "")

		// get the shared instance of dbx just to be able to close it when
		// the command exits.
		dbx := model.InitDB()
		defer func(dbx *sqlx.DB) {
			log.Println("Closing DB connection...")
			if err := dbx.Close(); err != nil {
				log.Error("Failed to close DB connection. err: ", err.Error())
			}
		}(dbx)

		opts := []goose.OptionsFunc{}
		if err := goose.Status(model.DBX().DB, viper.GetString("migration.dir"), opts...); err != nil {
			log.Fatalln(err)
		}
	},
}
