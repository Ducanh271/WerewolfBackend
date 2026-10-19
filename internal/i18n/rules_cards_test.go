package i18n_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	_ "masoi/internal/roles" // every role registers itself, so CardOrder is complete
)

// rulesPath is docs/rules.md seen from this package's directory.
const rulesPath = "../../docs/rules.md"

// nineB is what the cards are checked against in rules §9b.
type nineB struct {
	roles      []roleRow         // the role table, in order
	statuses   map[string]string // "Thẻ phụ": name → text, "X" for the partner
	goals      string            // the "Mục tiêu theo phe:" line, without the label
	intro      string            // the text between the heading and the role table
	statusRows int
}

type roleRow struct {
	line                    int
	name, team, description string
}

// parseNineB reads rules §9b. Every error names the line or the part of the
// section it could not find, so a change of layout in rules.md fails clearly.
func parseNineB(text string) (nineB, error) {
	var out nineB
	lines := strings.Split(text, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "## 9b.") {
			start = i
			break
		}
	}
	if start < 0 {
		return out, fmt.Errorf(`rules.md: heading "## 9b." not found`)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	sec := lines[start:end]
	at := func(i int) int { return start + i + 1 } // 1-based line in rules.md

	roleHead := indexOf(sec, "| Vai | Phe | Mô tả ngắn |")
	if roleHead < 0 {
		return out, fmt.Errorf(`rules.md §9b (line %d): role table header "| Vai | Phe | Mô tả ngắn |" not found`, at(0))
	}
	out.intro = strings.Join(sec[1:roleHead], "\n")
	rows, next, err := tableRows(sec, roleHead, 3, at)
	if err != nil {
		return out, fmt.Errorf("rules.md §9b role table: %w", err)
	}
	for _, r := range rows {
		out.roles = append(out.roles, roleRow{line: r.line, name: r.cells[0], team: r.cells[1], description: r.cells[2]})
	}

	const goalLabel = "Mục tiêu theo phe: "
	for _, l := range sec[next:] {
		if strings.HasPrefix(l, goalLabel) {
			out.goals = strings.TrimPrefix(l, goalLabel)
			break
		}
	}
	if out.goals == "" {
		return out, fmt.Errorf("rules.md §9b: no line starting with %q after the role table (line %d)", goalLabel, at(next))
	}

	sub := indexOf(sec, "### Thẻ phụ")
	if sub < 0 {
		return out, fmt.Errorf(`rules.md §9b: subheading "### Thẻ phụ" not found`)
	}
	statusHead := -1
	for i := sub; i < len(sec); i++ {
		if strings.TrimSpace(sec[i]) == "| Trạng thái | Nội dung |" {
			statusHead = i
			break
		}
	}
	if statusHead < 0 {
		return out, fmt.Errorf(`rules.md §9b "Thẻ phụ" (line %d): table header "| Trạng thái | Nội dung |" not found`, at(sub))
	}
	rows, _, err = tableRows(sec, statusHead, 2, at)
	if err != nil {
		return out, fmt.Errorf(`rules.md §9b "Thẻ phụ" table: %w`, err)
	}
	out.statuses = map[string]string{}
	for _, r := range rows {
		out.statuses[r.cells[0]] = r.cells[1]
	}
	out.statusRows = len(rows)
	return out, nil
}

type tableRow struct {
	line  int
	cells []string
}

// tableRows reads the rows of the markdown table whose header is sec[head]:
// a separator line, then rows of exactly n cells until the first line that is
// not a row. next is the index of that line.
func tableRows(sec []string, head, n int, at func(int) int) (rows []tableRow, next int, err error) {
	sep := head + 1
	if sep >= len(sec) || !strings.HasPrefix(strings.TrimSpace(sec[sep]), "| ---") {
		return nil, 0, fmt.Errorf("line %d: want the separator row \"| --- |...\" right under the header", at(sep))
	}
	i := sep + 1
	for ; i < len(sec) && strings.HasPrefix(strings.TrimSpace(sec[i]), "|"); i++ {
		line := strings.TrimSpace(sec[i])
		if !strings.HasSuffix(line, "|") {
			return nil, 0, fmt.Errorf("line %d: row does not end with \"|\": %q", at(i), line)
		}
		cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"), "|")
		if len(cells) != n {
			return nil, 0, fmt.Errorf("line %d: row has %d cells, want %d: %q", at(i), len(cells), n, line)
		}
		for j := range cells {
			cells[j] = strings.TrimSpace(cells[j])
		}
		rows = append(rows, tableRow{line: at(i), cells: cells})
	}
	if len(rows) == 0 {
		return nil, 0, fmt.Errorf("line %d: the table has no rows", at(sep+1))
	}
	return rows, i, nil
}

func indexOf(lines []string, want string) int {
	for i, l := range lines {
		if strings.TrimSpace(l) == want {
			return i
		}
	}
	return -1
}

func readNineB(t *testing.T) nineB {
	t.Helper()
	raw, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read %s: %v", rulesPath, err)
	}
	nb, err := parseNineB(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	return nb
}

