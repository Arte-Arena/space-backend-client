# Schemas

Este diretório contém as definições de tipos e estruturas de dados utilizadas em toda a aplicação.

## Organização

A estrutura é organizada por domínio e finalidade para facilitar a manutenção:

```
schemas/
  ├── common.go     # Tipos comuns usados em vários domínios
  ├── auth.go       # Tipos relacionados à autenticação
  ├── clients.go    # Tipos relacionados a clientes
  └── [domain].go   # Outros domínios específicos
```

## Convenções

Para manter o código organizado e consistente, siga estas convenções:

### Nomenclatura

- Use nomes claros e descritivos
- Siga o padrão de nomenclatura CamelCase para tipos e campos
- Use sufixos para indicar o propósito:
  - `Request` - Dados enviados pelo cliente
  - `Response` - Dados retornados para o cliente
  - `Model` ou `DB` - Dados armazenados no banco de dados
  - `Update` - Dados para atualização

### Organização dentro dos arquivos

Organize os tipos com comentários claros para separar as seções:

```go
// DOMÍNIO DATABASE MODELS
// Modelos que representam como os dados são armazenados

// DOMÍNIO API REQUESTS/RESPONSES
// Estruturas para comunicação com os clientes

// DOMÍNIO INTERNAL MODELS
// Estruturas internas usadas apenas no código
```

### Boas Práticas

1. **Reutilize tipos comuns**: Use os tipos definidos em `common.go` sempre que possível
2. **Use comentários**: Documente campos importantes, especialmente requisitos
3. **Mantenha as tags JSON/BSON consistentes**: Use o mesmo estilo de tags em todos os lugares
4. **Separe modelos de banco e API**: Evite expor campos internos nas respostas
5. **Evite duplicação**: Extraia estruturas compartilhadas em vez de duplicar campos

## Exemplos

### Definindo um novo domínio

```go
// produtos.go
package schemas

// PRODUTOS DATABASE MODELS

// ProdutoFromDB representa um produto no banco de dados
type ProdutoFromDB struct {
    ID        bson.ObjectID `bson:"_id"`
    Nome      string        `bson:"nome"`
    Preco     float64       `bson:"preco"`
    // ...
}

// PRODUTOS API REQUESTS/RESPONSES

// ProdutoCreateRequest para criar um novo produto
type ProdutoCreateRequest struct {
    Nome   string  `json:"nome"`
    Preco  float64 `json:"preco"`
    // ...
}

// ProdutoResponse para retornar dados de produto
type ProdutoResponse struct {
    ID     string  `json:"id"`
    Nome   string  `json:"nome"`
    Preco  float64 `json:"preco"`
    // ...
}
``` 