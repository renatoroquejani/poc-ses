# Amazon SES API

Esta aplicação oferece uma API REST para interagir com o Amazon Simple Email Service (SES), permitindo cadastrar remetentes de e-mail e coletar métricas detalhadas sobre os envios.

## Funcionalidades

- Gerenciamento de remetentes (cadastro, listagem, detalhes, exclusão)
- Coleta de métricas gerais de envio de e-mails
- Coleta de métricas específicas por remetente
- Documentação completa da API via Swagger
- Envio de e-mails com suporte a anexos

## Requisitos

- Go 1.16+
- Acesso à AWS com permissões para SES e CloudWatch
- Credenciais da AWS configuradas

## Configuração

A aplicação utiliza variáveis de ambiente para configuração:

```
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=sua_access_key
AWS_SECRET_ACCESS_KEY=sua_secret_key
SERVER_PORT=8080
```

## Instalação

### Instalar dependências

```bash
go mod tidy
```

### Gerar documentação Swagger

```bash
go get -u github.com/swaggo/swag/cmd/swag
swag init -g cmd/main.go
```

### Compilar a aplicação

```bash
go build -o poc-ses ./cmd
```

## Executar

```bash
./poc-ses
```

A API estará disponível em `http://localhost:8080` e a documentação do Swagger em `http://localhost:8080/swagger/index.html`.

## Endpoints

### Gerenciamento de Remetentes

- `POST /api/v1/senders` - Cadastra um novo remetente
- `GET /api/v1/senders` - Lista todos os remetentes
- `GET /api/v1/senders/{email}` - Obtém detalhes de um remetente
- `DELETE /api/v1/senders/{email}` - Remove um remetente

### Métricas

- `GET /api/v1/metrics` - Obtém métricas gerais de envios
- `GET /api/v1/metrics/sender/{email}` - Obtém métricas de um remetente específico

### Envio de E-mails

- `POST /api/v1/emails/send` - Envia um e-mail usando um remetente verificado

## Exemplo de uso

### Cadastrar um remetente

```bash
curl -X POST http://localhost:8080/api/v1/senders \
  -H "Content-Type: application/json" \
  -d '{"email":"seu-email@exemplo.com","name":"Nome do Remetente"}'
```

### Obter métricas de um remetente

```bash
curl -X GET "http://localhost:8080/api/v1/metrics/sender/seu-email@exemplo.com?startDate=2023-01-01&endDate=2023-02-01"
```

### Enviar um e-mail

```bash
curl -X POST http://localhost:8080/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "seu-email-verificado@exemplo.com",
    "to": ["destinatario@exemplo.com"],
    "cc": ["copia@exemplo.com"],
    "subject": "Teste de Envio",
    "htmlBody": "<h1>Olá</h1><p>Este é um e-mail de teste enviado pelo Amazon SES.</p>",
    "textBody": "Olá! Este é um e-mail de teste enviado pelo Amazon SES."
  }'
```

### Enviar um e-mail com anexo

```bash
curl -X POST http://localhost:8080/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "seu-email-verificado@exemplo.com",
    "to": ["destinatario@exemplo.com"],
    "subject": "E-mail com anexo",
    "htmlBody": "<h1>Olá</h1><p>Este é um e-mail com anexo enviado pelo Amazon SES.</p>",
    "attachments": [
      {
        "filename": "documento.txt",
        "content": "SGVsbG8sIHRoaXMgaXMgYSB0ZXN0IGZpbGUgY29udGVudCE="
      }
    ]
  }'
```

**Observação**: O campo `content` do anexo deve estar codificado em Base64.

## Observações importantes

1. Ao cadastrar um novo remetente, o Amazon SES enviará um e-mail para o endereço informado para verificação.
2. O remetente só poderá ser utilizado para envios após a verificação.
3. As métricas são coletadas do CloudWatch e podem ter um atraso de até 15 minutos.
