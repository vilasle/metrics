package rest

import (
	"context"

	"github.com/go-chi/chi/v5"
)

func filled(v ...string) bool {
	for _, v := range v {
		if v == "" {
			return false
		}
	}
	return true
}

func notFilled(v ...string) bool {
	return !filled(v...)
}

func emptyBody() []byte {
	return []byte{}
}

func getRawDataFromContext(ctx context.Context) rawData {
	return rawData{
		Kind:  chi.URLParamFromCtx(ctx, "type"),
		Name:  chi.URLParamFromCtx(ctx, "name"),
		Value: chi.URLParamFromCtx(ctx, "value"),
	}
}
