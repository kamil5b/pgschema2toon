package toon

type Table struct {
	Name        string       `json:"table"`
	Comment     string       `json:"comment,omitempty"`
	Columns     []Column     `json:"columns"`
	Constraints []Constraint `json:"constraints,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
}

type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Comment  string `json:"comment,omitempty"`
	Nullable bool   `json:"nullable,omitempty"`
	IsPK     bool   `json:"is_pk,omitempty"`
}

type Constraint struct {
	Def string `json:"def"`
}

type Index struct {
	Name string `json:"name"`
	Def  string `json:"def"`
}
