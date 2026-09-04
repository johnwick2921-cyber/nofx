// Settings integrity (D5) — GET /api/config/resolved.
//
// One place a client can ask "what does this build actually honour?". The
// answer is the registry's own classification and the registry's own label:
// the UI renders what is here, it does not re-derive a status from a field
// name or invent wording for one.
//
// The owner's two-label ruling is the reason this endpoint exists rather than
// a boolean "active" flag. "No consumer" and "reads it but cannot take effect"
// are different findings and must stay legible as different findings:
//
//	ineffective            read; does not take effect (<reason>)
//	candidate-unverified   no known reader — pending verification
//
// Neither reads as dead. A candidate stays listed until someone runs the
// method-level grep and quotes it.
//
// A25: KnobEntry holds no values, and this payload adds none. Infra knobs are
// ports, paths and keys; classification is safe to serve, values are not.

package api

import (
	"net/http"

	"nofx/store"

	"github.com/gin-gonic/gin"
)

type resolvedKnob struct {
	Path      string   `json:"path"`
	Status    string   `json:"status"`
	UILabel   string   `json:"ui_label"`
	Consumers []string `json:"consumers"`
	DualLevel bool     `json:"dual_level"`
	Clamp     string   `json:"clamp,omitempty"`
	Note      string   `json:"note,omitempty"`
}

type resolvedSummary struct {
	Schema              int      `json:"schema"`
	Classified          int      `json:"classified"`
	Live                int      `json:"live"`
	Ineffective         int      `json:"ineffective"`
	CandidateUnverified int      `json:"candidate_unverified"`
	Suspended           int      `json:"suspended"`
	Advisory            int      `json:"advisory"`
	DisplayOnly         int      `json:"display_only"`
	Infra               int      `json:"infra"`
	EnvShadows          int      `json:"env_shadows"`
	EnvShadowPaths      []string `json:"env_shadow_paths"`
}

// handleConfigResolved serves the knob registry. The counts are taken from the
// same KnobStatusSummary the ⚙ boot line prints, so the page and the log can
// never disagree about how many knobs take effect.
func (s *Server) handleConfigResolved(c *gin.Context) {
	sum := store.KnobStatusSummary()

	entries := store.AllKnobs()
	knobs := make([]resolvedKnob, 0, len(entries))
	for _, e := range entries {
		consumers := e.Consumers
		if consumers == nil {
			// An empty computed list is [], never null: a knob with no
			// recorded consumer is a finding, not missing data.
			consumers = []string{}
		}
		knobs = append(knobs, resolvedKnob{
			Path:      e.Path,
			Status:    string(e.Status),
			UILabel:   e.UILabel(),
			Consumers: consumers,
			DualLevel: e.DualLevel,
			Clamp:     e.Clamp,
			Note:      e.Note,
		})
	}

	paths := sum.EnvShadowPaths
	if paths == nil {
		paths = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": resolvedSummary{
			Schema:              len(store.EnumerateSchemaKnobs()),
			Classified:          sum.Total,
			Live:                sum.Live,
			Ineffective:         sum.Ineffective,
			CandidateUnverified: sum.Candidate,
			Suspended:           sum.Suspended,
			Advisory:            sum.Advisory,
			DisplayOnly:         sum.DisplayOnly,
			Infra:               sum.Infra,
			EnvShadows:          sum.EnvShadows,
			EnvShadowPaths:      paths,
		},
		"knobs": knobs,
	})
}
