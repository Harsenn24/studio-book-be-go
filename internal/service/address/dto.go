package address

import ProvinceRepository "studio-book-be-go/internal/repository/province"

type ListProvinceRequest struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Page   int    `json:"page"`
}

type ListProvinceResponse struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	TotalData int
	MaxPage   int
	Data      []ProvinceRepository.Province
}
