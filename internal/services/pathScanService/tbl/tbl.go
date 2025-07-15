package tbl

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func PrintScanResultsTable(rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tSize\tSizeParsed\tCreated\tModified\tOwner\tPermissions")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row[0], row[1], row[2], row[3], row[4], row[5], row[6])
	}
	w.Flush()
}
