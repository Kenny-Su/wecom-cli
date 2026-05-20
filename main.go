package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, rest, err := parseGlobalFlags(args)
	if err != nil {
		return err
	}
	if len(rest) == 0 || isHelp(rest[0]) {
		printUsage()
		return nil
	}

	switch rest[0] {
	case "calendar":
		if len(rest) == 1 || isHelp(rest[1]) {
			printCalendarUsage()
			return nil
		}
		c := &wecomClient{cfg: cfg, http: cfg.HTTPClient}
		return runCalendar(c, rest[1:])
	case "schedule":
		if len(rest) == 1 || isHelp(rest[1]) {
			printScheduleUsage()
			return nil
		}
		c := &wecomClient{cfg: cfg, http: cfg.HTTPClient}
		return runSchedule(c, rest[1:])
	case "meeting":
		if len(rest) == 1 || isHelp(rest[1]) {
			printMeetingUsage()
			return nil
		}
		c := &wecomClient{cfg: cfg, http: cfg.HTTPClient}
		return runMeeting(c, rest[1:])
	default:
		return fmt.Errorf("unknown command %q", rest[0])
	}
}
