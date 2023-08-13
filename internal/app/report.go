package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func NewReport(cfg *Config) *Report {
	return &Report{
		cfg:  cfg,
		data: NewReportData(cfg),
	}
}

type Report struct {
	cfg *Config

	fromDate time.Time
	toDate   time.Time

	data *ReportData
}

func (self *Report) WithFromDate(t time.Time) *Report {
	self.fromDate = t
	return self
}

func (self *Report) WithToDate(t time.Time) *Report {
	self.toDate = t
	return self
}

func (self *Report) Parse(file io.Reader) error {
	parser, err := NewParser(file)
	if err != nil {
		return err
	}

	for {
		record, err := parser.Next()
		switch {
		case err != nil:
			return err
		case !record.Valid():
			self.data.finish()
			return nil
		case !record.Out() || !record.Between(self.fromDate, self.toDate):
			continue
		}
		sectName, sectKey, err := self.cfg.FindSection(record)
		if err != nil {
			return err
		} else if sectName == "" || sectKey == "" {
			return fmt.Errorf("unknown record: line %d: %+v", record.Line(), record)
		}
		self.data.addRecord(sectName, sectKey, record)
	}
}

func (self *Report) Print() error {
	tmplPath, err := expandHomeDir(self.cfg.Template)
	if err != nil {
		return fmt.Errorf("expand home dir in %q: %w", self.cfg.Template, err)
	} else if tmplPath == "" {
		return fmt.Errorf("'template' not defined or expanded to empty string")
	}

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", tmplPath, err)
	}

	if err := tmpl.Execute(os.Stdout, self.data); err != nil {
		return fmt.Errorf("exec %q: %w", tmplPath, err)
	}

	return nil
}

func expandHomeDir(path string) (string, error) {
	tilda := "~" + string(os.PathSeparator)
	if strings.HasPrefix(path, tilda) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return os.ExpandEnv(filepath.Join(home, path[len(tilda):])), nil
	}
	return os.ExpandEnv(path), nil
}

func (self *Report) Data() *ReportData {
	return self.data
}
