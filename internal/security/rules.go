package security

import "regexp"

type Severity string

const (
	Critical Severity = "CRITICAL"
	Warning  Severity = "WARNING"
)

type scanMode string

const (
	modePermission scanMode = "permission"
	modePane       scanMode = "pane"
	modeBoth       scanMode = "both"
)

type Rule struct {
	ID       string
	Category string
	Severity Severity
	Mode     scanMode
	Pattern  *regexp.Regexp
	Message  string
}

var allRules = []Rule{
	// CRITICAL — Compromised agent indicators
	{ID: "EXF001", Category: "exfiltration", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(curl|wget|nc|ncat)\s.*(@|<|\$\(cat)\s*~?/`),
		Message: "Agent sending file content to external URL — possible data exfiltration"},
	{ID: "EXF002", Category: "exfiltration", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)curl\s+.*(-d\s+@|-X\s+POST\s+.*-d|--data-binary\s+@)`),
		Message: "Agent POSTing file data to external server"},
	{ID: "SEC001", Category: "secret_access", Severity: Critical, Mode: modeBoth,
		Pattern: regexp.MustCompile(`(?i)(~/\.ssh/|~/\.aws/|~/\.gnupg/|_TOKEN|_KEY|_SECRET|API_KEY|PRIVATE_KEY)`),
		Message: "Agent accessing secrets or sensitive credentials"},
	{ID: "SEC002", Category: "secret_access", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(cat|head|tail|less|more)\s+.*(\.(env|pem|key)|credentials|secrets)`),
		Message: "Agent reading sensitive configuration files"},
	{ID: "BDR001", Category: "backdoor", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(curl|wget)\s.*\|\s*(sh|bash|zsh|eval|python|node)`),
		Message: "Agent downloading and executing remote script — possible backdoor"},
	{ID: "BDR002", Category: "backdoor", Severity: Critical, Mode: modeBoth,
		Pattern: regexp.MustCompile(`(?i)eval\s*\(.*?(fetch|http|url|request)`),
		Message: "Agent evaluating code fetched from remote URL"},
	{ID: "OBF001", Category: "obfuscation", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)base64\s+(-d|--decode)\s*\|\s*(sh|bash|eval|python)`),
		Message: "Agent decoding and executing obfuscated command"},
	{ID: "OBF002", Category: "obfuscation", Severity: Critical, Mode: modePermission,
		Pattern: regexp.MustCompile(`[A-Za-z0-9+/=]{100,}`),
		Message: "Suspiciously long encoded string in command — possible obfuscation"},

	// WARNING — Reckless behavior
	{ID: "DST001", Category: "destructive", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`rm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+(/|~/|\.\s)`),
		Message: "Destructive recursive delete on broad path"},
	{ID: "DST002", Category: "destructive", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)git\s+(reset\s+--hard|push\s+.*--force|push\s+-f|clean\s+-[a-zA-Z]*f)`),
		Message: "Destructive git operation — may lose work"},
	{ID: "DST003", Category: "destructive", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(DROP\s+(TABLE|DATABASE)|DELETE\s+FROM\s+\w+\s*;)`),
		Message: "Destructive database operation"},
	{ID: "SYS001", Category: "system_mod", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(^|\s)(/etc/|~/\.(bashrc|zshrc|bash_profile|zprofile|gitconfig))`),
		Message: "Agent modifying system or shell configuration files"},
	{ID: "SYS002", Category: "system_mod", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)chmod\s+777|chown\s`),
		Message: "Agent changing file permissions or ownership broadly"},
	{ID: "SUP001", Category: "supply_chain", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(npm\s+install|pip\s+install|go\s+get|yarn\s+add|pnpm\s+add)`),
		Message: "Agent installing packages — review before approving"},
	{ID: "SUP002", Category: "supply_chain", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(npm\s+install|pip\s+install)\s+.*(git\+|https?://|\.tar|\.tgz)`),
		Message: "Agent installing from URL/tarball instead of registry"},
	{ID: "ESC001", Category: "escalation", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)(^|\s)(sudo|doas|su\s+-)\s`),
		Message: "Agent attempting privilege escalation"},
	{ID: "ESC002", Category: "escalation", Severity: Warning, Mode: modePermission,
		Pattern: regexp.MustCompile(`(?i)chmod\s+[ugo]*\+s|setuid`),
		Message: "Agent setting setuid bit — privilege escalation"},
}

func AllRules() []Rule {
	return allRules
}
