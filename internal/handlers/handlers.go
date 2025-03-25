package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/renat/poc-ses/internal/services"
)

// Handler struct holds services for API handlers
type Handler struct {
	sesService      *services.SESService
	deliveryService *services.DeliveryService
}

// NewHandler creates a new Handler instance
func NewHandler() *Handler {
	sesService := services.NewSESService()
	return &Handler{
		sesService:      sesService,
		deliveryService: services.NewDeliveryService(sesService.GetCloudWatchClient()),
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

// SendEmail godoc
// @Summary      Envia um e-mail usando um remetente verificado
// @Description  Envia um e-mail usando um remetente previamente verificado no Amazon SES
// @Tags         emails
// @Accept       json
// @Produce      json
// @Param        email  body      services.EmailRequest  true  "Detalhes do e-mail"
// @Success      200    {object}  services.EmailResponse
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /emails/send [post]

// CreateTemplate godoc
// @Summary      Cria um novo template de e-mail
// @Description  Cria um novo template para uso no envio de e-mails
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        template  body      services.TemplateRequest  true  "Detalhes do template"
// @Success      201       {object}  services.Template
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /templates [post]

// ListTemplates godoc
// @Summary      Lista todos os templates disponíveis
// @Description  Retorna uma lista de todos os templates de e-mail cadastrados
// @Tags         templates
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.Template
// @Failure      500  {object}  map[string]string
// @Router       /templates [get]

// GetTemplate godoc
// @Summary      Obtém informações de um template específico
// @Description  Retorna detalhes de um template de e-mail específico
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do template"
// @Success      200  {object}  services.Template
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /templates/{id} [get]

// DeleteTemplate godoc
// @Summary      Remove um template
// @Description  Remove um template de e-mail
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID do template"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /templates/{id} [delete]

// CancelEmail godoc
// @Summary      Cancela um e-mail agendado
// @Description  Cancela o envio de um e-mail que ainda não foi processado
// @Tags         emails
// @Accept       json
// @Produce      json
// @Param        messageId  path      string  true  "ID da mensagem a ser cancelada"
// @Success      200        {object}  map[string]string
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /emails/cancel/{messageId} [delete]

// GetDeliveryStatus godoc
// @Summary      Obtém o status de entrega de um e-mail
// @Description  Retorna o status atual de entrega de um e-mail específico
// @Tags         delivery
// @Accept       json
// @Produce      json
// @Param        messageId  path      string  true  "ID da mensagem"
// @Success      200        {object}  services.DeliveryStatus
// @Failure      404        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /delivery/status/{messageId} [get]

// GetAllDeliveryStatus godoc
// @Summary      Lista todos os status de entrega recentes
// @Description  Retorna uma lista com o status de entrega dos e-mails enviados recentemente
// @Tags         delivery
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.DeliveryStatus
// @Failure      500  {object}  map[string]string
// @Router       /delivery/status [get]

// GetRealTimeReport godoc
// @Summary      Obtém relatório em tempo real de entregas
// @Description  Retorna um relatório detalhado das entregas de e-mail nas últimas horas
// @Tags         delivery
// @Accept       json
// @Produce      json
// @Param        hours  query     int  false  "Número de horas para análise (padrão: 24)"
// @Success      200    {object}  services.DeliveryReport
// @Failure      500    {object}  map[string]string
// @Router       /delivery/report [get]
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

// SendEmail processa a requisição de envio de e-mail
func (h *Handler) SendEmail(c *gin.Context) {
	var req services.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}
	
	// Validação adicional para envio sem template
	if req.TemplateId == "" && req.HtmlBody == "" && req.TextBody == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pelo menos um tipo de corpo (HTML ou texto) deve ser fornecido"})
		return
	}
	
	// Enviar e-mail
	result, err := h.sesService.SendEmail(req)
	if err != nil {
		status := http.StatusInternalServerError
		
		// Verificar erros específicos
		if strings.Contains(err.Error(), "remetente não encontrado") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "remetente não verificado") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "template não encontrado") {
			status = http.StatusNotFound
		}
		
		c.JSON(status, gin.H{"error": "Falha ao enviar e-mail: " + err.Error()})
		return
	}
	
	// Rastrear o status de entrega
	h.deliveryService.TrackDelivery(req.From, result.MessageID, req.Subject)
	
	c.JSON(http.StatusOK, result)
}

// CreateTemplate processa a requisição de criação de template
func (h *Handler) CreateTemplate(c *gin.Context) {
	var req services.TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}
	
	// Validação adicional
	if req.HtmlPart == "" && req.TextPart == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pelo menos um tipo de corpo (HTML ou texto) deve ser fornecido"})
		return
	}
	
	// Criar template
	result, err := h.sesService.CreateTemplate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar template: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, result)
}

// ListTemplates lista todos os templates disponíveis
func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.sesService.ListTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao listar templates: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, templates)
}

// GetTemplate obtém um template específico
func (h *Handler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	
	template, err := h.sesService.GetTemplate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter template: " + err.Error()})
		return
	}
	
	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template não encontrado"})
		return
	}
	
	c.JSON(http.StatusOK, template)
}

// DeleteTemplate remove um template
func (h *Handler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	
	err := h.sesService.DeleteTemplate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao remover template: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Template removido com sucesso"})
}

// CancelEmail cancela um e-mail agendado
func (h *Handler) CancelEmail(c *gin.Context) {
	messageId := c.Param("messageId")
	
	err := h.sesService.CancelScheduledEmail(messageId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao cancelar e-mail: " + err.Error()})
		return
	}
	
	// Atualizar o status da entrega
	h.deliveryService.UpdateDeliveryStatus(messageId, "CANCELLED", "Envio cancelado pelo usuário")
	
	c.JSON(http.StatusOK, gin.H{"message": "E-mail cancelado com sucesso"})
}

// GetDeliveryStatus obtém o status de entrega de um e-mail
func (h *Handler) GetDeliveryStatus(c *gin.Context) {
	messageId := c.Param("messageId")
	
	status, err := h.deliveryService.GetDeliveryStatus(messageId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status de entrega não encontrado: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, status)
}

// GetAllDeliveryStatus lista todos os status de entrega recentes
func (h *Handler) GetAllDeliveryStatus(c *gin.Context) {
	statuses := h.deliveryService.GetAllDeliveryStatus()
	c.JSON(http.StatusOK, statuses)
}

// GetRealTimeReport gera um relatório em tempo real das entregas recentes
func (h *Handler) GetRealTimeReport(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours := 24 // valor padrão
	
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}
	
	report, err := h.deliveryService.GetRealTimeReport(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar relatório: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, report)
}
