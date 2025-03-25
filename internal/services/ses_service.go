package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"gopkg.in/gomail.v2"
)

// SenderRequest representa os dados para cadastro de um remetente
type SenderRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"`
}

// EmailRequest representa os dados para envio de um e-mail
type EmailRequest struct {
	From        string   `json:"from" binding:"required,email"`
	To          []string `json:"to" binding:"required,dive,email"`
	Cc          []string `json:"cc,omitempty" binding:"omitempty,dive,email"`
	Bcc         []string `json:"bcc,omitempty" binding:"omitempty,dive,email"`
	Subject     string   `json:"subject" binding:"required"`
	HtmlBody    string   `json:"htmlBody,omitempty"`
	TextBody    string   `json:"textBody,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	TemplateId  string   `json:"templateId,omitempty"`
	TemplateData map[string]interface{} `json:"templateData,omitempty"`
}

// Attachment representa um anexo de e-mail
type Attachment struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// Template representa um template de e-mail
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	HtmlPart    string    `json:"htmlPart"`
	TextPart    string    `json:"textPart"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TemplateRequest representa uma solicitação para criar um template
type TemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	HtmlPart    string `json:"htmlPart,omitempty"`
	TextPart    string `json:"textPart,omitempty"`
}

// EmailResponse representa a resposta após o envio de um e-mail
type EmailResponse struct {
	MessageID  string    `json:"messageId"`
	From       string    `json:"from"`
	To         []string  `json:"to"`
	Subject    string    `json:"subject"`
	SentAt     time.Time `json:"sentAt"`
	StatusCode int       `json:"statusCode"`
	Status     string    `json:"status"`
}

// SenderResponse representa os dados de resposta de um remetente
type SenderResponse struct {
	Email            string    `json:"email"`
	Name             string    `json:"name,omitempty"`
	VerificationStatus string  `json:"verificationStatus"`
	RegisteredAt     time.Time `json:"registeredAt"`
}

// MetricsResponse representa as métricas gerais de envio de e-mails
type MetricsResponse struct {
	Period          string  `json:"period"`
	TotalSent       int64   `json:"totalSent"`
	TotalDelivered  int64   `json:"totalDelivered"`
	TotalOpened     int64   `json:"totalOpened"`
	TotalClicked    int64   `json:"totalClicked"`
	TotalBounced    int64   `json:"totalBounced"`
	TotalComplaints int64   `json:"totalComplaints"`
	DeliveryRate    float64 `json:"deliveryRate"`
	OpenRate        float64 `json:"openRate"`
	ClickRate       float64 `json:"clickRate"`
	BounceRate      float64 `json:"bounceRate"`
	ComplaintRate   float64 `json:"complaintRate"`
}

// SenderMetricsResponse representa as métricas de envio para um remetente específico
type SenderMetricsResponse struct {
	Email           string            `json:"email"`
	Period          string            `json:"period"`
	TotalSent       int64             `json:"totalSent"`
	TotalDelivered  int64             `json:"totalDelivered"`
	TotalOpened     int64             `json:"totalOpened"`
	TotalClicked    int64             `json:"totalClicked"`
	TotalBounced    int64             `json:"totalBounced"`
	TotalComplaints int64             `json:"totalComplaints"`
	DeliveryRate    float64           `json:"deliveryRate"`
	OpenRate        float64           `json:"openRate"`
	ClickRate       float64           `json:"clickRate"`
	BounceRate      float64           `json:"bounceRate"`
	ComplaintRate   float64           `json:"complaintRate"`
	DailyStats      []DailyMetrics    `json:"dailyStats,omitempty"`
}

// DailyMetrics representa métricas diárias
type DailyMetrics struct {
	Date            string  `json:"date"`
	Sent            int64   `json:"sent"`
	Delivered       int64   `json:"delivered"`
	Opened          int64   `json:"opened"`
	Clicked         int64   `json:"clicked"`
	Bounced         int64   `json:"bounced"`
	Complaints      int64   `json:"complaints"`
	DeliveryRate    float64 `json:"deliveryRate"`
	OpenRate        float64 `json:"openRate"`
	ClickRate       float64 `json:"clickRate"`
}

// SESService gerencia as operações com o Amazon SES
type SESService struct {
	sesClient        *ses.Client
	cloudWatchClient *cloudwatch.Client
}

// GetCloudWatchClient retorna o cliente CloudWatch para outros serviços
func (s *SESService) GetCloudWatchClient() *cloudwatch.Client {
	return s.cloudWatchClient
}

