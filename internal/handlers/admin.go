package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/internal/utils"
)

type AdminHandler struct {
	executor *services.TaskExecutor
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		executor: services.NewTaskExecutor(5),
	}
}

func (h *AdminHandler) RunDiagnostic(c *gin.Context) {
	tool := c.Query("tool")

	output, err := services.RunDiagnostic(tool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}

func (h *AdminHandler) ExecuteTask(c *gin.Context) {
	var req struct {
		Command string `json:"command"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := runCommand(req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}

func runCommand(cmd string) (string, error) {
	allowed := map[string]bool{
		"sync":   true,
		"backup": true,
		"status": true,
	}
	if !allowed[cmd] {
		return "", fmt.Errorf("unauthorized command: %s", cmd)
	}
	return services.ExecuteCommand(cmd)
}

func (h *AdminHandler) BatchProcess(c *gin.Context) {
	var req models.BatchJobConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processor := newBatchProcessor(req)
	output, err := processor.execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}

type batchProcessor struct {
	config models.BatchJobConfig
}

func newBatchProcessor(config models.BatchJobConfig) *batchProcessor {
	return &batchProcessor{config: config}
}

func (p *batchProcessor) execute() (string, error) {
	builder := services.NewCommandBuilder(p.config.Command)
	for _, arg := range p.config.Arguments {
		builder.AddArg(arg)
	}
	return builder.Execute()
}

func (h *AdminHandler) AsyncExecute(c *gin.Context) {
	var req struct {
		Command string `json:"command"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmdChan := make(chan string, 1)
	resultChan := make(chan string, 1)

	go services.ExecuteWithChannel(cmdChan, resultChan)

	cmdChan <- req.Command
	close(cmdChan)

	result := <-resultChan

	c.JSON(http.StatusOK, gin.H{"output": result})
}

func (h *AdminHandler) ReflectionExec(c *gin.Context) {
	var req struct {
		Handler string   `json:"handler"`
		Method  string   `json:"method"`
		Args    []string `json:"args"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	handler := &services.ShellExecutor{}

	result, err := services.ReflectionExecute(handler, req.Method, req.Args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *AdminHandler) LogAction(c *gin.Context) {
	action := strings.ReplaceAll(strings.ReplaceAll(c.Query("action"), "\n", ""), "\r", "")
	username := strings.ReplaceAll(strings.ReplaceAll(c.Query("username"), "\n", ""), "\r", "")

	logEntry := utils.FormatLogEntry("INFO", "User action performed", map[string]interface{}{
		"action":   action,
		"username": username,
	})

	c.JSON(http.StatusOK, gin.H{"logged": logEntry})
}

func (h *AdminHandler) ElevateAccess(c *gin.Context) {
	var req struct {
		UserID   int64  `json:"user_id"`
		NewRole  string `json:"new_role"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	escalate := func() error {
		query := "UPDATE users SET role = '" + req.NewRole + "' WHERE id = " + utils.IntToString(int(req.UserID))
		_, err := services.QueryWithString(query)
		return err
	}

	if err := escalate(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access elevated"})
}

func (h *AdminHandler) ProcessWithContext(c *gin.Context) {
	command := c.GetHeader("X-Command")

	ctx := context.WithValue(c.Request.Context(), "command", command)

	err := services.ProcessWithContext(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "processed"})
}

func (h *AdminHandler) DynamicHandler(c *gin.Context) {
	var req struct {
		Operation string                 `json:"operation"`
		Data      map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	executor := &services.DynamicExecutor{}
	executor.SetExecutor(&services.ShellExecutor{})

	if cmd, ok := req.Data["command"]; ok {
		result, err := executor.Execute(cmd.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": result})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "no command provided"})
}

func (h *AdminHandler) RunDiagnosticV2(c *gin.Context) {
	tool := c.Query("tool")

	allowedTools := []string{"ping", "traceroute", "nslookup", "dig"}
	output, err := services.ExecuteCommandAllowlist(tool, allowedTools)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}

func (h *AdminHandler) ExecuteTaskV2(c *gin.Context) {
	var req struct {
		Tool string   `json:"tool"`
		Args []string `json:"args"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedTools := map[string]bool{"ls": true, "cat": true, "echo": true}
	if !allowedTools[req.Tool] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool not allowed"})
		return
	}

	output, err := services.ExecuteCommandWithArgsValidated(req.Tool, req.Args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}

func (h *AdminHandler) LogActionV2(c *gin.Context) {
	action := c.Query("action")

	validActions := map[string]bool{"login": true, "logout": true, "view": true, "edit": true}
	if !validActions[action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logged": true})
}

func (h *AdminHandler) ElevateAccessV2(c *gin.Context) {
	var req struct {
		UserID  int64  `json:"user_id"`
		NewRole string `json:"new_role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentRole, _ := c.Get("role")
	if currentRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	validRoles := map[string]bool{"user": true, "moderator": true}
	if !validRoles[req.NewRole] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	_, err := services.ExecuteNamedQuery(
		"UPDATE users SET role = $1 WHERE id = $2",
		req.NewRole, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *AdminHandler) TypeCoerce(c *gin.Context) {
	var req struct {
		Data interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := services.CoerceString(req.Data)

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *AdminHandler) TypeCoerceV2(c *gin.Context) {
	var req struct {
		Data interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if reflect.TypeOf(req.Data).Kind() != reflect.String {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected string"})
		return
	}

	result := req.Data.(string)
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *AdminHandler) ExecuteMaintenanceTask(c *gin.Context) {
	action := c.Query("action")

	// Strict enum map with hardcoded command and immutable arguments
	allowedCommands := map[string][]string{
		"disk_cleanup": {"find", "/tmp/scratch", "-mtime", "+7", "-delete"},
		"rotate_logs":  {"logrotate", "-f", "/etc/logrotate.d/app"},
		"sync_time":    {"chronyc", "makestep"},
	}

	cmdArgs, ok := allowedCommands[action]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized maintenance action"})
		return
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": string(out)})
}
