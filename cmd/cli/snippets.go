package main

import (
	"context"
	"fmt"
	"forester/cmd/cli/edit"
	"forester/internal/api/ctl"
	"net/http"
	"strings"
)

type snippetCreateCmd struct {
	Name string `arg:"-n,required" help:"unique snippet name"`
	Kind string `arg:"-k,required" help:"snippet type ('disk' or 'post')"`
}

type snippetEditCmd struct {
	Contents string `arg:"positional" help:"replace snippet contents with file (or edit via $EDITOR)" placeholder:"FILE"`
	Name     string `arg:"-n,required"`
}

type snippetListCmd struct {
	Limit  int64 `arg:"-m" default:"100"`
	Offset int64 `arg:"-o" default:"0"`
}

type snippetDeleteCmd struct {
	Name string `arg:"-n,required"`
}

type snippetCmd struct {
	Create *snippetCreateCmd `arg:"subcommand:create" help:"create snippet"`
	Edit   *snippetEditCmd   `arg:"subcommand:edit" help:"edit snippet"`
	List   *snippetListCmd   `arg:"subcommand:list" help:"list snippets"`
	Delete *snippetDeleteCmd `arg:"subcommand:delete" help:"delete snippet"`
}

func snippetKindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "disk":
		return 1
	case "post":
		return 2
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func snippetIntToKind(kind int16) string {
	switch kind {
	case 1:
		return "disk"
	case 2:
		return "post"
	default:
		panic(fmt.Sprintf("unknown kind: %d", kind))
	}
}

var snippetTemplate = `# Edit this file and save and quit when done. Use Anaconda Kickstart syntax.

# Uncomment the following example, if you want to create 'disk' snippet:
#zerombr
#bootloader --location=mbr --timeout=1
#clearpart --all --initlabel
#autopart

# Uncomment the following example, if you want to create 'post' snippet:
#%post
#echo "HELLO WORLD"
#%end
`

func snippetCreate(ctx context.Context, cmdArgs *snippetCreateCmd) error {
	client := ctl.NewSnippetServiceClient(args.URL, http.DefaultClient)
	kind := snippetKindToInt(cmdArgs.Kind)

	session := edit.Session{Input: snippetTemplate}
	err := session.Edit()
	if err != nil {
		return fmt.Errorf("snippet edit error: %w", err)
	}

	err = client.Create(ctx, cmdArgs.Name, kind, session.Output)
	if err != nil {
		return fmt.Errorf("cannot create snippet: %w", err)
	}

	return nil
}

func snippetEdit(ctx context.Context, cmdArgs *snippetEditCmd) error {
	client := ctl.NewSnippetServiceClient(args.URL, http.DefaultClient)

	snippet, err := client.Find(ctx, cmdArgs.Name)
	if err != nil {
		return fmt.Errorf("cannot find snippet: %w", err)
	}

	session := edit.Session{Input: snippet.Body}
	err = session.Edit()
	if err != nil {
		return fmt.Errorf("snippet edit error: %w", err)
	}

	err = client.Edit(ctx, cmdArgs.Name, session.Output)
	if err != nil {
		return fmt.Errorf("cannot edit snippet: %w", err)
	}

	return nil
}

func snippetList(ctx context.Context, cmdArgs *snippetListCmd) error {
	client := ctl.NewSnippetServiceClient(args.URL, http.DefaultClient)

	snippets, err := client.List(ctx, cmdArgs.Limit, cmdArgs.Offset)
	if err != nil {
		return fmt.Errorf("cannot list snippets: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "ID\tName\tKind")
	for _, a := range snippets {
		fmt.Fprintf(w, "%d\t%s\t%s\n", a.ID, a.Name, snippetIntToKind(a.Kind))
	}
	w.Flush()

	return nil
}

func snippetDelete(ctx context.Context, cmdArgs *snippetDeleteCmd) error {
	client := ctl.NewSnippetServiceClient(args.URL, http.DefaultClient)

	err := client.Delete(ctx, cmdArgs.Name)
	if err != nil {
		return fmt.Errorf("cannot delete snippet: %w", err)
	}
	return nil
}
