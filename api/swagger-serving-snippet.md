# Serve OpenAPI at `/swagger`

Use the generated spec file at `api/openapi.yaml` and mount a static route.

```go
import (
  "net/http"
)

// Raw spec file endpoint.
r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "api/openapi.yaml")
})
```

Optional UI:
- Host Swagger UI separately and point it to `/swagger`.
