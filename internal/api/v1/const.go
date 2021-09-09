package v1

const (
	HttpContentLengthKey  = "Content-Length"
	HttpContentTypeKey    = "Content-Type"
	HttpImagePng          = "image/png"
	HttpImageJPG          = "image/jpg"
	HttpApplicationJson   = "application/json"
	HttpResponseServerKey = "Server"
	HttpTokenKey          = "Authorization"
	HttpTokenPrefix       = "Bearer "
	HttpMultipartFormData = "multipart/form-data"
	UserIdKey             = "id"
	UserPwdKey            = "pwd"
)

const (
	HttpErrorContentTypeInvalid = iota + 100
	HttpErrorContentLenEmpty
	HttpErrorContentLenInvalid
	HttpErrorReadBodyFail
	HttpErrorMarshalJsonFail
	HttpTokenEmpty
	HttpErrorGenerateTokenFail
)
