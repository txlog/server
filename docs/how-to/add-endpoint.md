# How to Add a New Endpoint

This guide explains the standard workflow for adding a new API endpoint to the Txlog Server.

## Workflow

1. **Define the Route**: Open `main.go` and add the route definition.
2. **Create the Controller**: Create a new function in `controllers/` (or `controllers/api/v1/`).
3. **Implement Logic**: Write the business logic, accessing the database if needed.
4. **Add Swagger Comments**: Document the endpoint for Swagger generation.

## Example: Adding a "Ping" Endpoint

### 1. Define Route (`main.go`)

```go
// In main.go, inside the v1Group block:
v1Group.GET("/ping", v1API.GetPing())
```

### 2. Create Controller (`controllers/api/v1/ping_controller.go`)

Create a new file `controllers/api/v1/ping_controller.go`:

```go
package v1

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// GetPing returns a simple pong response
//
// @Summary      Ping
// @Description  Returns pong
// @Tags         utility
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /v1/ping [get]
func GetPing() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "pong",
        })
    }
}
```

### 3. Update Documentation

Run the make command to regenerate Swagger docs:

```bash
make doc
```

### 4. Test

Run the server and call the endpoint:

```bash
curl http://localhost:8080/v1/ping
```
