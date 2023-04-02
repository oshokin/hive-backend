package city

type (
	// City represents a city entity with its ID and name.
	City struct {
		ID   int16
		Name string
	}

	// GetListRequest contains parameters for fetching a list of cities.
	GetListRequest struct {
		Search string // Search query to filter city names.
		Limit  uint64 // Maximum number of cities to return.
		Cursor int16  // ID of the last city from the previous page of results.
	}

	// GetListResponse contains a list of cities and a boolean flag indicating whether there are more cities available.
	GetListResponse struct {
		Items   []*City // List of cities returned from the query.
		HasNext bool    // True if there are more cities to fetch, false otherwise.
	}
)
