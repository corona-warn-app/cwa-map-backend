package main

import (
	"com.t-systems-mms.cwa/api"
	"com.t-systems-mms.cwa/external/geocoding"
	"com.t-systems-mms.cwa/repositories"
	"com.t-systems-mms.cwa/services"
	"crypto/rsa"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/s12v/go-jwks"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	logrus.Info("Starting application")
	rand.Seed(time.Now().Unix())

	if err := LoadConfig(); err != nil {
		logrus.WithError(err).Fatal("Error loading config")
		os.Exit(1)
	}

	if lvl, err := logrus.ParseLevel(appConfig.Logging.Level); err == nil {
		logrus.SetLevel(lvl)
	} else {
		logrus.WithError(err).Fatal("Error setting log level")
		os.Exit(1)
	}

	geocoder, err := geocoding.NewGoogleGeocoder(appConfig.Google)
	if err != nil {
		logrus.WithError(err).Fatal("Error creating google client")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d",
		appConfig.Database.Host,
		appConfig.Database.User,
		appConfig.Database.Password,
		appConfig.Database.Database,
		appConfig.Database.Port)), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("Error creating database connection")
		os.Exit(1)
	}

	if appConfig.Logging.LogSQL {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		logrus.WithError(err).Fatal("Error getting database connection")
		os.Exit(1)
	}
	sqlDB.SetMaxIdleConns(appConfig.Database.IdlePoolSize)
	sqlDB.SetMaxOpenConns(appConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(appConfig.Database.ConnMaxLifetime) * time.Minute)

	settingsRepository := repositories.NewSystemSettingsRepository(db)

	centersRepository := repositories.NewCentersRepository(db)
	operatorsRepository := repositories.NewOperatorsRepository(db)
	operatorsService := services.NewOperatorsService(operatorsRepository)
	centersService := services.NewCentersService(centersRepository, operatorsRepository, operatorsService, geocoder)

	mailService := services.NewMailService(appConfig.Email)

	bugReportsRepository := repositories.NewBugReportsRepository(db)
	bugReportsService := services.NewBugReportsService(appConfig.BugReports,
		mailService, centersRepository, bugReportsRepository, settingsRepository)

	// configure authentication
	jwksSource := jwks.NewWebSource(appConfig.Authentication.JwksUrl)
	jwksClient := jwks.NewDefaultClient(jwksSource, time.Hour, 12*time.Hour)
	key, err := jwksClient.GetSignatureKey(appConfig.Authentication.KeyId)
	if err != nil {
		logrus.WithError(err).Fatal("Error getting jwks signature key")
		os.Exit(1)
	}

	tokenAuth := jwtauth.New(appConfig.Authentication.KeyAlg, nil, key.Certificates[0].PublicKey.(*rsa.PublicKey))

	router := chi.NewRouter()
	router.Use(middleware.DefaultLogger)
	router.Handle("/metrics", initMetricsHandler(centersRepository, operatorsRepository))
	router.Mount("/api/statistics", api.NewStatisticsAPI(bugReportsRepository, tokenAuth))
	router.Mount("/api/centers", api.NewCentersAPI(centersService, centersRepository, bugReportsService, operatorsService, geocoder, tokenAuth))
	router.Mount("/api/operators", api.NewOperatorsAPI(operatorsRepository, operatorsService, tokenAuth))

	server := &http.Server{
		Addr:    appConfig.Server.Listen,
		Handler: router,
		// TODO: read und write timeouts definieren
		// TODO: tlsConfig ergänzen
	}

	serverWaitHandle := &sync.WaitGroup{}
	go func() {
		serverWaitHandle.Add(1)
		logrus.WithFields(logrus.Fields{"listen": server.Addr}).Info("Start listening for connections")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Error listening for connections")
		}
		serverWaitHandle.Done()
	}()

	go bugReportsService.PublishScheduler()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	<-signals

	logrus.Info("Stopping application")
	if err := server.Close(); err != nil {
		logrus.WithError(err).Fatal("Error closing server")
	}

	serverWaitHandle.Wait()
	logrus.Info("Application stopped")
}
