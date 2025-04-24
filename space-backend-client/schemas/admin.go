package schemas

type AdminUniformCreateRequest struct {
	ClientEmail string   `json:"client_email"`
	BudgetID    int      `json:"budget_id"`
	Sketches    []Sketch `json:"sketches"`
}

type UpdatePlayersDataRequest struct {
	BudgetID int      `json:"budget_id"`
	Players  []Player `json:"players,omitempty"`
}

type OctaChat struct {
	ID   string `json:"id"`
	Chat string `json:"chat"`
}
