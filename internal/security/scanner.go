package security

import (
	"time"

	"github.com/delvop-dev/delvop/internal/provider"
)

type Alert struct {
	Severity  Severity
	Category  string
	RuleID    string
	Match     string
	Message   string
	Timestamp time.Time
}

type Scanner struct {
	disabledRules map[string]bool
	allowedHosts  []string
}

func New() *Scanner {
	return &Scanner{
		disabledRules: make(map[string]bool),
	}
}

func (s *Scanner) DisableRule(id string) {
	s.disabledRules[id] = true
}

func (s *Scanner) SetAllowedHosts(hosts []string) {
	s.allowedHosts = hosts
}

func (s *Scanner) ScanPermission(perm *provider.PermissionRequest) []Alert {
	if perm == nil {
		return nil
	}
	text := perm.Description + "\n" + perm.RawContent
	return s.scan(text, modePermission)
}

func (s *Scanner) ScanPaneContent(content string) []Alert {
	return s.scan(content, modePane)
}

func (s *Scanner) scan(text string, mode scanMode) []Alert {
	var alerts []Alert
	for _, r := range allRules {
		if s.disabledRules[r.ID] {
			continue
		}
		if r.Mode != mode && r.Mode != modeBoth {
			continue
		}
		if loc := r.Pattern.FindStringIndex(text); loc != nil {
			match := text[loc[0]:loc[1]]
			if len(match) > 120 {
				match = match[:120] + "..."
			}
			alerts = append(alerts, Alert{
				Severity:  r.Severity,
				Category:  r.Category,
				RuleID:    r.ID,
				Match:     match,
				Message:   r.Message,
				Timestamp: time.Now(),
			})
		}
	}
	return alerts
}
