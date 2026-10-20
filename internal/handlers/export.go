package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/services"
)

type ExportHandler struct{}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

func (h *ExportHandler) ExportData(c *gin.Context) {
	format := c.Query("format")
	query := c.Query("query")

	var result string
	switch format {
	case "csv":
		result = h.exportCSV(query)
	case "json":
		result = h.exportJSON(query)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	c.String(http.StatusOK, result)
}

func (h *ExportHandler) exportCSV(query string) string {
	sqlQuery := "SELECT * FROM data WHERE " + query
	rows, _ := services.QueryWithString(sqlQuery)
	if rows != nil {
		defer rows.Close()
	}
	return "id,name,value\n1,test,100"
}

func (h *ExportHandler) exportJSON(query string) string {
	sqlQuery := "SELECT * FROM data WHERE " + query
	rows, _ := services.QueryWithString(sqlQuery)
	if rows != nil {
		defer rows.Close()
	}
	return `[{"id":1,"name":"test","value":100}]`
}

func (h *ExportHandler) GoroutineFlow(c *gin.Context) {
	userInput := c.Query("filter")

	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		query := "SELECT * FROM exports WHERE status = '" + userInput + "'"
		rows, _ := services.QueryWithString(query)
		if rows != nil {
			rows.Close()
		}
		resultChan <- "completed"
	}()

	wg.Wait()
	close(resultChan)

	c.JSON(http.StatusOK, gin.H{"result": <-resultChan})
}

func (h *ExportHandler) AliasedFlow(c *gin.Context) {
	input := c.Query("input")

	alias1 := input
	alias2 := alias1
	alias3 := alias2

	processValue := func(v string) string {
		return v
	}

	processed := processValue(alias3)

	query := "SELECT * FROM items WHERE name = '" + processed + "'"
	rows, _ := services.QueryWithString(query)
	if rows != nil {
		defer rows.Close()
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ExportHandler) BatchExport(c *gin.Context) {
	var req struct {
		Queries []string `json:"queries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := make(chan string, len(req.Queries))
	var wg sync.WaitGroup

	for _, q := range req.Queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			sqlQuery := "SELECT * FROM data WHERE " + query
			rows, _ := services.QueryWithString(sqlQuery)
			if rows != nil {
				rows.Close()
			}
			results <- "done"
		}(q)
	}

	wg.Wait()
	close(results)

	c.JSON(http.StatusOK, gin.H{"exported": len(req.Queries)})
}

func (h *ExportHandler) ExportDataV2(c *gin.Context) {
	format := c.Query("format")
	filterID := c.Query("id")

	allowedFormats := map[string]bool{"csv": true, "json": true}
	if !allowedFormats[format] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	rows, err := services.ExecuteNamedQuery(
		"SELECT * FROM data WHERE id = $1",
		filterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"format": format})
}

func (h *ExportHandler) GoroutineFlowV2(c *gin.Context) {
	userInput := c.Query("filter")

	allowedFilters := map[string]bool{"pending": true, "completed": true, "failed": true}
	if !allowedFilters[userInput] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, _ := services.ExecuteNamedQuery(
			"SELECT * FROM exports WHERE status = $1",
			userInput)
		if rows != nil {
			rows.Close()
		}
		resultChan <- "completed"
	}()

	wg.Wait()
	close(resultChan)

	c.JSON(http.StatusOK, gin.H{"result": <-resultChan})
}

func (h *ExportHandler) StreamExport(c *gin.Context) {
	query := c.Query("query")

	dataChan := make(chan map[string]interface{}, 100)

	go func() {
		sqlQuery := "SELECT * FROM large_data WHERE " + query
		rows, _ := services.QueryWithString(sqlQuery)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				dataChan <- map[string]interface{}{"row": "data"}
			}
		}
		close(dataChan)
	}()

	var results []map[string]interface{}
	for data := range dataChan {
		results = append(results, data)
	}

	c.JSON(http.StatusOK, results)
}

func (h *ExportHandler) StreamExportV2(c *gin.Context) {
	tableID := c.Query("table_id")

	allowedTables := map[string]string{
		"users":    "SELECT id, username, email FROM users",
		"products": "SELECT id, name, price FROM products",
	}

	query, ok := allowedTables[tableID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table"})
		return
	}

	rows, err := services.QueryWithString(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"status": "exported"})
}
