package api

// CheckRequest is the JSON body for POST /v1/check.
type CheckRequest struct {
	IPAddress        string   `json:"ip_address"`
	AllowedCountries []string `json:"allowed_countries"`
}

// CheckResponse is the JSON body returned on successful check.
type CheckResponse struct {
	Allowed bool   `json:"allowed"`
	Country string `json:"country"`
}

// ErrorResponse is the JSON body returned on error.
type ErrorResponse struct {
	Error string `json:"error"`
}
