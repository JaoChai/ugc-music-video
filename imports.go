//go:build imports
// +build imports

package main

import (
	_ "github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/hibiken/asynq"
	_ "github.com/aws/aws-sdk-go-v2"
	_ "github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/credentials"
	_ "github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/spf13/viper"
	_ "go.uber.org/zap"
	_ "github.com/google/uuid"
	_ "golang.org/x/crypto/bcrypt"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/rs/cors"
)