// NewSESService cria uma nova instância do SESService
func NewSESService() *SESService {
	// Carregar configuração da AWS
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Falha ao carregar configuração da AWS: %v", err))
	}
	
	return &SESService{
		sesClient:      ses.NewFromConfig(cfg),
		cloudWatchClient: cloudwatch.NewFromConfig(cfg),
	}
}

// RegisterSender registra um novo remetente no Amazon SES
func (s *SESService) RegisterSender(req SenderRequest) (*SenderResponse, error) {
	// Verificar identidade de e-mail no SES
	input := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(req.Email),
	}
	
	_, err := s.sesClient.VerifyEmailIdentity(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar identidade do e-mail: %w", err)
	}
	
	// Retornar resposta com status pendente
	return &SenderResponse{
		Email:            req.Email,
		Name:             req.Name,
		VerificationStatus: "PENDING",
		RegisteredAt:     time.Now(),
	}, nil
}

// ListSenders lista todos os remetentes cadastrados
func (s *SESService) ListSenders() ([]SenderResponse, error) {
	input := &ses.ListIdentitiesInput{
		IdentityType: types.IdentityTypeEmailAddress,
		MaxItems:     aws.Int32(100),
	}
	
	result, err := s.sesClient.ListIdentities(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar identidades: %w", err)
	}
	
	// Se não houver identidades, retornar slice vazio
	if len(result.Identities) == 0 {
		return []SenderResponse{}, nil
	}
	
	// Obter status de verificação para cada identidade
	vInput := &ses.GetIdentityVerificationAttributesInput{
		Identities: result.Identities,
	}
	
	vResult, err := s.sesClient.GetIdentityVerificationAttributes(context.Background(), vInput)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter atributos de verificação: %w", err)
	}
	
	// Construir resposta
	senders := make([]SenderResponse, 0, len(result.Identities))
	for _, identity := range result.Identities {
		status := "UNKNOWN"
		if attr, ok := vResult.VerificationAttributes[identity]; ok {
			status = string(attr.VerificationStatus)
		}
		
		senders = append(senders, SenderResponse{
			Email:            identity,
			VerificationStatus: status,
			RegisteredAt:     time.Now(), // Na prática, seria armazenado em banco de dados
		})
	}
	
	return senders, nil
}

// GetSender obtém informações de um remetente específico
func (s *SESService) GetSender(email string) (*SenderResponse, error) {
	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []string{email},
	}
	
	result, err := s.sesClient.GetIdentityVerificationAttributes(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter atributos de verificação: %w", err)
	}
	
	// Verificar se a identidade existe
	attr, ok := result.VerificationAttributes[email]
	if !ok {
		return nil, nil // Remetente não encontrado
	}
	
	return &SenderResponse{
		Email:            email,
		VerificationStatus: string(attr.VerificationStatus),
		RegisteredAt:     time.Now(), // Na prática, seria armazenado em banco de dados
	}, nil
}

// DeleteSender remove um remetente
func (s *SESService) DeleteSender(email string) error {
	input := &ses.DeleteIdentityInput{
		Identity: aws.String(email),
	}
	
	_, err := s.sesClient.DeleteIdentity(context.Background(), input)
	if err != nil {
		return fmt.Errorf("falha ao remover identidade: %w", err)
	}
	
	return nil
}

