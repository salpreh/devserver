package servercommons

import (
	"net/http"
	"strconv"
)

const ResponseCodeHeader string = "X-Response-Code"

func GetResponseCode(r *http.Request, defaultStatusCode int) int {
	resStatusCode := defaultStatusCode

	clientStatusCode := r.Header.Get(ResponseCodeHeader)
	r.Header.Del(ResponseCodeHeader)
	if clientStatusCode != "" {
		resStatusCode, _ = strconv.Atoi(clientStatusCode)
	}

	return resStatusCode
}
