package schemas

type OrderResult struct {
	NumeroPedido     string  `json:"numero_pedido" bson:"numero_pedido"`
	OrcamentoId      int     `json:"orcamento_id" bson:"orcamento_id"`
	ValorOrcamento   float64 `json:"valor_orcamento" bson:"valor_orcamento"`
	EstagioDescricao string  `json:"estagio_descricao" bson:"estagio_descricao"`
	DataPrevista     string  `json:"data_prevista" bson:"data_prevista"`
	DataCriacao      string  `json:"data_criacao" bson:"data_criacao"`
	Produtos         any     `json:"produtos" bson:"produtos"`
}

type OrderResponse struct {
	Resultados   []OrderResult `json:"resultados" bson:"resultados"`
	TotalPedidos int           `json:"total_pedidos" bson:"total_pedidos"`
}