// GetMetrics obtém métricas gerais de envio de e-mails
func (s *SESService) GetMetrics(startDateStr, endDateStr string) (*MetricsResponse, error) {
	// Definir período de consulta
	startDate, endDate, err := parseDateRange(startDateStr, endDateStr)
	if err != nil {
		return nil, err
	}
	
	// Configurar período para CloudWatch
	period := int32(86400) // 1 dia em segundos
	
	// Métricas a serem coletadas
	metrics := map[string]string{
		"Send":       "AWS/SES",
		"Delivery":   "AWS/SES",
		"Open":       "AWS/SES",
		"Click":      "AWS/SES",
		"Bounce":     "AWS/SES",
		"Complaint":  "AWS/SES",
	}
	
	// Coletar métricas do CloudWatch
	metricData := make(map[string]int64)
	for metricName, namespace := range metrics {
		input := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String(namespace),
			MetricName: aws.String(metricName),
			StartTime:  &startDate,
			EndTime:    &endDate,
			Period:     &period,
			Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
		}
		
		result, err := s.cloudWatchClient.GetMetricStatistics(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("falha ao obter métrica %s: %w", metricName, err)
		}
		
		// Somar valores
		var sum int64
		for _, datapoint := range result.Datapoints {
			sum += int64(*datapoint.Sum)
		}
		
		metricData[metricName] = sum
	}
	
	// Calcular taxas
	sent := metricData["Send"]
	delivered := metricData["Delivery"]
	opened := metricData["Open"]
	clicked := metricData["Click"]
	bounced := metricData["Bounce"]
	complaints := metricData["Complaint"]
	
	var deliveryRate, openRate, clickRate, bounceRate, complaintRate float64
	
	if sent > 0 {
		deliveryRate = float64(delivered) / float64(sent) * 100
		bounceRate = float64(bounced) / float64(sent) * 100
		complaintRate = float64(complaints) / float64(sent) * 100
	}
	
	if delivered > 0 {
		openRate = float64(opened) / float64(delivered) * 100
		clickRate = float64(clicked) / float64(delivered) * 100
	}
	
	// Formatar período
	periodStr := fmt.Sprintf("%s a %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	return &MetricsResponse{
		Period:          periodStr,
		TotalSent:       sent,
		TotalDelivered:  delivered,
		TotalOpened:     opened,
		TotalClicked:    clicked,
		TotalBounced:    bounced,
		TotalComplaints: complaints,
		DeliveryRate:    deliveryRate,
		OpenRate:        openRate,
		ClickRate:       clickRate,
		BounceRate:      bounceRate,
		ComplaintRate:   complaintRate,
	}, nil
}

// GetSenderMetrics obtém métricas de envio para um remetente específico
func (s *SESService) GetSenderMetrics(email, startDateStr, endDateStr string) (*SenderMetricsResponse, error) {
	// Verificar se o remetente existe
	sender, err := s.GetSender(email)
	if err != nil {
		return nil, err
	}
	
	if sender == nil {
		return nil, nil // Remetente não encontrado
	}
	
	// Definir período de consulta
	startDate, endDate, err := parseDateRange(startDateStr, endDateStr)
	if err != nil {
		return nil, err
	}
	
	// Configurar período para CloudWatch
	period := int32(86400) // 1 dia em segundos
	
	// Dimensão para filtrar pelo remetente específico
	dimensions := []cwtypes.Dimension{
		{
			Name:  aws.String("Source"),
			Value: aws.String(email),
		},
	}
	
	// Métricas a serem coletadas
	metrics := map[string]string{
		"Send":       "AWS/SES",
		"Delivery":   "AWS/SES",
		"Open":       "AWS/SES",
		"Click":      "AWS/SES",
		"Bounce":     "AWS/SES",
		"Complaint":  "AWS/SES",
	}
	
	// Coletar métricas do CloudWatch
	metricData := make(map[string]int64)
	dailyMetrics := make(map[string]map[string]int64)
	
	for metricName, namespace := range metrics {
		input := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String(namespace),
			MetricName: aws.String(metricName),
			StartTime:  &startDate,
			EndTime:    &endDate,
			Period:     &period,
			Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
			Dimensions: dimensions,
		}
		
		result, err := s.cloudWatchClient.GetMetricStatistics(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("falha ao obter métrica %s: %w", metricName, err)
		}
		
		// Somar valores totais
		var sum int64
		for _, datapoint := range result.Datapoints {
			sum += int64(*datapoint.Sum)
			
			// Agrupar por dia
			dateKey := datapoint.Timestamp.Format("2006-01-02")
			if _, ok := dailyMetrics[dateKey]; !ok {
				dailyMetrics[dateKey] = make(map[string]int64)
			}
			dailyMetrics[dateKey][metricName] += int64(*datapoint.Sum)
		}
		
		metricData[metricName] = sum
	}
	
	// Calcular taxas globais
	sent := metricData["Send"]
	delivered := metricData["Delivery"]
	opened := metricData["Open"]
	clicked := metricData["Click"]
	bounced := metricData["Bounce"]
	complaints := metricData["Complaint"]
	
	var deliveryRate, openRate, clickRate, bounceRate, complaintRate float64
	
	if sent > 0 {
		deliveryRate = float64(delivered) / float64(sent) * 100
		bounceRate = float64(bounced) / float64(sent) * 100
		complaintRate = float64(complaints) / float64(sent) * 100
	}
	
	if delivered > 0 {
		openRate = float64(opened) / float64(delivered) * 100
		clickRate = float64(clicked) / float64(delivered) * 100
	}
	
	// Preparar dados diários
	daily := make([]DailyMetrics, 0, len(dailyMetrics))
	for date, metrics := range dailyMetrics {
		dailySent := metrics["Send"]
		dailyDelivered := metrics["Delivery"]
		dailyOpened := metrics["Open"]
		dailyClicked := metrics["Click"]
		dailyBounced := metrics["Bounce"]
		dailyComplaints := metrics["Complaint"]
		
		var dailyDeliveryRate, dailyOpenRate, dailyClickRate float64
		
		if dailySent > 0 {
			dailyDeliveryRate = float64(dailyDelivered) / float64(dailySent) * 100
		}
		
		if dailyDelivered > 0 {
			dailyOpenRate = float64(dailyOpened) / float64(dailyDelivered) * 100
			dailyClickRate = float64(dailyClicked) / float64(dailyDelivered) * 100
		}
		
		daily = append(daily, DailyMetrics{
			Date:         date,
			Sent:         dailySent,
			Delivered:    dailyDelivered,
			Opened:       dailyOpened,
			Clicked:      dailyClicked,
			Bounced:      dailyBounced,
			Complaints:   dailyComplaints,
			DeliveryRate: dailyDeliveryRate,
			OpenRate:     dailyOpenRate,
			ClickRate:    dailyClickRate,
		})
	}
	
	// Formatar período
	periodStr := fmt.Sprintf("%s a %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	return &SenderMetricsResponse{
		Email:           email,
		Period:          periodStr,
		TotalSent:       sent,
		TotalDelivered:  delivered,
		TotalOpened:     opened,
		TotalClicked:    clicked,
		TotalBounced:    bounced,
		TotalComplaints: complaints,
		DeliveryRate:    deliveryRate,
		OpenRate:        openRate,
		ClickRate:       clickRate,
		BounceRate:      bounceRate,
		ComplaintRate:   complaintRate,
		DailyStats:      daily,
	}, nil
}

