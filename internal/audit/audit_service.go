package audit

import (
	"fmt"
	"organizer/internal/abstractions/entities"
	"os"
	"time"
)

type AuditService struct {
	writer *os.File
}

func New() (*AuditService, error) {
	filename := fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02T15-04-05"))

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	return &AuditService{
		writer: file,
	}, nil
}

func (a *AuditService) Log(audit entities.Audit) {
	_, err := a.writer.WriteString(fmt.Sprintf("%s | %s | %s\n", audit.Severity.String(), audit.Timestamp.String(), audit.Text))

	if err != nil {
	}

	err = a.writer.Sync()

	if err != nil {
	}
}
