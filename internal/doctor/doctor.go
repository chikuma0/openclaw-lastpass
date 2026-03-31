package doctor

import (
	"context"
	"fmt"
	"io"

	"github.com/openclaw/openclaw-lastpass/internal/config"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	Name    string
	Status  Status
	Message string
}

type Report struct {
	Checks []Check
}

type Client interface {
	LookPath() (string, error)
	Status(ctx context.Context) error
	Resolve(ctx context.Context, entry string, field config.FieldSelector) (string, error)
}

func Run(ctx context.Context, configPath string, client Client) Report {
	report := Report{}

	if path, err := client.LookPath(); err != nil {
		report.add("lpass binary", StatusFail, "not found in PATH")
	} else {
		report.add("lpass binary", StatusOK, fmt.Sprintf("found at %s", path))
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		report.add("config file", StatusFail, err.Error())
	} else {
		report.add("config file", StatusOK, fmt.Sprintf("loaded %s with %d mapping(s)", configPath, len(cfg.Mappings)))
		if len(cfg.Mappings) == 0 {
			report.add("mapping entries", StatusWarn, "config is valid but contains no secret mappings")
		}
	}

	statusErr := client.Status(ctx)
	if statusErr != nil {
		report.add("LastPass session", StatusFail, statusErr.Error())
	} else {
		report.add("LastPass session", StatusOK, "LastPass CLI access looks usable")
	}

	if err != nil || statusErr != nil {
		return report
	}

	for _, id := range cfg.SortedIDs() {
		mapping, _ := cfg.Lookup(id)
		field, parseErr := config.ParseField(mapping.Field)
		if parseErr != nil {
			report.add(id, StatusFail, parseErr.Error())
			continue
		}

		if _, resolveErr := client.Resolve(ctx, mapping.Entry, field); resolveErr != nil {
			report.add(id, StatusFail, fmt.Sprintf("entry %q field %q is not readable: %v", mapping.Entry, field.DisplayName(), resolveErr))
			continue
		}

		report.add(id, StatusOK, fmt.Sprintf("entry %q field %q is readable", mapping.Entry, field.DisplayName()))
	}

	return report
}

func (r Report) HasFailures() bool {
	for _, check := range r.Checks {
		if check.Status == StatusFail {
			return true
		}
	}
	return false
}

func (r Report) WriteTo(w io.Writer) error {
	okCount := 0
	warnCount := 0
	failCount := 0

	for _, check := range r.Checks {
		switch check.Status {
		case StatusOK:
			okCount++
		case StatusWarn:
			warnCount++
		case StatusFail:
			failCount++
		}

		if _, err := fmt.Fprintf(w, "[%s] %s: %s\n", check.Status, check.Name, check.Message); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "\nSummary: %d ok, %d warn, %d fail\n", okCount, warnCount, failCount)
	return err
}

func (r *Report) add(name string, status Status, message string) {
	r.Checks = append(r.Checks, Check{
		Name:    name,
		Status:  status,
		Message: message,
	})
}
