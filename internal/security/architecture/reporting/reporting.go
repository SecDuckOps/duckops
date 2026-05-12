package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
	"github.com/SecDuckOps/duckops/internal/security/architecture/c4"
	"github.com/SecDuckOps/duckops/internal/security/architecture/diagram_validation"
	"github.com/SecDuckOps/duckops/internal/security/architecture/mermaid_generator"
)

type Bundle struct {
	IRJSON           string
	Markdown         string
	SARIF            string
	HTML             string
	C4Context        string
	C4Container      string
	C4Component      string
	PlantUML         string
	Structurizr      string
	MermaidSetJSON   string
	CytoscapeJSON    string
	ValidationJSON   string
	AttackReplayJSON string
}

func Build(ir architecture.IR, diagrams c4.Models, mermaidSet mermaid_generator.Set, validations []diagram_validation.Result, cytoscapeJSON string, attackReplayJSON string, attackPaths [][]string) (Bundle, error) {
	raw, err := json.MarshalIndent(ir, "", "  ")
	if err != nil {
		return Bundle{}, err
	}
	mermaidRaw, err := json.MarshalIndent(mermaidSet, "", "  ")
	if err != nil {
		return Bundle{}, err
	}
	validationRaw, err := json.MarshalIndent(validations, "", "  ")
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{
		IRJSON:           string(raw),
		Markdown:         markdownReport(ir, diagrams, attackPaths),
		SARIF:            sarifReport(ir),
		HTML:             htmlDashboard(ir, attackPaths, string(mermaidRaw), cytoscapeJSON, string(validationRaw), attackReplayJSON),
		C4Context:        diagrams.Context,
		C4Container:      diagrams.Container,
		C4Component:      diagrams.Component,
		PlantUML:         diagrams.PlantUML,
		Structurizr:      diagrams.Structurizr,
		MermaidSetJSON:   string(mermaidRaw),
		CytoscapeJSON:    cytoscapeJSON,
		ValidationJSON:   string(validationRaw),
		AttackReplayJSON: attackReplayJSON,
	}, nil
}