// TestCards_MatchRulesNineB: the card text players see (vi.go) is the text of
// rules §9b, word for word. Changing one without the other fails here.
func TestCards_MatchRulesNineB(t *testing.T) {
	nb := readNineB(t)
	order := engine.CardOrder()

	t.Run("one card per row, in the table's order", func(t *testing.T) {
		var want, got []string
		for _, r := range nb.roles {
			want = append(want, r.name)
		}
		for _, id := range order {
			c, ok := i18n.Card(string(id))
			if !ok {
				t.Fatalf("role %s has no card in vi.go", id)
			}
			got = append(got, c.Name)
		}
		if strings.Join(got, ", ") != strings.Join(want, ", ") {
			t.Fatalf("role list order (engine.CardOrder) differs from the rules §9b table\n got: %s\nwant: %s",
				strings.Join(got, ", "), strings.Join(want, ", "))
		}
	})

	t.Run("name, team, goal and description of each card", func(t *testing.T) {
		sides := map[string]string{"Dân": i18n.SideVillage, "Sói": i18n.SideWolves}
		for i, id := range order {
			if i >= len(nb.roles) {
				break // reported by the order subtest
			}
			row := nb.roles[i]
			c, _ := i18n.Card(string(id))
			side, ok := sides[row.team]
			if !ok {
				t.Errorf("rules.md:%d: team %q of %s is neither \"Dân\" nor \"Sói\"", row.line, row.team, row.name)
				continue
			}
			team, goal := i18n.Side(side)
			if c.Name != row.name || c.Team != team || c.Goal != goal || c.Description != row.description {
				t.Errorf("card of %s differs from rules.md:%d\n got: %+v\nwant: {Name:%s Team:%s Goal:%s Description:%s}",
					id, row.line, c, row.name, team, goal, row.description)
			}
		}
	})

	t.Run("goals of the sides", func(t *testing.T) {
		_, village := i18n.Side(i18n.SideVillage)
		_, wolves := i18n.Side(i18n.SideWolves)
		if want := village + " " + wolves; nb.goals != want {
			t.Errorf("\"Mục tiêu theo phe\" in rules §9b = %q, vi.go has %q", nb.goals, want)
		}
		third, goal := i18n.Side(i18n.SideThird)
		quoted := "“" + third + "” với mục tiêu “" + goal + "”"
		if !strings.Contains(nb.intro, quoted) {
			t.Errorf("rules §9b intro does not contain the third side's card text %s", quoted)
		}
	})

	t.Run("extra cards (Thẻ phụ)", func(t *testing.T) {
		kinds := []string{i18n.StatusCursed, i18n.StatusLovers, i18n.StatusLoversCross, i18n.StatusChief}
		if nb.statusRows != len(kinds) {
			t.Errorf("rules §9b \"Thẻ phụ\" has %d rows, vi.go has %d extra cards", nb.statusRows, len(kinds))
		}
		for _, kind := range kinds {
			name, text := i18n.StatusCard(kind, "X") // rules.md writes the partner as X
			want, ok := nb.statuses[name]
			if !ok {
				t.Errorf("extra card %q (%s) has no row in rules §9b \"Thẻ phụ\"", name, kind)
				continue
			}
			if text != want {
				t.Errorf("extra card %q differs from rules §9b\n got: %s\nwant: %s", name, text, want)
			}
		}
	})
}

// TestParseNineB_ReportsWhereItFails: a layout change in rules.md names what is
// missing or which line is wrong, instead of a confusing card mismatch.
func TestParseNineB_ReportsWhereItFails(t *testing.T) {
	good := strings.Join([]string{
		"## 9b. Thẻ vai",
		"intro",
		"| Vai | Phe | Mô tả ngắn |",
		"| --- | --- | --- |",
		"| Sói | Sói | cắn |",
		"",
		"Mục tiêu theo phe: a b",
		"### Thẻ phụ",
		"| Trạng thái | Nội dung |",
		"| --- | --- |",
		"| Bị nguyền | text |",
		"## 10. Next",
	}, "\n")
	if _, err := parseNineB(good); err != nil {
		t.Fatalf("well-formed section refused: %v", err)
	}
	tests := []struct {
		name, old, new, want string
	}{
		{"heading renamed", "## 9b. Thẻ vai", "## 9c. Thẻ vai", `heading "## 9b." not found`},
		{"role table header changed", "| Vai | Phe | Mô tả ngắn |", "| Vai | Mô tả |", "role table header"},
		{"separator missing", "| --- | --- | --- |\n", "", "line 4: want the separator row"},
		{"extra cell in a row", "| Sói | Sói | cắn |", "| Sói | Sói | cắn | x |", "line 5: row has 4 cells, want 3"},
		{"goal line missing", "Mục tiêu theo phe: a b", "Mục tiêu: a b", `no line starting with "Mục tiêu theo phe: "`},
		{"extra cards subheading missing", "### Thẻ phụ", "### Khác", `"### Thẻ phụ" not found`},
		{"extra cards header changed", "| Trạng thái | Nội dung |", "| Thẻ | Nội dung |", `"Thẻ phụ" (line 8): table header`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseNineB(strings.Replace(good, tc.old, tc.new, 1))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
