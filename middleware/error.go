package middleware

type MyError struct {
	Code uint32 `json:"code"`
	Message string `json:"message"`
	Data any `json:"data,omitempty"`
}
