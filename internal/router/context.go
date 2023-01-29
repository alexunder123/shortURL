package router

import (
	"fmt"
	"net/http"
)

func ReadContextID(r *http.Request) string {
	context := r.Context()
	id := context.Value(USER_ID)
	if id == nil {
		return ""
	}
	ids := fmt.Sprintf("%s", id)
	return ids
}
