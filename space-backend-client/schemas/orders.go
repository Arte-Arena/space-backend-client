package schemas

type Product struct {
	Nome       string  `json:"nome" bson:"nome"`
	Preco      float64 `json:"preco" bson:"preco"`
	Quantidade int     `json:"quantidade" bson:"quantidade"`
}

type OrderResult struct {
	NumeroPedido     string    `json:"numero_pedido" bson:"numero_pedido"`
	OrcamentoId      int       `json:"orcamento_id" bson:"orcamento_id"`
	ValorOrcamento   string    `json:"valor_orcamento" bson:"valor_orcamento"`
	EstagioDescricao string    `json:"estagio_descricao" bson:"estagio_descricao"`
	DataPrevista     string    `json:"data_prevista" bson:"data_prevista"`
	DataCriacao      string    `json:"data_criacao" bson:"data_criacao"`
	Produtos         []Product `json:"produtos" bson:"produtos"`
}

type OrderResponse struct {
	Resultados   []OrderResult `json:"resultados" bson:"resultados"`
	TotalPedidos int           `json:"total_pedidos" bson:"total_pedidos"`
}
