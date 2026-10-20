package services

import (
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"sync"
)

func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func ExecuteCommandWithArgs(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func RunDiagnostic(toolName string) (string, error) {
	allowed := map[string]string{
		"ping":   "ping -c 1 127.0.0.1",
		"uptime": "uptime",
		"df":     "df -h",
	}
	cmdStr, ok := allowed[toolName]
	if !ok {
		return "", fmt.Errorf("unsupported diagnostic tool: %s", toolName)
	}
	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.Output()
	return string(output), err
}

type CommandBuilder struct {
	command string
	args    []string
}

func NewCommandBuilder(cmd string) *CommandBuilder {
	return &CommandBuilder{command: cmd}
}

func (cb *CommandBuilder) AddArg(arg string) *CommandBuilder {
	cb.args = append(cb.args, arg)
	return cb
}

func (cb *CommandBuilder) Execute() (string, error) {
	fullCmd := cb.command + " " + strings.Join(cb.args, " ")
	return ExecuteCommand(fullCmd)
}

type TaskExecutor struct {
	tasks chan string
	wg    sync.WaitGroup
}

func NewTaskExecutor(workers int) *TaskExecutor {
	te := &TaskExecutor{
		tasks: make(chan string, 100),
	}

	for i := 0; i < workers; i++ {
		te.wg.Add(1)
		go te.worker()
	}

	return te
}

func (te *TaskExecutor) worker() {
	defer te.wg.Done()
	for cmd := range te.tasks {
		ExecuteCommand(cmd)
	}
}

func (te *TaskExecutor) Submit(command string) {
	te.tasks <- command
}

func (te *TaskExecutor) Stop() {
	close(te.tasks)
	te.wg.Wait()
}

func AsyncExecute(command string, resultChan chan<- string) {
	go func() {
		result, _ := ExecuteCommand(command)
		resultChan <- result
	}()
}

func ExecuteWithChannel(cmdChan <-chan string, resultChan chan<- string) {
	for cmd := range cmdChan {
		result, _ := ExecuteCommand(cmd)
		resultChan <- result
	}
}

type CommandExecutor interface {
	Run(cmd string) (string, error)
}

type ShellExecutor struct{}

func (e *ShellExecutor) Run(cmd string) (string, error) {
	return ExecuteCommand(cmd)
}

type DynamicExecutor struct {
	executor CommandExecutor
}

func (de *DynamicExecutor) SetExecutor(e CommandExecutor) {
	de.executor = e
}

func (de *DynamicExecutor) Execute(cmd string) (string, error) {
	return de.executor.Run(cmd)
}

func ReflectionExecute(handler interface{}, methodName string, args ...string) (interface{}, error) {
	v := reflect.ValueOf(handler)
	method := v.MethodByName(methodName)

	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	var callArgs []reflect.Value
	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	results := method.Call(callArgs)
	if len(results) > 0 {
		return results[0].Interface(), nil
	}
	return nil, nil
}

func ProcessWithContext(ctx context.Context, data interface{}) error {
	cmdStr, ok := ctx.Value("command").(string)
	if ok && cmdStr != "" {
		_, err := ExecuteCommand(cmdStr)
		return err
	}
	return nil
}

func ExecuteCommandAllowlist(command string, allowedCommands []string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	baseCmd := parts[0]
	allowed := false
	for _, cmd := range allowedCommands {
		if baseCmd == cmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("command not allowed: %s", baseCmd)
	}

	cmd := exec.Command(baseCmd, parts[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func ExecuteCommandWithArgsValidated(command string, args []string) (string, error) {
	for _, arg := range args {
		if strings.ContainsAny(arg, ";|&$`") {
			return "", fmt.Errorf("invalid characters in argument")
		}
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type BatchProcessor struct {
	jobs     []func() error
	maxConcurrent int
}

func NewBatchProcessor(maxConcurrent int) *BatchProcessor {
	return &BatchProcessor{
		maxConcurrent: maxConcurrent,
	}
}

func (bp *BatchProcessor) AddJob(job func() error) {
	bp.jobs = append(bp.jobs, job)
}

func (bp *BatchProcessor) Run() error {
	sem := make(chan struct{}, bp.maxConcurrent)
	errChan := make(chan error, len(bp.jobs))

	for _, job := range bp.jobs {
		sem <- struct{}{}
		go func(j func() error) {
			defer func() { <-sem }()
			if err := j(); err != nil {
				errChan <- err
			}
		}(job)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func TypeAssertProcess(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		if cmd, ok := v["command"]; ok {
			return ExecuteCommand(cmd.(string))
		}
	}
	return "", fmt.Errorf("unsupported type")
}

func CoerceString(data interface{}) string {
	return data.(string)
}
