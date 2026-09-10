package ProvinceRepository

type Province struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (Province) TableName() string {
	return "provinces"
}
