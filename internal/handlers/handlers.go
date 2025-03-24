package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/renat/poc-ses/internal/services"
)

// Handler struct holds services for API handlers
type Handler struct {
	sesService *services.SESService
}

// NewHandler creates a new Handler instance
func NewHandler() *Handler {
	return &Handler{
		sesService: services.NewSESService(),
	}
}

// RegisterSender godoc
// @Summary      Registra um novo remetente de e-mail
// @Description  Registra um novo endereço de e-mail como remetente no Amazon SES
// @Tags         senders
// @Accept       json
// @Produce      json
// @Param        sender  body      services.SenderRequest  true  "Detalhes do remetente"
// @Success      201     {object}  services.SenderResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /senders [post]
func (h *Handler) RegisterSender(c *gin.Context) {
	var req services.SenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}
	
	result, err := h.sesService.RegisterSender(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao registrar remetente: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, result)
}

// ListSenders godoc
// @Summary      Lista todos os remetentes cadastrados
// @Description  Retorna uma lista de todos os remetentes de e-mail cadastrados no Amazon SES
// @Tags         senders
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.SenderResponse
// @Failure      500  {object}  map[string]string
// @Router       /senders [get]
func (h *Handler) ListSenders(c *gin.Context) {
	senders, err := h.sesService.ListSenders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao listar remetentes: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, senders)
}

// GetSender godoc
// @Summary      Obtém informações de um remetente específico
// @Description  Retorna detalhes de um remetente de e-mail específico
// @Tags         senders
// @Accept       json
// @Produce      json
// @Param        email  path      string  true  "Endereço de e-mail do remetente"
// @Success      200    {object}  services.SenderResponse
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /senders/{email} [get]
func (h *Handler) GetSender(c *gin.Context) {
	email := c.Param("email")
	
	sender, err := h.sesService.GetSender(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter remetente: " + err.Error()})
		return
	}
	
	if sender == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Remetente não encontrado"})
		return
	}
	
	c.JSON(http.StatusOK, sender)
}

// DeleteSender godoc
// @Summary      Remove um remetente
// @Description  Remove um endereço de e-mail da lista de remetentes verificados
// @Tags         senders
// @Accept       json
// @Produce      json
// @Param        email  path      string  true  "Endereço de e-mail do remetente"
// @Success      200    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /senders/{email} [delete]
func (h *Handler) DeleteSender(c *gin.Context) {
	email := c.Param("email")
	
	err := h.sesService.DeleteSender(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao remover remetente: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Remetente removido com sucesso"})
}

// GetMetrics godoc
// @Summary      Obtém métricas gerais de envio de e-mails
// @Description  Retorna métricas gerais de todos os envios de e-mails
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Param        startDate  query     string  false  "Data inicial (formato: YYYY-MM-DD)"
// @Param        endDate    query     string  false  "Data final (formato: YYYY-MM-DD)"
// @Success      200        {object}  services.MetricsResponse
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	
	metrics, err := h.sesService.GetMetrics(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter métricas: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
}

// GetSenderMetrics godoc
// @Summary      Obtém métricas de envio para um remetente específico
// @Description  Retorna métricas detalhadas de envios de e-mails para um remetente específico
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Param        email      path      string  true   "Endereço de e-mail do remetente"
// @Param        startDate  query     string  false  "Data inicial (formato: YYYY-MM-DD)"
// @Param        endDate    query     string  false  "Data final (formato: YYYY-MM-DD)"
// @Success      200        {object}  services.SenderMetricsResponse
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /metrics/sender/{email} [get]
func (h *Handler) GetSenderMetrics(c *gin.Context) {
	email := c.Param("email")
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	
	metrics, err := h.sesService.GetSenderMetrics(email, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter métricas do remetente: " + err.Error()})
		return
	}
	
	if metrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Remetente não encontrado ou sem métricas disponíveis"})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
}
