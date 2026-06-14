package http_server

type CreateItemRequest struct {
	ItemOpt1 string `json:"item_opt1" validate:"required"`
	ItemOpt2 string `json:"item_opt2" validate:"required"`
}

type CreateItemResponse struct {
	ID       int64  `json:"id"`
	ItemOpt1 string `json:"item_opt1"`
	ItemOpt2 string `json:"item_opt2"`
}

type GetItemResponse = CreateItemResponse
