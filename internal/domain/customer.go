package domain

type Customer struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
