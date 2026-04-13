package toon

import (
	"encoding/json"
	"fmt"
	"strings"
)
func ToToon(data []byte) (string, error) {
	var schema []Table
	if err := json.Unmarshal(data, &schema); err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, t := range schema {
		// 1. Map constraints to columns for inline references
		// Key: local_column_name, Value: referenced_table(ref_column)
		inlineRefs := make(map[string]string)
		var multiColRefs []string

		for _, con := range t.Constraints {
			if strings.Contains(con.Def, "REFERENCES") {
				// Parse: FOREIGN KEY (col) REFERENCES table(id)
				// We look for the content between the first set of parentheses
				parts := strings.SplitN(con.Def, "(", 2)
				if len(parts) < 2 { continue }

				colPart := strings.SplitN(parts[1], ")", 2)
				if len(colPart) < 1 { continue }

				colName := strings.TrimSpace(colPart[0])
				refTarget := strings.SplitN(con.Def, "REFERENCES ", 2)[1]

				// Only inline if it's a single column.
				// If it contains a comma, it's a composite key; keep it at the bottom.
				if !strings.Contains(colName, ",") {
					inlineRefs[colName] = "-> " + refTarget
				} else {
					multiColRefs = append(multiColRefs, refTarget)
				}
			}
		}

		sb.WriteString(fmt.Sprintf("[%s]\n", t.Name))
		if t.Comment != "" {
			sb.WriteString(fmt.Sprintf("# %s\n", t.Comment))
		}

		// 2. Render Columns
		for _, col := range t.Columns {
			var tags []string
			if col.IsPK { tags = append(tags, "pk") }
			if !col.Nullable { tags = append(tags, "req") }

			tagStr := ""
			if len(tags) > 0 { tagStr = " {" + strings.Join(tags, ",") + "}" }

			// Check for inline reference
			refStr := ""
			if r, ok := inlineRefs[col.Name]; ok {
				refStr = " " + r
			}

			commentStr := ""
			if col.Comment != "" { commentStr = " // " + col.Comment }

			// Construct line: name type {tags} -> ref // comment
			sb.WriteString(fmt.Sprintf("  %s %s%s%s%s\n",
				col.Name, shrink(col.Type), tagStr, refStr, commentStr))
		}

		// 3. Render Multi-column references (if any)
		for _, ref := range multiColRefs {
			sb.WriteString(fmt.Sprintf("  ref -> %s\n", ref))
		}

		// 4. Render Indices
		if len(t.Indexes) > 0 {
			sb.WriteString("@indices\n")
			for _, idx := range t.Indexes {
				parts := strings.Split(idx.Def, " USING ")
				def := idx.Def
				if len(parts) > 1 { def = parts[1] }
				sb.WriteString(fmt.Sprintf("  %s: %s\n", idx.Name, def))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func shrink(t string) string {
	t = strings.ReplaceAll(t, "character varying", "varchar")
	t = strings.ReplaceAll(t, "timestamp with time zone", "timestamptz")
	return strings.TrimSpace(t)
}