// parseDateRange analisa e valida os parâmetros de data
func parseDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error
	
	// Se não informada, usar últimos 30 dias
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("formato de data inicial inválido (use YYYY-MM-DD): %w", err)
		}
	}
	
	// Se não informada, usar data atual
	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("formato de data final inválido (use YYYY-MM-DD): %w", err)
		}
		
		// Ajustar para fim do dia
		endDate = endDate.Add(24*time.Hour - time.Second)
	}
	
	// Validar período
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("data final não pode ser anterior à data inicial")
	}
	
	return startDate, endDate, nil
}

// CreateTemplate cria um template de e-mail
func (s *SESService) CreateTemplate(req TemplateRequest) (*Template, error) {
	// Verificar se pelo menos um tipo de corpo foi fornecido
	if req.HtmlPart == "" && req.TextPart == "" {
		return nil, fmt.Errorf("pelo menos um tipo de corpo (HTML ou texto) deve ser fornecido")
	}

	// ID do template: nome em minúsculas, sem espaços e com timestamp
	templateID := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	templateID = templateID + "-" + fmt.Sprintf("%d", time.Now().Unix())

	// Criar input para SES
	input := &ses.CreateTemplateInput{
		Template: &types.Template{
			TemplateName: aws.String(templateID),
			SubjectPart:  aws.String(req.Subject),
			HtmlPart:     aws.String(req.HtmlPart),
			TextPart:     aws.String(req.TextPart),
		},
	}

	// Criar template no SES
	_, err := s.sesClient.CreateTemplate(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar template: %w", err)
	}

	// Retornar o template criado
	return &Template{
		ID:        templateID,
		Name:      req.Name,
		Subject:   req.Subject,
		HtmlPart:  req.HtmlPart,
		TextPart:  req.TextPart,
		CreatedAt: time.Now(),
	}, nil
}

// ListTemplates lista todos os templates disponíveis
func (s *SESService) ListTemplates() ([]Template, error) {
	input := &ses.ListTemplatesInput{
		MaxItems: aws.Int32(100),
	}

	result, err := s.sesClient.ListTemplates(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar templates: %w", err)
	}

	templates := make([]Template, 0, len(result.TemplatesMetadata))
	for _, metadata := range result.TemplatesMetadata {
		// Obter detalhes de cada template
		getInput := &ses.GetTemplateInput{
			TemplateName: metadata.Name,
		}

		templateDetail, err := s.sesClient.GetTemplate(context.Background(), getInput)
		if err != nil {
			continue // Ignorar este template e continuar
		}

		// Adicionar à lista
		templates = append(templates, Template{
			ID:        *templateDetail.Template.TemplateName,
			Name:      *templateDetail.Template.TemplateName,
			Subject:   *templateDetail.Template.SubjectPart,
			HtmlPart:  *templateDetail.Template.HtmlPart,
			TextPart:  *templateDetail.Template.TextPart,
			CreatedAt: *metadata.CreatedTimestamp,
		})
	}

	return templates, nil
}

