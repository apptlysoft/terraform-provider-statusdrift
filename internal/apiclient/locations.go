package apiclient

// Location represents a monitoring location returned by the API.
type Location struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	LocationKey string  `json:"locationKey"`
	IsGlobal    bool    `json:"isGlobal"`
	Region      *string `json:"region"`
}

// ListLocations fetches all locations available to the authenticated user.
func (c *Client) ListLocations() ([]Location, error) {
	return AutoPaginate[Location](c, "/api/v1/location")
}
