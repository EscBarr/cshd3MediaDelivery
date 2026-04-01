package services

import (
	"context"
	"cshdMediaDelivery/internal/config"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"
)

type Scheduler struct {
	interval time.Duration
	task     func() error
	stopChan chan struct{}
	Media    MediaService
	DB       config.DatabaseConfig
}

func NewScheduler(task func() error, service MediaService, db config.DatabaseConfig) *Scheduler {
	return &Scheduler{
		interval: time.Duration(db.TimeBackupMinutes) * time.Minute,
		task:     task,
		stopChan: make(chan struct{}),
		Media:    service,
		DB:       db,
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runWithRetry(3)
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) runWithRetry(maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		if err := s.task(); err == nil {
			return
		} else if i < maxRetries-1 {
			log.Printf("Task failed (attempt %d/%d), retrying...", i+1, maxRetries)
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Task failed after %d attempts", maxRetries)
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) makeDatabaseCopy() (string, error) {
	fmt.Println("Creating database copy at:", time.Now())

	reader, err := pgDumpReader(BuildDSN(s.DB))
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("backup_%d.sql", time.Now().Unix())

	return s.Media.Upload(
		context.Background(),
		reader,
		filename,
	)
}

func BuildDSN(c config.DatabaseConfig) string {
	// Формат: postgres://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SslModel,
	)
}

type cmdReader struct {
	io.Reader
	cmd *exec.Cmd
}

func (r *cmdReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		_ = r.cmd.Wait()
	}
	return n, err
}

func pgDumpReader(dsn string) (io.Reader, error) {
	cmd := exec.Command(
		"pg_dump",
		dsn,
		"--format=plain",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &cmdReader{
		Reader: stdout,
		cmd:    cmd,
	}, nil
}
