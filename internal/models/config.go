package models

type SystemConfig struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Category  string `json:"category"`
	Encrypted bool   `json:"encrypted"`
}

type WebhookConfig struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Secret      string            `json:"secret"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers"`
	RetryCount  int               `json:"retry_count"`
	TimeoutSecs int               `json:"timeout_secs"`
}

type IntegrationConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Endpoint    string            `json:"endpoint"`
	Credentials map[string]string `json:"credentials"`
	Options     map[string]string `json:"options"`
}

type BatchJobConfig struct {
	JobID       string   `json:"job_id"`
	Command     string   `json:"command"`
	Arguments   []string `json:"arguments"`
	Schedule    string   `json:"schedule"`
	Environment []string `json:"environment"`
	WorkingDir  string   `json:"working_dir"`
}

type SearchFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Pattern  string      `json:"pattern"`
}

type QueryBuilder struct {
	Table      string
	Columns    []string
	Conditions []string
	OrderBy    string
	Limit      int
}

func (qb *QueryBuilder) AddCondition(condition string) {
	qb.Conditions = append(qb.Conditions, condition)
}

func (qb *QueryBuilder) SetOrder(order string) {
	qb.OrderBy = order
}
