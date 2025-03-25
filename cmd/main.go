package main

import (
	"log"
	"github.com/renat/poc-ses/internal/handlers"
	_ "github.com/renat/poc-ses/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           SES Email Management API
// @version         1.0
// @description     API para gerenciar remetentes de e-mail no Amazon SES e coletar métricas de envio
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	r := gin.Default()
	
	// Configurando versão da API
	v1 := r.Group("/api/v1")
	
	// Configurando handlers
	h := handlers.NewHandler()
	
	// Rotas para gerenciar remetentes
	v1.POST("/senders", h.RegisterSender)
	v1.GET("/senders", h.ListSenders)
	v1.GET("/senders/:email", h.GetSender)
	v1.DELETE("/senders/:email", h.DeleteSender)
	
	// Rotas para métricas
	v1.GET("/metrics", h.GetMetrics)
	v1.GET("/metrics/sender/:email", h.GetSenderMetrics)
	
	// Rotas para envio de e-mails
	v1.POST("/emails/send", h.SendEmail)
	v1.DELETE("/emails/cancel/:messageId", h.CancelEmail)
	
	// Rotas para templates
	v1.POST("/templates", h.CreateTemplate)
	v1.GET("/templates", h.ListTemplates)
	v1.GET("/templates/:id", h.GetTemplate)
	v1.DELETE("/templates/:id", h.DeleteTemplate)
	
	// Rotas para monitoramento de entregas
	v1.GET("/delivery/status/:messageId", h.GetDeliveryStatus)
	v1.GET("/delivery/status", h.GetAllDeliveryStatus)
	v1.GET("/delivery/report", h.GetRealTimeReport)
	
	// Rotas para envio de e-mails
	v1.POST("/emails/send", h.SendEmail)
	
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	log.Println("Iniciando servidor na porta 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}
