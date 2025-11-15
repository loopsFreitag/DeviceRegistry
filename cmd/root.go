package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/loopsFreitag/DeviceRegistry/internal/config"
	"github.com/loopsFreitag/DeviceRegistry/internal/middleware"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "deviceregistry",
	Short: "Device Registry Platform",
	Run: func(cmd *cobra.Command, args []string) {
		config.ReadConfig(model.Environment, "")

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// get the shared instance of dbx just to be able to close it when
		// the command exits.
		dbx := model.InitDB()
		defer func(dbx *sqlx.DB) {
			log.Println("Closing DB connection...")
			if err := dbx.Close(); err != nil {
				log.Error("Failed to close DB connection. err: ", err.Error())
			}
		}(dbx)

		// Create router
		router := middleware.NewAppRouter()

		// Create server
		port := viper.GetInt("port")
		if port == 0 {
			port = 8081 // default port
		}

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		}

		// Handle graceful shutdown
		go func() {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs

			log.Println("Shutting down server...")
			if err := server.Shutdown(ctx); err != nil {
				log.Error("Server shutdown error: ", err)
			}
			cancel()
		}()

		log.Printf("Starting server at port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Server error: ", err)
		}

		log.Println("Server closed")
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DeviceRegistry v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringVarP(&model.Environment, "env", "e", "development", "Environment (development/staging/production)")
	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
