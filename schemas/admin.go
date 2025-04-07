package schemas

type AdminUniformCreateRequest struct {
	ClientEmail string   `json:"client_email"`
	BudgetID    int      `json:"budget_id"`
	Sketches    []Sketch `json:"sketches"`
	Editable    bool     `json:"editable"`
}

type AllowUniformToEditRequest struct {
	BudgetID int `json:"budget_id"`
}
