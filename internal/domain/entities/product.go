package entities

type RequestProducts struct {
	Products []Product `json:"products"`
}

type ResponseProducts struct {
	Products []EnrichedProduct `json:"products"`
}

type Product struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
}

type EnrichedProduct struct {
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	Availability string  `json:"availability"`
	Price        float64 `json:"price"`
	TotalCost    float64 `json:"total_cost"`
}
