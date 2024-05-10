package config

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

var AppConfig Config

type Config struct {
	S3BucketName    string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	//Loglevel        string
	//ImageFormat     string
	SaveTo string
}

func init() {
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/AiImageOperator/internal/config/")

	if err := viper.ReadInConfig(); err != nil {
		if !(errors.As(err, &viper.ConfigFileNotFoundError{})) {
			log.Fatalf("Error reading config file: %s", err)
		}
	}

	viper.SetDefault("S3BucketName", "dummyS3")
	viper.SetDefault("AccessKeyID", "")
	viper.SetDefault("SecretAccessKey", "")
	viper.SetDefault("SessionToken", "")
	//viper.SetDefault("Loglevel", "Info")
	//viper.SetDefault("ImageFormat", "png")
	viper.SetDefault("SaveTo", "redis")

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

}
