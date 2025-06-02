package handlers

type LoginRequest struct {
	Email    string `json:"email"`
	password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
	Erorr string `json:"error,omitempty"`
}

var jwtSecretKey []byte
