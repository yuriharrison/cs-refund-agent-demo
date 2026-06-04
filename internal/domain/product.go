package domain

type ProductType string

const (
	ProductTypeElectronics  ProductType = "electronics"
	ProductTypeClothing     ProductType = "clothing"
	ProductTypeFood         ProductType = "food"
	ProductTypeSoftware     ProductType = "software"
	ProductTypeSubscription ProductType = "subscription"
)

type Product struct {
	ID          uint        `gorm:"primarykey" json:"id"`
	Name        string      `json:"name"`
	Type        ProductType `json:"type"`
	Price       float64     `json:"price"`
	Description string      `json:"description"`
}
