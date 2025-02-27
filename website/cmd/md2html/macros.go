// The macros program implements an ad-hoc preprocessor for Markdown files, used
// in Elvish's website.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"src.elv.sh/pkg/mods/doc"
)

func filter(in io.Reader, out io.Writer) {
	f := filterer{}
	f.filter(in, out)
}

type filterer struct {
	module string
}

var macros = map[string]func(*filterer, string) string{
	"@module ":  (*filterer).expandModule,
	"@ttyshot ": (*filterer).expandTtyshot,
	"@dl ":      (*filterer).expandDl,
}

func (f *filterer) filter(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		for leader, expand := range macros {
			i := strings.Index(line, leader)
			if i >= 0 {
				line = line[:i] + expand(f, line[i+len(leader):])
				break
			}
		}
		fmt.Fprintln(out, line)
	}
	if f.module != "" {
		ns := f.module + ":"
		if f.module == "builtin" {
			ns = ""
		}
		docs := doc.Docs()[ns]

		var buf bytes.Buffer
		writeElvdocSections(&buf, ns, docs)
		filter(&buf, out)
	}
}

func (f *filterer) expandModule(rest string) string {
	f.module = rest
	// Module doc will be added at end of file
	return fmt.Sprintf(
		"<a name='//apple_ref/cpp/Module/%s' class='dashAnchor'></a>", f.module)
}

func (f *filterer) expandTtyshot(name string) string {
	content, err := os.ReadFile(name + ".ttyshot.html")
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(`<pre class="ttyshot"><code>%s</code></pre>`,
		bytes.Replace(content, []byte("\n"), []byte("<br>"), -1))
}

func (f *filterer) expandDl(rest string) string {
	fields := strings.SplitN(rest, " ", 2)
	name := fields[0]
	url := name
	if len(fields) == 2 {
		url = fields[1]
	}
	return fmt.Sprintf(`<a href="https://dl.elv.sh/%s">%s</a>`, url, name)
}
