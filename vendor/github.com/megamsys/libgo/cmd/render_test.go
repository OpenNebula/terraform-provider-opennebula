package cmd

import (
	"gopkg.in/check.v1"
	"sort"
)

func (s *S) TestAddOneRow(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"Three", "foo"})
	c.Assert(table.String(), check.Equals, "+-------+-----+\n| Three | foo |\n+-------+-----+\n")
}

func (s *S) TestAddRows(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	expected := `+-------+---+
| One   | 1 |
| Two   | 2 |
| Three | 3 |
+-------+---+
`
	c.Assert(table.String(), check.Equals, expected)
}

func (s *S) TestRows(c *check.C) {
	table := NewTable()
	c.Assert(table.Rows(), check.Equals, 0)
	table.AddRow(Row{"One", "1"})
	c.Assert(table.Rows(), check.Equals, 1)
	table.AddRow(Row{"One", "1"})
	c.Assert(table.Rows(), check.Equals, 2)
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"One", "1"})
	c.Assert(table.Rows(), check.Equals, 5)
}

func (s *S) TestSort(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"Three", "3"})
	table.AddRow(Row{"Zero", "0"})
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	expected := `+-------+---+
| One   | 1 |
| Three | 3 |
| Two   | 2 |
| Zero  | 0 |
+-------+---+
`
	table.Sort()
	c.Assert(table.String(), check.Equals, expected)
}

func (s *S) TestColumnsSize(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	c.Assert(table.columnsSize(), check.DeepEquals, []int{5, 1})
}

func (s *S) TestSeparator(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	expected := "+-------+---+\n"
	c.Assert(table.separator(), check.Equals, expected)
}

func (s *S) TestHeadings(c *check.C) {
	table := NewTable()
	table.Headers = Row{"Word", "Number"}
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	expected := `+-------+--------+
| Word  | Number |
+-------+--------+
| One   | 1      |
| Two   | 2      |
| Three | 3      |
+-------+--------+
`
	c.Assert(table.String(), check.Equals, expected)
}

func (s *S) TestString(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	expected := `+-------+---+
| One   | 1 |
| Two   | 2 |
| Three | 3 |
+-------+---+
`
	c.Assert(table.String(), check.Equals, expected)
}

func (s *S) TestRenderNoRows(c *check.C) {
	table := NewTable()
	table.Headers = Row{"Word", "Number"}
	expected := `+------+--------+
| Word | Number |
+------+--------+
+------+--------+
`
	c.Assert(table.String(), check.Equals, expected)
}

func (s *S) TestRenderEmpty(c *check.C) {
	table := NewTable()
	c.Assert(table.String(), check.Equals, "")
}

func (s *S) TestBytes(c *check.C) {
	table := NewTable()
	table.AddRow(Row{"One", "1"})
	table.AddRow(Row{"Two", "2"})
	table.AddRow(Row{"Three", "3"})
	c.Assert(table.Bytes(), check.DeepEquals, []byte(table.String()))
}

func (s *S) TestRowListAdd(c *check.C) {
	l := rowSlice([]Row{{"one", "1"}})
	l.add(Row{"two", "2"})
	c.Assert(len(l), check.Equals, 2)
}

func (s *S) TestRowListLen(c *check.C) {
	l := rowSlice([]Row{{"one", "1"}})
	c.Assert(l.Len(), check.Equals, 1)
	l.add(Row{"two", "2"})
	c.Assert(l.Len(), check.Equals, 2)
}

func (s *S) TestRowListLess(c *check.C) {
	l := rowSlice([]Row{{"zero", "0"}, {"one", "1"}, {"two", "2"}})
	c.Assert(l.Less(0, 1), check.Equals, false)
	c.Assert(l.Less(0, 2), check.Equals, false)
	c.Assert(l.Less(1, 2), check.Equals, true)
	c.Assert(l.Less(1, 0), check.Equals, true)
}

func (s *S) TestRowListLessDifferentCase(c *check.C) {
	l := rowSlice([]Row{{"Zero", "0"}, {"one", "1"}, {"two", "2"}})
	c.Assert(l.Less(0, 1), check.Equals, false)
	c.Assert(l.Less(0, 2), check.Equals, false)
	c.Assert(l.Less(1, 2), check.Equals, true)
	c.Assert(l.Less(1, 0), check.Equals, true)
}

func (s *S) TestRowListSwap(c *check.C) {
	l := rowSlice([]Row{{"zero", "0"}, {"one", "1"}, {"two", "2"}})
	l.Swap(0, 2)
	c.Assert(l.Less(0, 2), check.Equals, true)
}

func (s *S) TestRowListIsSortable(c *check.C) {
	var _ sort.Interface = rowSlice{}
}

func (s *S) TestColorRed(c *check.C) {
	output := Colorfy("must return a red font pattern", "red", "", "")
	c.Assert(output, check.Equals, "\033[0;31;10mmust return a red font pattern\033[0m")
}

func (s *S) TestColorGreen(c *check.C) {
	output := Colorfy("must return a green font pattern", "green", "", "")
	c.Assert(output, check.Equals, "\033[0;32;10mmust return a green font pattern\033[0m")
}

func (s *S) TestColorBoldWhite(c *check.C) {
	output := Colorfy("must return a bold white font pattern", "white", "", "bold")
	c.Assert(output, check.Equals, "\033[1;37;10mmust return a bold white font pattern\033[0m")
}

func (s *S) TestColorBoldYellowGreenBG(c *check.C) {
	output := Colorfy("must return a bold yellow with green background", "yellow", "green", "bold")
	c.Assert(output, check.Equals, "\033[1;33;42mmust return a bold yellow with green background\033[0m")
}
