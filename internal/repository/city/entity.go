package city

type (
	City struct {
		ID   int16
		Name string
	}

	GetListRequest struct {
		Search string
		Limit  uint64
		Cursor int16
	}

	GetListResponse struct {
		Items   []*City
		HasNext bool
	}
)
