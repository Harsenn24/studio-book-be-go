package address

import (
	"net/http"
	"studio-book-be-go/internal/helper"

	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	addressService AddressService
}

func NewAddressHandler(addressService AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

func (h *AddressHandler) ProvincePagination(c *gin.Context) {
	var input ListProvinceRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.addressService.ProvincePagination(c.Request.Context(), input)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	helper.SendSuccess(c, http.StatusOK, "StatusOk", response)
}
