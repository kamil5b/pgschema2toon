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
		sb.WriteString(fmt.Sprintf("[%s]\n", t.Name))
		if t.Comment != "" {
			sb.WriteString(fmt.Sprintf("# %s\n", t.Comment))
		}

		for _, col := range t.Columns {
			var tags []string
			if col.IsPK { tags = append(tags, "pk") }
			if !col.Nullable { tags = append(tags, "req") }

			tagStr := ""
			if len(tags) > 0 { tagStr = " {" + strings.Join(tags, ",") + "}" }

			commentStr := ""
			if col.Comment != "" { commentStr = " // " + col.Comment }

			sb.WriteString(fmt.Sprintf("  %s %s%s%s\n", col.Name, shrink(col.Type), tagStr, commentStr))
		}

		for _, con := range t.Constraints {
			if strings.Contains(con.Def, "REFERENCES") {
				ref := strings.Split(con.Def, "REFERENCES ")[1]
				sb.WriteString(fmt.Sprintf("  ref -> %s\n", ref))
			}
		}

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
