package serverclient

type Vulnerability struct {
	UUID           string  `json:"uuid"`
	ScanID         string  `json:"scan_id"`
	ProjectID      string  `json:"project_id"`
	WorkspaceID    string  `json:"workspace_id"`
	CVEID          string  `json:"cve_id,omitempty"`
	Title          string  `json:"title"`
	Severity       string  `json:"severity"`
	CVSSScore      float64 `json:"cvss_score,omitempty"`
	PackageName    string  `json:"package_name,omitempty"`
	PackageVersion string  `json:"package_version,omitempty"`
	FixedVersion   string  `json:"fixed_version,omitempty"`
	Status         string  `json:"status"`
	DetectedAt     string  `json:"detected_at"`
	ResolvedAt     string  `json:"resolved_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type VulnerabilitiesData struct {
	Vulnerability  *Vulnerability  `json:"vulnerability,omitempty"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

func (c *Client) ListVulnerabilities(params ...map[string]string) ([]Vulnerability, error) {
	path := "/api/v1/vulnerabilities"
	if len(params) > 0 {
		first := true
		for k, v := range params[0] {
			if first {
				path += "?" + k + "=" + v
				first = false
			} else {
				path += "&" + k + "=" + v
			}
		}
	}
	var data VulnerabilitiesData
	err := c.Get(path, &data)
	return data.Vulnerabilities, err
}

func (c *Client) GetVulnerability(uuid string) (*Vulnerability, error) {
	var data VulnerabilitiesData
	err := c.Get("/api/v1/vulnerabilities/"+uuid, &data)
	return data.Vulnerability, err
}

func (c *Client) UpdateVulnerability(uuid string, payload map[string]interface{}) (*Vulnerability, error) {
	var data VulnerabilitiesData
	err := c.Patch("/api/v1/vulnerabilities/"+uuid, payload, &data)
	return data.Vulnerability, err
}

type Finding struct {
	UUID           string  `json:"uuid"`
	ScanID         string  `json:"scan_id"`
	Tool           string  `json:"tool"`
	Severity       string  `json:"severity"`
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	FilePath       string  `json:"file_path,omitempty"`
	LineStart      int     `json:"line_start,omitempty"`
	LineEnd        int     `json:"line_end,omitempty"`
	RuleID         string  `json:"rule_id,omitempty"`
	CWEID          string  `json:"cwe_id,omitempty"`
	CVEID          string  `json:"cve_id,omitempty"`
	CVSSScore      float64 `json:"cvss_score,omitempty"`
	PackageName    string  `json:"package_name,omitempty"`
	FixedVersion   string  `json:"fixed_version,omitempty"`
	Status         string  `json:"status"`
	DetectedAt     string  `json:"detected_at"`
	CreatedAt      string  `json:"created_at"`
}

type FindingsData struct {
	Finding  *Finding  `json:"finding,omitempty"`
	Findings []Finding `json:"findings,omitempty"`
}

func (c *Client) ListFindings(params ...map[string]string) ([]Finding, error) {
	path := "/api/v1/findings"
	if len(params) > 0 {
		first := true
		for k, v := range params[0] {
			if first {
				path += "?" + k + "=" + v
				first = false
			} else {
				path += "&" + k + "=" + v
			}
		}
	}
	var data FindingsData
	err := c.Get(path, &data)
	return data.Findings, err
}

func (c *Client) GetFinding(uuid string) (*Finding, error) {
	var data FindingsData
	err := c.Get("/api/v1/findings/"+uuid, &data)
	return data.Finding, err
}
