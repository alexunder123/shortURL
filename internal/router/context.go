package router

import (
	"fmt"
	"net/http"
)

func ReadContextID(r *http.Request) string {
	UserID := r.Context()
	ID := UserID.Value(name)
	if ID == nil {
		return ""
	}
	IDs := fmt.Sprintf("%s", ID)
	return IDs
}