func WriteAll(outDir string, b Bundle) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	files := map[string]string{
		"report.md":             b.Markdown,
		"report.json":           b.IRJSON,
		"report.sarif":          b.SARIF,
		"dashboard.html":        b.HTML,
		"c4_context.mmd":        b.C4Context,
		"c4_container.mmd":      b.C4Container,
		"c4_component.mmd":      b.C4Component,
		"c4.puml":               b.PlantUML,
		"workspace.dsl":         b.Structurizr,
		"mermaid_set.json":      b.MermaidSetJSON,
		"graph_cytoscape.json":  b.CytoscapeJSON,
		"diagram_validation.json": b.ValidationJSON,
		"attack_replay.json":    b.AttackReplayJSON,
	}
	for file, content := range files {
		if err := os.WriteFile(filepath.Join(outDir, file), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func markdownReport(ir architecture.IR, diagrams c4.Models, attackPaths [][]string) string {
	var b strings.Builder
	b.WriteString("# Security Architecture Review\n\n")
	b.WriteString("## C4 Context\n```mermaid\n" + diagrams.Context + "```\n\n")
	b.WriteString("## C4 Container\n```mermaid\n" + diagrams.Container + "```\n\n")
	b.WriteString("## C4 Component\n```mermaid\n" + diagrams.Component + "```\n\n")
	b.WriteString("## Findings\n")
	for _, f := range ir.Findings {
		b.WriteString(fmt.Sprintf("- **%s** [%s] `%s` CWE:%s\n", f.Title, strings.ToUpper(f.Severity), f.Component, f.CWE))
	}
	b.WriteString("\n## Attack Paths\n")
	for _, p := range attackPaths {
		b.WriteString("- " + strings.Join(p, " -> ") + "\n")
	}
	return b.String()
}

func sarifReport(ir architecture.IR) string {
	rules := []string{}
	results := []string{}
	for _, f := range ir.Findings {
		rules = append(rules, fmt.Sprintf(`{"id":"%s","name":"%s","shortDescription":{"text":"%s"}}`, f.ID, f.ID, escape(f.Title)))
		results = append(results, fmt.Sprintf(`{"ruleId":"%s","level":"warning","message":{"text":"%s"}}`, f.ID, escape(f.Title)))
	}
	return `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"duckops-architecture","rules":[` + strings.Join(rules, ",") + `]}},"results":[` + strings.Join(results, ",") + `]}]}`
}

func htmlDashboard(ir architecture.IR, paths [][]string, mermaidSetJSON, cytoscapeJSON, validationJSON, replayJSON string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><title>Threat Model Dashboard</title><script src="https://unpkg.com/react@18/umd/react.production.min.js"></script><script src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script><script src="https://unpkg.com/cytoscape/dist/cytoscape.min.js"></script><script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script><script src="https://d3js.org/d3.v7.min.js"></script></head><body><div id="app"></div><script>
const e = React.createElement;
function App() {
  const findings = ` + fmt.Sprintf("%d", len(ir.Findings)) + `;
  const paths = ` + fmt.Sprintf("%d", len(paths)) + `;
  const mermaidSet = ` + mermaidSetJSON + `;
  const diagramMap = {
    C4Context: mermaidSet.c4_context,
    C4Container: mermaidSet.c4_container,
    C4Component: mermaidSet.c4_component,
    Sequence: mermaidSet.sequence,
    Flowchart: mermaidSet.flowchart,
    ThreatGraph: mermaidSet.threat_graph,
    AttackPaths: mermaidSet.attack_paths,
    TrustBoundaries: mermaidSet.trust_boundaries
  };
  const graphElements = ` + cytoscapeJSON + `;
  const validations = ` + validationJSON + `;
  const replay = ` + replayJSON + `;
  const [diagram, setDiagram] = React.useState("C4Context");
  const [severityFilter, setSeverityFilter] = React.useState("all");
  const [showThreatsOnly, setShowThreatsOnly] = React.useState(false);
  const [replayStep, setReplayStep] = React.useState(0);

  React.useEffect(() => {
    mermaid.initialize({ startOnLoad: false, securityLevel: "loose" });
  }, []);

  React.useEffect(() => {
    const source = diagramMap[diagram] || mermaidSet.c4_context || "";
    const renderTarget = document.getElementById("mermaid-view");
    if (!renderTarget) return;
    renderTarget.innerHTML = source ? ('<pre class="mermaid">' + source + '</pre>') : "<em>No diagram</em>";
    try { mermaid.run({nodes: [renderTarget]}); } catch (_e) {}
  }, [diagram]);

  React.useEffect(() => {
    const el = document.getElementById("graph");
    if (!el) return;
    const filtered = (graphElements || []).filter(item => {
      if (item.group !== "nodes") return true;
      if (severityFilter === "all") return true;
      return (item.data.severity || "safe") === severityFilter;
    }).filter(item => {
      if (!showThreatsOnly) return true;
      if (item.group !== "nodes") return false;
      return item.data.type === "finding";
    });
    const cy = cytoscape({
      container: el,
      elements: filtered,
      style: [
        { selector: "node", style: { "label": "data(label)", "background-color": "#48a868", "color": "#fff", "font-size": 10 } },
        { selector: "node[severity = 'critical']", style: { "background-color": "#d64545" } },
        { selector: "node[severity = 'high']", style: { "background-color": "#e77436" } },
        { selector: "node[severity = 'medium']", style: { "background-color": "#d39a2f" } },
        { selector: "edge", style: { "line-color": "#999", "target-arrow-color": "#999", "target-arrow-shape": "triangle", "curve-style": "bezier", "width": 1.5 } }
      ],
      layout: { name: "cose", animate: false }
    });
    cy.on("tap", "node", evt => {
      const data = evt.target.data();
      const panel = document.getElementById("node-panel");
      if (panel) panel.textContent = JSON.stringify(data, null, 2);
    });
  }, [severityFilter, showThreatsOnly]);

  React.useEffect(() => {
    const heat = document.getElementById("heatmap");
    if (!heat) return;
    heat.innerHTML = "";
    const entries = Object.entries((graphElements || []).reduce((acc, item) => {
      if (item.group === "nodes") {
        const t = item.data.type || "unknown";
        acc[t] = (acc[t] || 0) + 1;
      }
      return acc;
    }, {}));
    const svg = d3.select(heat).append("svg").attr("width", 420).attr("height", 120);
    const x = d3.scaleBand().domain(entries.map(d => d[0])).range([0, 400]).padding(0.2);
    const y = d3.scaleLinear().domain([0, d3.max(entries, d => d[1]) || 1]).range([100, 0]);
    svg.selectAll("rect").data(entries).enter().append("rect")
      .attr("x", d => x(d[0]))
      .attr("y", d => y(d[1]))
      .attr("width", x.bandwidth())
      .attr("height", d => 100 - y(d[1]))
      .attr("fill", "#4e79a7");
    svg.selectAll("text").data(entries).enter().append("text")
      .attr("x", d => x(d[0]) + (x.bandwidth()/2))
      .attr("y", d => y(d[1]) - 5)
      .attr("text-anchor", "middle")
      .attr("fill", "#333")
      .text(d => d[0] + ":" + d[1]);
  }, []);

  const replayPath = replay && replay.length ? replay[0] : [];
  const replayNode = replayPath[replayStep] || null;

  return e("div", {style: {fontFamily: "sans-serif", padding: "16px"}},
    e("h2", null, "Threat Modeling Dashboard"),
    e("p", null, "Findings: " + findings + " | Attack paths: " + paths),
    e("div", {style: {display: "flex", gap: "8px", marginBottom: "8px"}},
      e("select", {value: diagram, onChange: ev => setDiagram(ev.target.value)},
        e("option", {value: "C4Context"}, "C4 Context"),
        e("option", {value: "C4Container"}, "C4 Container"),
        e("option", {value: "C4Component"}, "C4 Component"),
        e("option", {value: "Sequence"}, "Sequence"),
        e("option", {value: "Flowchart"}, "Flowchart"),
        e("option", {value: "ThreatGraph"}, "Threat Graph"),
        e("option", {value: "AttackPaths"}, "Attack Paths"),
        e("option", {value: "TrustBoundaries"}, "Trust Boundaries")
      ),
      e("select", {value: severityFilter, onChange: ev => setSeverityFilter(ev.target.value)},
        e("option", {value: "all"}, "all severities"),
        e("option", {value: "critical"}, "critical"),
        e("option", {value: "high"}, "high"),
        e("option", {value: "medium"}, "medium"),
        e("option", {value: "safe"}, "safe")
      ),
      e("label", null, e("input", {type: "checkbox", checked: showThreatsOnly, onChange: ev => setShowThreatsOnly(ev.target.checked)}), " threats only")
    ),
    e("div", {id: "mermaid-view", style: {border: "1px solid #ccc", padding: "8px", marginBottom: "10px", background: "#fafafa"}}),
    e("div", {style: {display: "grid", gridTemplateColumns: "2fr 1fr", gap: "10px"}},
      e("div", {id: "graph", style: {width: "100%", height: "420px", border: "1px solid #ccc"}}),
      e("pre", {id: "node-panel", style: {height: "420px", border: "1px solid #ccc", margin: 0, overflow: "auto", padding: "8px"}}, "Click a node to inspect...")
    ),
    e("h3", null, "Security Heatmap"),
    e("div", {id: "heatmap"}),
    e("h3", null, "Attack Replay"),
    e("div", null, "Current step node: " + (replayNode ? replayNode.node : "none")),
    e("input", {type: "range", min: 0, max: Math.max(0, replayPath.length - 1), value: replayStep, onChange: ev => setReplayStep(parseInt(ev.target.value || "0", 10))}),
    e("h3", null, "Diagram Validation"),
    e("pre", {style: {border: "1px solid #ddd", padding: "8px"}}, JSON.stringify(validations, null, 2))
  );
}
ReactDOM.render(e(App), document.getElementById("app"));
</script></body></html>`
}

func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
}
