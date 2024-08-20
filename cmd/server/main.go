package main

import (
	"GO_Assignment_3/internal/database"
	service "GO_Assignment_3/internal/service"
	transportHTTP "GO_Assignment_3/internal/transport/Htt"

	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func Run() error {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(file)
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("Setting Up Our APP")

	err2 := godotenv.Load("C:\\Personal\\GO_Assignment_3\\env_setup.env")
	if err2 != nil {
		log.Fatalf("Error loading .env file: %v", err2)
	}

	db, err := database.NewDatabase()
	if err != nil {
		log.Error("failed to setup connection to the database")
		return err
	}

	studentService := service.NewService(db)

	handler := transportHTTP.NewHandler(studentService)

	if err := handler.Serve(); err != nil {
		log.Error("failed to gracefully serve our application")
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Error(err)
		log.Fatal("Error starting up our REST API")
	}
}
