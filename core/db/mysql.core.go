package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-sql-driver/mysql"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"microservice-go/core/db/internal"
	"microservice-go/model/config"
	"microservice-go/store"
	"os"
	"time"
)

func (Entrance) SetupMysql(config *config.DBConfigEntity, tables *[]interface{}) *gorm.DB {
	logPrefix := "setup mysql"
	store.Use.Logger.Func.Info(fmt.Sprintf("%s start ->", logPrefix))

	clientOptions := mysql.Config{
		Net:    "tcp",
		Addr:   config.Address,
		DBName: config.DB,
		Loc:    time.Local,
	}

	switch config.Auth {
	case 1: // not auth
		break
	case 2: // account password
		clientOptions.User = config.Account
		clientOptions.Passwd = config.Password
		break
	case 3: // tls
		clientOptions.User = config.Account
		clientOptions.Passwd = config.Password

		if config.CaCert == "" {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s %s", logPrefix, "no CA certificate found"))
			return nil
		}

		if config.ClientCert == "" {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s %s", logPrefix, "no client certificate found"))
			return nil
		}

		if config.ClientCertKey == "" {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s %s", logPrefix, "no client certificate key found"))
			return nil
		}

		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(config.CaCert)
		if CAErr != nil {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s read %s error: %s", logPrefix, config.CaCert, CAErr.Error()))
			return nil
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(config.ClientCert, config.ClientCertKey)
		if clientCertErr != nil {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s tls.LoadX509KeyPair err: %v", logPrefix, clientCertErr))
			return nil
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		if err := mysql.RegisterTLSConfig("custom", &tlsConfig); err != nil {
			store.Use.Logger.Func.Error(fmt.Sprintf("%s tls.LoadX509KeyPair err: %v", logPrefix, err.Error()))
			return nil
		}

		clientOptions.TLSConfig = "custom"
		break
	}

	_default := logger.New(internal.NewWriter(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})
	db, _oe := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
		PrepareStmt:            config.PrepareStmt,
		Logger:                 _default,
	})
	if _oe != nil {
		panic(fmt.Errorf("gorm open mysql error: %s", _oe))
	}

	if len(*tables) != 0 {
		// 初始化表结构
		if _te := db.AutoMigrate(*tables...); _te != nil {
			panic(fmt.Errorf("gorm db batch create table error: %s", _te))
		}
	}

	d, _de := db.DB()
	if _de != nil {
		panic(fmt.Errorf("gorm open db error: %s", _de))
	}
	d.SetMaxOpenConns(config.MaxOpenConnects)
	d.SetMaxIdleConns(config.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(config.ConnMaxLifeTime))

	store.Use.Logger.Func.Info(fmt.Sprintf("%s success ->", logPrefix))

	return nil
}
