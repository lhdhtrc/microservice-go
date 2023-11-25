package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"microservice-go/model/config"
	"microservice-go/utils"
	"os"
	"time"
)

func (Entrance) SetupMongo(config *config.DBConfigEntity) *mongo.Database {
	logPrefix := "setup mongo"
	utils.Use.Logger.Info(fmt.Sprintf("%s %s", logPrefix, "start ->"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var clientOptions options.ClientOptions

	switch config.Auth {
	case 1: // not auth
		break
	case 2: // account password
		clientOptions.SetAuth(options.Credential{
			Username: config.Account,
			Password: config.Password,
		})
		break
	case 3: // tls
		if config.CaCert == "" {
			utils.Use.Logger.Error(fmt.Sprintf("%s %s", logPrefix, "no CA certificate found"))
			return nil
		}

		if config.ClientCert == "" {
			utils.Use.Logger.Error(fmt.Sprintf("%s %s", logPrefix, "no client certificate found"))
			return nil
		}

		if config.ClientCertKey == "" {
			utils.Use.Logger.Error(fmt.Sprintf("%s %s", logPrefix, "no client certificate key found"))
			return nil
		}

		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(config.CaCert)
		if CAErr != nil {
			utils.Use.Logger.Error(fmt.Sprintf("%s read %s error: %s", logPrefix, config.CaCert, CAErr.Error()))
			return nil
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(config.ClientCert, config.ClientCertKey)
		if clientCertErr != nil {
			utils.Use.Logger.Error(fmt.Sprintf("%s tls.LoadX509KeyPair err: %v", logPrefix, clientCertErr))
			return nil
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}
		clientOptions.SetTLSConfig(&tlsConfig)
		break
	}

	uri := fmt.Sprintf("mongodb://%s", config.Address)
	if config.Mode { // cluster
	} else {
	} // stand alone
	clientOptions.ApplyURI(uri)

	clientOptions.SetBSONOptions(&options.BSONOptions{
		UseLocalTimeZone: true,
	})

	clientOptions.SetMaxConnecting(uint64(config.MaxOpenConnects))
	clientOptions.SetMaxPoolSize(uint64(config.MaxIdleConnects))
	clientOptions.SetMaxConnIdleTime(time.Second * time.Duration(config.MaxIdleConnects))

	client, cErr := mongo.Connect(ctx, &clientOptions)
	if cErr != nil {
		utils.Use.Logger.Error(fmt.Sprintf("%s mongo client connect: %v", logPrefix, cErr))
		return nil
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		utils.Use.Logger.Error(fmt.Sprintf("%s mongo client ping: %v", logPrefix, err))
		return nil
	}

	db := client.Database(config.DB)

	utils.Use.Logger.Info(fmt.Sprintf("%s %s", logPrefix, "success ->"))

	return db
}