// GetTemplate obtém um template específico pelo ID
func (s *SESService) GetTemplate(id string) (*Template, error) {
	input := &ses.GetTemplateInput{
		TemplateName: aws.String(id),
	}

	result, err := s.sesClient.GetTemplate(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter template: %w", err)
	}

	return &Template{
		ID:        *result.Template.TemplateName,
		Name:      *result.Template.TemplateName,
		Subject:   *result.Template.SubjectPart,
		HtmlPart:  *result.Template.HtmlPart,
		TextPart:  *result.Template.TextPart,
		CreatedAt: time.Now(), // Não é possível obter este valor aqui
	}, nil
}

// DeleteTemplate remove um template
func (s *SESService) DeleteTemplate(id string) error {
	input := &ses.DeleteTemplateInput{
		TemplateName: aws.String(id),
	}

	_, err := s.sesClient.DeleteTemplate(context.Background(), input)
	if err != nil {
		return fmt.Errorf("falha ao remover template: %w", err)
	}

	return nil
}

// SendEmail envia um e-mail utilizando o Amazon SES
func (s *SESService) SendEmail(req EmailRequest) (*EmailResponse, error) {
	// Verificar se o remetente existe e está verificado
	sender, err := s.GetSender(req.From)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar remetente: %w", err)
	}

	if sender == nil {
		return nil, fmt.Errorf("remetente não encontrado")
	}

	if sender.VerificationStatus != "Success" {
		return nil, fmt.Errorf("remetente não verificado. Status atual: %s", sender.VerificationStatus)
	}

	// Se estiver usando um template
	if req.TemplateId != "" {
		return s.sendEmailWithTemplate(req)
	}

	// Verificar se pelo menos um corpo (HTML ou texto) foi fornecido
	if req.HtmlBody == "" && req.TextBody == "" {
		return nil, fmt.Errorf("pelo menos um tipo de corpo (HTML ou texto) deve ser fornecido")
	}

	// Preparar conteúdo do e-mail
	var messageBodyHTML, messageBodyText *ses.Body
	emailContent := &ses.EmailContent{
		Simple: &ses.Message{},
	}

	// Corpo HTML
	if req.HtmlBody != "" {
		messageBodyHTML = &ses.Body{
			Html: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(req.HtmlBody),
			},
		}
	}

	// Corpo texto
	if req.TextBody != "" {
		messageBodyText = &ses.Body{
			Text: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(req.TextBody),
			},
		}
	}

	// Assunto
	subject := &ses.Content{
		Charset: aws.String("UTF-8"),
		Data:    aws.String(req.Subject),
	}

	// Montar corpo da mensagem
	emailContent.Simple.Body = &ses.Body{}
	if messageBodyHTML != nil {
		emailContent.Simple.Body.Html = messageBodyHTML.Html
	}
	if messageBodyText != nil {
		emailContent.Simple.Body.Text = messageBodyText.Text
	}
	emailContent.Simple.Subject = subject

	// Criar anexos se houver
	if len(req.Attachments) > 0 {
		// Convertemos para mensagem MIME com anexos
		rawMessage, err := s.createRawEmailWithAttachments(req)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar e-mail com anexos: %w", err)
		}
		
		// Enviar e-mail raw
		sendRawEmailInput := &ses.SendRawEmailInput{
			RawMessage: &types.RawMessage{
				Data: rawMessage,
			},
		}
		
		result, err := s.sesClient.SendRawEmail(context.Background(), sendRawEmailInput)
		if err != nil {
			return nil, fmt.Errorf("falha ao enviar e-mail com anexos: %w", err)
		}
		
		return &EmailResponse{
			MessageID:  *result.MessageId,
			From:       req.From,
			To:         req.To,
			Subject:    req.Subject,
			SentAt:     time.Now(),
			StatusCode: 200,
			Status:     "success",
		}, nil
	}

	// Preparar destinatários
	destination := &ses.Destination{
		ToAddresses: req.To,
	}
	
	if len(req.Cc) > 0 {
		destination.CcAddresses = req.Cc
	}
	
	if len(req.Bcc) > 0 {
		destination.BccAddresses = req.Bcc
	}

	// Criar input para envio
	input := &ses.SendEmailInput{
		FromEmailAddress: aws.String(req.From),
		Destination:      destination,
		Content:          emailContent,
	}

	// Enviar e-mail
	result, err := s.sesClient.SendEmail(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao enviar e-mail: %w", err)
	}

	// Retornar resposta de sucesso
	return &EmailResponse{
		MessageID:  *result.MessageId,
		From:       req.From,
		To:         req.To,
		Subject:    req.Subject,
		SentAt:     time.Now(),
		StatusCode: 200,
		Status:     "success",
	}, nil
}

