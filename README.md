## Índice

- [Estrutura do Projeto](#estrutura-do-projeto)
- [Requisitos](#requisitos)
- [Configuração e execução](#configuração-e-execução)
- [Endpoints](#endpoints)
- [Utilitários Go](#utilitários-go)
- [Middlewares](#middlewares)
- [Licença](#licença)

## Estrutura do Projeto

```
.
├── clients/             # Clientes para interação com serviços externos
├── health/              # Endpoints e lógica para verificação de saúde do sistema
│   ├── handler.go       # Manipulador de requisições de saúde
│   ├── schema.go        # Estruturas de dados para respostas de saúde
│   └── tests.go         # Testes para endpoints de saúde
├── middlewares/         # Middlewares para processamento de requisições
│   ├── cors.go          # Configuração de CORS
│   ├── logging.go       # Middleware de logging
│   └── security_headers.go # Headers de segurança
├── payments/            # Recursos relacionados a pagamentos
├── uniforms/            # Recursos relacionados a uniformes
├── utils/               # Utilitários compartilhados
│   ├── api_config_schema.go    # Esquemas de configuração da API
│   ├── api_response_schema.go  # Estruturas de resposta da API
│   ├── enviroment_handler.go   # Gerenciamento de variáveis de ambiente
│   └── logger.go               # Configuração de logging
├── go.mod               # Gerenciamento de dependências Go
├── LICENSE              # Arquivo de licença
├── main.go              # Ponto de entrada da aplicação
└── README.md            # Este arquivo
```

## Requisitos

- Go 1.24.1 ou superior
- Um ambiente de desenvolvimento para Go configurado

## Configuração e execução

#### Instalação

1. Clone o repositório:
   ```bash
   git clone https://github.com/seu-usuario/space-backend-client.git
   cd space-backend-client
   ```

2. Instale as dependências:
   ```bash
   go mod download
   ```

#### Execução

Para executar o servidor em modo de desenvolvimento:

```bash
go run main.go
```

O servidor será iniciado na porta 8080 por padrão. Você pode acessar a API em http://localhost:8080.

#### Compilação

Para compilar o projeto e gerar um executável:

```bash
go build -o space-backend
```

Para executar o binário compilado:

```bash
./space-backend
```

## Endpoints

A API implementa os seguintes endpoints:

- **GET /v1/health** - Verifica o status de saúde da API e retorna informações como versão e tempo de atividade

## Utilitários Go

Durante o desenvolvimento, você pode usar vários utilitários Go para manter o código íntegro:

#### Formatação de Código

```bash
go fmt ./...
```

O comando `go fmt` reformata automaticamente seu código para seguir o estilo padrão do Go.

#### Verificação de Erros

```bash
go vet ./...
```

O comando `go vet` examina seu código Go em busca de erros comuns e construções suspeitas.

#### Testes

```bash
go test ./...
```

Execute todos os testes do projeto.

#### Gerenciamento de Dependências

```bash
# Adicionar uma nova dependência
go get github.com/exemplo/pacote

# Atualizar todas as dependências
go get -u ./...

# Limpar dependências não utilizadas
go mod tidy
```

## Middlewares

O projeto implementa os seguintes middlewares:

- **CORS** - Gerencia cabeçalhos Cross-Origin Resource Sharing para permitir solicitações de outros domínios
- **Logging** - Registra informações sobre solicitações HTTP recebidas
- **Security Headers** - Adiciona cabeçalhos de segurança às respostas HTTP

## Licença

Este projeto está licenciado sob os termos da licença incluída no arquivo [LICENSE](LICENSE).