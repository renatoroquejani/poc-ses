package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// DeliveryStatus representa o status de entrega atual 
type DeliveryStatus struct {
	ID                string    `json:"id"`
	FromEmail         string    `json:"fromEmail"`
	MessageID         string    `json:"messageId"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	SentAt            time.Time `json:"sentAt"`
	DeliveredAt       time.Time `json:"deliveredAt,omitempty"`
	OpenedAt          time.Time `json:"openedAt,omitempty"`
	ClickCount        int       `json:"clickCount"`
	LastClickAt       time.Time `json:"lastClickAt,omitempty"`
	Subject           string    `json:"subject"`
}

// DeliveryReport representa um relatório de entregas
type DeliveryReport struct {
	Period          string            `json:"period"`
	TotalSent       int               `json:"totalSent"`
	TotalDelivered  int               `json:"totalDelivered"`
	TotalFailed     int               `json:"totalFailed"`
	TotalOpened     int               `json:"totalOpened"`
	TotalClicked    int               `json:"totalClicked"`
	DeliveryRate    float64           `json:"deliveryRate"`
	OpenRate        float64           `json:"openRate"`
	ClickRate       float64           `json:"clickRate"`
	DetailedStatus  []DeliveryStatus  `json:"detailedStatus"`
	HourlyStats     []HourlyStats     `json:"hourlyStats"`
}

// HourlyStats representa estatísticas de entrega por hora
type HourlyStats struct {
	Hour       int     `json:"hour"`
	Sent       int     `json:"sent"`
	Delivered  int     `json:"delivered"`
	Failed     int     `json:"failed"`
	Opened     int     `json:"opened"`
	Clicked    int     `json:"clicked"`
	OpenRate   float64 `json:"openRate"`
	ClickRate  float64 `json:"clickRate"`
}

// StatusCache representa um cache de status de entregas
type StatusCache struct {
	statuses map[string]DeliveryStatus
	mutex    sync.RWMutex
}

// DeliveryService gerencia informações sobre entregas de e-mails
type DeliveryService struct {
	cloudWatchClient *cloudwatch.Client
	cache            *StatusCache
}

// NewDeliveryService cria uma nova instância do DeliveryService
func NewDeliveryService(cwClient *cloudwatch.Client) *DeliveryService {
	return &DeliveryService{
		cloudWatchClient: cwClient,
		cache: &StatusCache{
			statuses: make(map[string]DeliveryStatus),
		},
	}
}

// TrackDelivery registra um novo e-mail enviado para rastreamento
func (s *DeliveryService) TrackDelivery(email, messageId, subject string) {
	status := DeliveryStatus{
		ID:                fmt.Sprintf("%s-%d", messageId, time.Now().Unix()),
		FromEmail:         email,
		MessageID:         messageId,
		Status:            "SENT",
		StatusDescription: "E-mail enviado e aguardando processamento",
		SentAt:            time.Now(),
		Subject:           subject,
		ClickCount:        0,
	}

	s.cache.mutex.Lock()
	s.cache.statuses[messageId] = status
	s.cache.mutex.Unlock()
}

// UpdateDeliveryStatus atualiza o status de uma entrega
func (s *DeliveryService) UpdateDeliveryStatus(messageId, status, description string) error {
	s.cache.mutex.Lock()
	defer s.cache.mutex.Unlock()

	current, exists := s.cache.statuses[messageId]
	if !exists {
		return fmt.Errorf("mensagem não encontrada: %s", messageId)
	}

	current.Status = status
	current.StatusDescription = description

	if status == "DELIVERED" {
		current.DeliveredAt = time.Now()
	} else if status == "OPENED" {
		current.OpenedAt = time.Now()
	} else if status == "CLICKED" {
		current.ClickCount++
		current.LastClickAt = time.Now()
	}

	s.cache.statuses[messageId] = current
	return nil
}

// GetDeliveryStatus obtém o status atual de uma entrega
func (s *DeliveryService) GetDeliveryStatus(messageId string) (*DeliveryStatus, error) {
	s.cache.mutex.RLock()
	defer s.cache.mutex.RUnlock()

	status, exists := s.cache.statuses[messageId]
	if !exists {
		return nil, fmt.Errorf("mensagem não encontrada: %s", messageId)
	}

	return &status, nil
}

// GetAllDeliveryStatus obtém todos os status de entrega
func (s *DeliveryService) GetAllDeliveryStatus() []DeliveryStatus {
	s.cache.mutex.RLock()
	defer s.cache.mutex.RUnlock()

	result := make([]DeliveryStatus, 0, len(s.cache.statuses))
	for _, status := range s.cache.statuses {
		result = append(result, status)
	}

	// Ordenar por data de envio (mais recente primeiro)
	sort.Slice(result, func(i, j int) bool {
		return result[i].SentAt.After(result[j].SentAt)
	})

	return result
}

// GetRealTimeReport gera um relatório em tempo real das entregas recentes
func (s *DeliveryService) GetRealTimeReport(hours int) (*DeliveryReport, error) {
	// Se não for especificado, usar as últimas 24 horas
	if hours <= 0 {
		hours = 24
	}

	// Calcular período
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	// Métricas a serem coletadas
	metrics := map[string]string{
		"Send":       "AWS/SES",
		"Delivery":   "AWS/SES",
		"Open":       "AWS/SES",
		"Click":      "AWS/SES",
		"Bounce":     "AWS/SES",
		"Complaint":  "AWS/SES",
	}

	// Configurar período para CloudWatch
	period := int32(3600) // 1 hora em segundos

	// Coletar métricas do CloudWatch
	metricData := make(map[string]int)
	hourlyData := make(map[int]map[string]int)

	for metricName, namespace := range metrics {
		input := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String(namespace),
			MetricName: aws.String(metricName),
			StartTime:  &startTime,
			EndTime:    &endTime,
			Period:     &period,
			Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
		}

		result, err := s.cloudWatchClient.GetMetricStatistics(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("falha ao obter métrica %s: %w", metricName, err)
		}

		// Somar valores totais
		sum := 0
		for _, datapoint := range result.Datapoints {
			sum += int(*datapoint.Sum)

			// Agrupar por hora
			hour := datapoint.Timestamp.Hour()
			if _, ok := hourlyData[hour]; !ok {
				hourlyData[hour] = make(map[string]int)
			}
			hourlyData[hour][metricName] += int(*datapoint.Sum)
		}

		metricData[metricName] = sum
	}

	// Preparar dados por hora
	hourlyStats := make([]HourlyStats, 0, len(hourlyData))
	for hour, data := range hourlyData {
		sent := data["Send"]
		delivered := data["Delivery"]
		failed := data["Bounce"] + data["Complaint"]
		opened := data["Open"]
		clicked := data["Click"]

		var openRate, clickRate float64
		if delivered > 0 {
			openRate = float64(opened) / float64(delivered) * 100
			clickRate = float64(clicked) / float64(delivered) * 100
		}

		hourlyStats = append(hourlyStats, HourlyStats{
			Hour:      hour,
			Sent:      sent,
			Delivered: delivered,
			Failed:    failed,
			Opened:    opened,
			Clicked:   clicked,
			OpenRate:  openRate,
			ClickRate: clickRate,
		})
	}

	// Ordenar por hora
	sort.Slice(hourlyStats, func(i, j int) bool {
		return hourlyStats[i].Hour < hourlyStats[j].Hour
	})

	// Calcular totais e taxas
	sent := metricData["Send"]
	delivered := metricData["Delivery"]
	failed := metricData["Bounce"] + metricData["Complaint"]
	opened := metricData["Open"]
	clicked := metricData["Click"]

	var deliveryRate, openRate, clickRate float64
	if sent > 0 {
		deliveryRate = float64(delivered) / float64(sent) * 100
	}
	if delivered > 0 {
		openRate = float64(opened) / float64(delivered) * 100
		clickRate = float64(clicked) / float64(delivered) * 100
	}

	// Obter status detalhados de entregas recentes (máximo 100)
	statuses := s.GetAllDeliveryStatus()
	if len(statuses) > 100 {
		statuses = statuses[:100]
	}

	// Criar relatório
	report := &DeliveryReport{
		Period:         fmt.Sprintf("Últimas %d horas", hours),
		TotalSent:      sent,
		TotalDelivered: delivered,
		TotalFailed:    failed,
		TotalOpened:    opened,
		TotalClicked:   clicked,
		DeliveryRate:   deliveryRate,
		OpenRate:       openRate,
		ClickRate:      clickRate,
		DetailedStatus: statuses,
		HourlyStats:    hourlyStats,
	}

	return report, nil
}
