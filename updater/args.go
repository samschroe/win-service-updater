package updater

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// https://wyday.com/wybuild/help/wyupdate-commandline.php

// rc, _ := try(WYUPDATE_EXE, "/quickcheck", "/justcheck", "/noerr",
// 	fmt.Sprintf("-urlargs=%s", AUTH), fmt.Sprintf("/outputinfo=%s", CHECK_LOG))

// wyupdateArgs := fmt.Sprintf("/fromservice -logfile=\"%s\" -urlargs=%s",
// 	WYUPDATE_LOG, AUTH)

// Args contains the parsed command-line arguments
type Args struct {
	Debug         bool
	Quickcheck    bool
	Justcheck     bool
	Noerr         bool
	Fromservice   bool
	Urlargs       string
	Outputinfo    bool
	OutputinfoLog string
	Logfile       string
	Cdata         string
	WYSTestServer string // Used for testing
	WYUTestServer string // Used for testing
}

var argRegexp *regexp.Regexp = regexp.MustCompile(`^/`)

// ParseArgs returns a struct with the parsed command-line arguments
func ParseArgs(argsSlice []string) (args Args, err error) {
	// remove the program argument
	argsSlice = argsSlice[1:]

	fs := flag.NewFlagSet("win-service-updater", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)

	// translate windows-style command line args to the normal
	// Golang style (- in the front as opposed to /)
	normalizedArgs := make([]string, 0, len(argsSlice))
	for _, arg := range argsSlice {
		normalizedArgs = append(normalizedArgs, argRegexp.ReplaceAllLiteralString(strings.ToLower(arg), "-"))
	}

	// debug := fs.Bool("debug", false, "Whether or not to log debug messages")
	fs.BoolVar(&args.Debug, "debug", false, "Whether or not to run as debug")
	fs.BoolVar(&args.Quickcheck, "quickcheck", false, "Whether or not to run a quickcheck")
	fs.BoolVar(&args.Justcheck, "justcheck", false, "Whether or not to run a justcheck")
	fs.BoolVar(&args.Noerr, "noerr", false, "Whether or not to error")
	fs.BoolVar(&args.Fromservice, "fromservice", false, "Whether or not to run from a service")
	fs.StringVar(&args.Urlargs, "urlargs", "", "Additonal string to add onto the URL")
	fs.StringVar(&args.Logfile, "logfile", "", "Name of log file")
	fs.StringVar(&args.OutputinfoLog, "outputinfo", "", "Output info")
	// default to client.wyc
	fs.StringVar(&args.Cdata, "cdata", filepath.Join(GetExeDir(), CLIENT_WYC), "Config data")
	// TODO: These overrides should only be available in a debug build, not in what gets shipped in production
	fs.StringVar(&args.WYSTestServer, "wysserver", "", "WYS Server")
	fs.StringVar(&args.WYUTestServer, "wyuserver", "", "WYU Server")

	err = fs.Parse(normalizedArgs)
	if err != nil {
		return args, err
	}

	// check to see if outputinfo was set. If so set outputinfo
	// bool to true
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "outputinfo" {
			args.Outputinfo = true
		}
	})

	return args, nil
}
