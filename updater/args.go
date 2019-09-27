package updater

import (
	"fmt"
	"path/filepath"
	"strings"
)

// https://wyday.com/wybuild/help/wyupdate-commandline.php

// rc, _ := try(WYUPDATE_EXE, "/quickcheck", "/justcheck", "/noerr",
// 	fmt.Sprintf("-urlargs=%s", AUTH), fmt.Sprintf("/outputinfo=%s", CHECK_LOG))

// wyupdateArgs := fmt.Sprintf("/fromservice -logfile=\"%s\" -urlargs=%s",
// 	WYUPDATE_LOG, AUTH)

// Args contains the parsed command-line arguments
type Args struct {
	Quickcheck    bool
	Justcheck     bool
	Noerr         bool
	Fromservice   bool
	Urlargs       string
	Outputinfo    bool
	OutputinfoLog string
	Logfile       string
	Cdata         string
	Server        string
	WYUTestServer string // Used for testing
}

// ParseArgs returns a struct with the parsed command-line arguments
func ParseArgs(argsSlice []string) (args Args, err error) {
	// default to client.wyc
	args.Cdata = filepath.Join(GetExeDir(), CLIENT_WYC)

	for _, arg := range argsSlice {
		larg := strings.ToLower(arg)

		switch {
		case larg == "/quickcheck":
			args.Quickcheck = true
		case larg == "/justcheck":
			args.Justcheck = true
		case larg == "/noerr":
			args.Noerr = true
		case larg == "/fromservice":
			args.Fromservice = true
		case strings.HasPrefix(larg, "-urlargs="):
			fields := strings.Split(larg, "=")
			args.Urlargs = fields[1]
		case strings.HasPrefix(larg, "-logfile="):
			fields := strings.Split(larg, "=")
			args.Logfile = fields[1]
		case strings.HasPrefix(larg, "/outputinfo"):
			args.Outputinfo = true
			if strings.Contains(larg, "=") {
				fields := strings.Split(larg, "=")
				args.OutputinfoLog = fields[1]
			}
		case strings.HasPrefix(larg, "-cdata="):
			fields := strings.Split(larg, "=")
			args.Cdata = fields[1]
		case strings.HasPrefix(larg, "-server="):
			fields := strings.Split(larg, "=")
			args.Server = fields[1]
		default:
			err = fmt.Errorf("unknown argument '%s'", larg)
			return args, err
		}
	}
	return args, nil
}
