package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Topology Patterns
// ─────────────────────────────────────────────────────────────────────────────

// PostAdminTopologyCreatePattern creates a new topology hostname template.
// Expects form fields: template (string), display_order (int).
func PostAdminTopologyCreatePattern(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		template := c.PostForm("template")
		orderStr := c.DefaultPostForm("display_order", "0")
		order, _ := strconv.Atoi(orderStr)

		if template == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "template is required"})
			return
		}

		tm := models.NewTopologyManager(db)
		p, err := tm.CreatePattern(template, order)
		if err != nil {
			logger.Error("Failed to create topology pattern: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Topology pattern created: " + p.Template)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyUpdatePattern updates an existing topology hostname template.
// Expects form fields: id (int), template (string), display_order (int).
func PostAdminTopologyUpdatePattern(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		template := c.PostForm("template")
		orderStr := c.DefaultPostForm("display_order", "0")
		order, _ := strconv.Atoi(orderStr)

		tm := models.NewTopologyManager(db)
		if err := tm.UpdatePattern(id, template, order); err != nil {
			logger.Error("Failed to update topology pattern: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Topology pattern updated: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyDeletePattern deletes a topology hostname template.
// Expects form field: id (int).
func PostAdminTopologyDeletePattern(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		tm := models.NewTopologyManager(db)
		if err := tm.DeletePattern(id); err != nil {
			logger.Error("Failed to delete topology pattern: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Topology pattern deleted: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_deleted=1")
	}
}

// GetAdminTopologyPreview returns hostnames that match a given template.
// Used by the admin UI for live pattern preview via AJAX.
// Expects query param: template (string).
func GetAdminTopologyPreview(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		template := c.Query("template")
		if template == "" {
			c.JSON(http.StatusOK, gin.H{"hostnames": []string{}, "compiled": ""})
			return
		}

		compiled, err := models.CompileTemplate(template)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tm := models.NewTopologyManager(db)
		hostnames, err := tm.PreviewPattern(compiled)
		if err != nil {
			logger.Error("Failed to preview topology pattern: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"compiled":  compiled,
			"hostnames": hostnames,
			"count":     len(hostnames),
		})
	}
}

// GetAdminTopologyPreviewEnv returns hostnames that match an environment match_value.
func GetAdminTopologyPreviewEnv(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchValue := c.Query("match_value")
		if matchValue == "" {
			c.JSON(http.StatusOK, gin.H{"hostnames": []string{}, "count": 0})
			return
		}

		tm := models.NewTopologyManager(db)
		hostnames, err := tm.PreviewEnvironment(matchValue)
		if err != nil {
			logger.Error("Failed to preview environment match: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"hostnames": hostnames,
			"count":     len(hostnames),
		})
	}
}

// GetAdminTopologyPreviewSvc returns hostnames that match a service match_value.
func GetAdminTopologyPreviewSvc(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchValue := c.Query("match_value")
		if matchValue == "" {
			c.JSON(http.StatusOK, gin.H{"hostnames": []string{}, "count": 0})
			return
		}

		tm := models.NewTopologyManager(db)
		hostnames, err := tm.PreviewService(matchValue)
		if err != nil {
			logger.Error("Failed to preview service match: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"hostnames": hostnames,
			"count":     len(hostnames),
		})
	}
}


// ─────────────────────────────────────────────────────────────────────────────
// Environment Names
// ─────────────────────────────────────────────────────────────────────────────

// PostAdminTopologyCreateEnvironment creates a new environment name mapping.
// Expects form fields: match_value (string), name (string), display_order (int).
func PostAdminTopologyCreateEnvironment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchValue := c.PostForm("match_value")
		name := c.PostForm("name")

		if matchValue == "" || name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "match_value and name are required"})
			return
		}

		tm := models.NewTopologyManager(db)
		e, err := tm.CreateEnvironmentName(matchValue, name)
		if err != nil {
			logger.Error("Failed to create environment name: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Environment name created: " + e.MatchValue + " -> " + e.Name)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyUpdateEnvironment updates an existing environment name mapping.
// Expects form fields: id (int), match_value (string), name (string), display_order (int).
func PostAdminTopologyUpdateEnvironment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		matchValue := c.PostForm("match_value")
		name := c.PostForm("name")

		tm := models.NewTopologyManager(db)
		if err := tm.UpdateEnvironmentName(id, matchValue, name); err != nil {
			logger.Error("Failed to update environment name: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Environment name updated: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyDeleteEnvironment deletes an environment name mapping.
// Expects form field: id (int).
func PostAdminTopologyDeleteEnvironment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		tm := models.NewTopologyManager(db)
		if err := tm.DeleteEnvironmentName(id); err != nil {
			logger.Error("Failed to delete environment name: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Environment name deleted: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_deleted=1")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Service Names
// ─────────────────────────────────────────────────────────────────────────────

// PostAdminTopologyCreateService creates a new service name mapping.
// Expects form fields: match_value (string), name (string), display_order (int).
func PostAdminTopologyCreateService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchValue := c.PostForm("match_value")
		name := c.PostForm("name")
		hasPods := c.PostForm("has_pods") == "on"

		if matchValue == "" || name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "match_value and name are required"})
			return
		}

		tm := models.NewTopologyManager(db)
		s, err := tm.CreateServiceName(matchValue, name, hasPods)
		if err != nil {
			logger.Error("Failed to create service name: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Service name created: " + s.MatchValue + " -> " + s.Name)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyUpdateService updates an existing service name mapping.
// Expects form fields: id (int), match_value (string), name (string), display_order (int).
func PostAdminTopologyUpdateService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		matchValue := c.PostForm("match_value")
		name := c.PostForm("name")
		hasPods := c.PostForm("has_pods") == "on"

		tm := models.NewTopologyManager(db)
		if err := tm.UpdateServiceName(id, matchValue, name, hasPods); err != nil {
			logger.Error("Failed to update service name: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Service name updated: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_saved=1")
	}
}

// PostAdminTopologyDeleteService deletes a service name mapping.
// Expects form field: id (int).
func PostAdminTopologyDeleteService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		tm := models.NewTopologyManager(db)
		if err := tm.DeleteServiceName(id); err != nil {
			logger.Error("Failed to delete service name: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Service name deleted: id=" + idStr)
		c.Redirect(http.StatusSeeOther, "/admin?topology_deleted=1")
	}
}