// createRawEmailWithAttachments cria uma mensagem de e-mail raw com anexos
func (s *SESService) createRawEmailWithAttachments(req EmailRequest) ([]byte, error) {
	// Criar uma nova mensagem de e-mail
	m := gomail.NewMessage()

	// Definir cabeçalhos básicos
	m.SetHeader("From", req.From)
	m.SetHeader("To", req.To...)
	
	if len(req.Cc) > 0 {
		m.SetHeader("Cc", req.Cc...)
	}
	
	if len(req.Bcc) > 0 {
		m.SetHeader("Bcc", req.Bcc...)
	}
	
	m.SetHeader("Subject", req.Subject)
	
	// Definir corpo de e-mail
	if req.HtmlBody != "" {
		if req.TextBody != "" {
			m.SetBody("text/plain", req.TextBody)
			m.AddAlternative("text/html", req.HtmlBody)
		} else {
			m.SetBody("text/html", req.HtmlBody)
		}
	} else {
		m.SetBody("text/plain", req.TextBody)
	}

	// Adicionar anexos
	for _, attachment := range req.Attachments {
		// Decodificar conteúdo base64
		data, err := base64.StdEncoding.DecodeString(attachment.Content)
		if err != nil {
			return nil, fmt.Errorf("falha ao decodificar anexo %s: %w", attachment.Filename, err)
		}
		
		// Anexar o arquivo decodificado
		m.AttachReader(attachment.Filename, bytes.NewReader(data))
	}

	// Criar um buffer para armazenar o e-mail
	var emailBuffer bytes.Buffer
	
	// Escrever a mensagem no buffer usando gomail
	_, err := m.WriteTo(&emailBuffer)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar mensagem raw: %w", err)
	}

	return emailBuffer.Bytes(), nil
}

// sendEmailWithTemplate envia um e-mail utilizando um template do SES
func (s *SESService) sendEmailWithTemplate(req EmailRequest) (*EmailResponse, error) {
	// Verificar se o template existe
	_, err := s.GetTemplate(req.TemplateId)
	if err != nil {
		return nil, fmt.Errorf("template não encontrado: %w", err)
	}

	// Converter dados do template para JSON
	templateData := "{}"
	if len(req.TemplateData) > 0 {
		dataBytes, err := json.Marshal(req.TemplateData)
		if err != nil {
			return nil, fmt.Errorf("falha ao serializar dados do template: %w", err)
		}
		templateData = string(dataBytes)
	}

	// Preparar destinatários
	destination := &ses.Destination{
		ToAddresses: req.To,
	}
	
	if len(req.Cc) > 0 {
		destination.CcAddresses = req.Cc
	}
	
	if len(req.Bcc) > 0 {
		destination.BccAddresses = req.Bcc
	}

	// Criar input para envio com template
	input := &ses.SendTemplatedEmailInput{
		Source:       aws.String(req.From),
		Destination:  destination,
		Template:     aws.String(req.TemplateId),
		TemplateData: aws.String(templateData),
	}

	// Enviar e-mail
	result, err := s.sesClient.SendTemplatedEmail(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("falha ao enviar e-mail com template: %w", err)
	}

	// Retornar resposta de sucesso
	return &EmailResponse{
		MessageID:  *result.MessageId,
		From:       req.From,
		To:         req.To,
		Subject:    "[Template: " + req.TemplateId + "]",
		SentAt:     time.Now(),
		StatusCode: 200,
		Status:     "success",
	}, nil
}

// CancelScheduledEmail cancela o envio de um e-mail agendado
func (s *SESService) CancelScheduledEmail(messageId string) error {
	input := &ses.CancelScheduledSendingInput{
		MessageId: aws.String(messageId),
	}

	_, err := s.sesClient.CancelScheduledSending(context.Background(), input)
	if err != nil {
		return fmt.Errorf("falha ao cancelar envio de e-mail: %w", err)
	}

	return nil
}
