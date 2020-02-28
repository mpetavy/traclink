package main

import (
	"flag"
	"fmt"
	"time"

	"strconv"

	"os/exec"

	"strings"

	"bufio"

	"regexp"

	"github.com/atotto/clipboard"
	"github.com/mpetavy/common"
)

var (
	url    *string
	trac   *string
	search *string
	rang   *string
	table  *bool
)

const (
	default_url = "http://svn-medmuc/lehel/trunk"
)

func init() {
	common.Init("1.0.5", "2018", "generates markdown links to TRAC changesets", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, false, nil, nil, run, 0)

	url = flag.String("u", default_url, "URL to the SVN repository or relative path")
	trac = flag.String("t", "http://trac-medmuc/trac/lehel/changeset/", "TRAC prefix")
	search = flag.String("s", "", "Search string to look for. Multiple searches separated by ;")
	rang = flag.String("r", "{"+strconv.Itoa(time.Now().Year())+"-5-1}:HEAD", fmt.Sprintf("SVN log range"))
	table = flag.Bool("tab", true, "Output as Markdown table or plain text")
}

func run() error {
	if strings.Index(*url, "//") == -1 {
		*url = default_url[:strings.LastIndex(default_url, "/")+1] + *url
	}

	p := []string{"log", *url}

	if len(*rang) > 0 {
		p = append(p, "-r")
		p = append(p, *rang)
	}

	if len(*search) > 0 {
		ss := strings.Split(*search, ";")

		for _, s := range ss {
			p = append(p, "--search")
			p = append(p, s)
		}
	}

	path, err := exec.LookPath("svn")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, p...)

	fmt.Printf("Scan SVN log: %s ... ", common.CmdToString(cmd))

	common.Debug("exec: %s", common.CmdToString(cmd))

	b, err := cmd.Output()
	output, err := common.ToUTF8String(string(b[:]), common.DefaultEncoding())

	if err != nil {
		return err
	}

	var lines []string
	var svn string
	var description string

	blc := -1

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		readLine := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(readLine, "------") {
			if len(description) > 0 {
				var re = regexp.MustCompile("^LEH-(.*?) ")
				description := re.ReplaceAllString(description, "")

				for {
					if len(description) > 0 && (strings.HasPrefix(description, "-") || strings.HasPrefix(description, " ")) {
						description = description[1:]
					} else {
						break
					}
				}

				//for _, r := range "`*_{}[]()#>+-.!" {
				for _, r := range "|" {
					description = strings.Replace(description, string(r), "\\"+string(r), -1)
				}

				lines = append(lines, svn+" "+description+common.Eval(*table, " |", "").(string))
			}
			readLine = ""
			description = ""
			blc = 0
		}

		if blc == 1 {
			items := strings.Split(readLine, " | ")
			rev := strings.TrimSpace(items[0][1:])

			items[0] = "[" + rev + "|" + *trac + rev + "]"

			if *table {
				svn = "| " + items[0] + " |"
			} else {
				svn = items[0]
			}

			readLine = ""
		}

		if len(readLine) > 0 && blc > 1 {
			if blc > 3 {
				description += " "
			}

			description += readLine
		}
		blc++
	}

	fmt.Printf("\n\n")
	if len(lines) == 0 {
		fmt.Println("SVN search did not return any logs")

		return nil
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	clipboard.WriteAll(strings.Join(lines, "\n"))

	return nil
}

func main() {
	defer common.Done()

	common.Run([]string{"s"})
}
