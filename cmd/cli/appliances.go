package main

import (
	"context"
	"fmt"
	"forester/internal/api/ctl"
	"net/http"
)

type applianceCreateCmd struct {
	Name string `arg:"-n,required"`
	Kind string `arg:"-k" default:"noop"`
	URI  string `arg:"-u" default:"noop:///"`
}

type applianceListCmd struct {
	Limit  int64 `arg:"-m" default:"100"`
	Offset int64 `arg:"-o" default:"0"`
}

type applianceEnlistCmd struct {
	Name          string `arg:"positional,required" placeholder:"APPLIANCE_NAME"`
	SystemPattern string `arg:"-n" placeholder:"REGEXP_SYSTEM_PATTERN" default:".*"`
}

type applianceCmd struct {
	Create *applianceCreateCmd `arg:"subcommand:create" help:"create appliance"`
	List   *applianceListCmd   `arg:"subcommand:list" help:"list appliances"`
	Enlist *applianceEnlistCmd `arg:"subcommand:enlist" help:"enlist systems of appliance"`
}

func applianceCreate(ctx context.Context, cmdArgs *applianceCreateCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	err := client.Create(ctx, cmdArgs.Name, ctl.ApplianceKindToInt(cmdArgs.Kind), cmdArgs.URI)
	if err != nil {
		return fmt.Errorf("cannot create appliance: %w", err)
	}

	return nil
}

func applianceList(ctx context.Context, cmdArgs *applianceListCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	appliances, err := client.List(ctx, cmdArgs.Limit, cmdArgs.Offset)
	if err != nil {
		return fmt.Errorf("cannot list appliances: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "ID\tName\tKind\tURI")
	for _, a := range appliances {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", a.ID, a.Name, ctl.ApplianceIntToKind(a.Kind), a.URI)
	}
	w.Flush()

	return nil
}

func applianceEnlist(ctx context.Context, cmdArgs *applianceEnlistCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	err := client.Enlist(ctx, cmdArgs.Name, cmdArgs.SystemPattern)
	if err != nil {
		return fmt.Errorf("cannot enlist systems: %w", err)
	}

	return nil
}
