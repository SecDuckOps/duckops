package serverclient

type BulkUploadResult struct {
	Created int `json:"created"`
	Failed  int `json:"failed"`
}

type BulkScansPayload struct {
	Scans []CreateScanPayload `json:"scans"`
}

type BulkVulnerabilitiesPayload struct {
	Vulnerabilities []map[string]interface{} `json:"vulnerabilities"`
}

type BulkFindingsPayload struct {
	Findings []map[string]interface{} `json:"findings"`
}

func (c *Client) BulkUploadScans(payload BulkScansPayload) (*BulkUploadResult, error) {
	var result BulkUploadResult
	err := c.Post("/api/v1/bulk/scans", payload, &result)
	return &result, err
}

func (c *Client) BulkUploadVulnerabilities(payload BulkVulnerabilitiesPayload) (*BulkUploadResult, error) {
	var result BulkUploadResult
	err := c.Post("/api/v1/bulk/vulnerabilities", payload, &result)
	return &result, err
}

func (c *Client) BulkUploadFindings(payload BulkFindingsPayload) (*BulkUploadResult, error) {
	var result BulkUploadResult
	err := c.Post("/api/v1/bulk/findings", payload, &result)
	return &result, err
}
