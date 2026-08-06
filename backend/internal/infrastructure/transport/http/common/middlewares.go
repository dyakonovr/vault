package common

// ---------- REQUEST ID ----------

// RequestIDMiddleware генерирует уникальный идентификатор запроса,
// помещает его в контекст и в заголовок ответа X-Request-ID.
// func RequestIDMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		id := uuid.New().String()
// 		w.Header().Set("X-Request-ID", id)
// 		ctx := context.WithValue(r.Context(), ctxkeys.RequestIDKey, id)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
