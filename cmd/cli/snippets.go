package main

import (
	"context"
	"fmt"
	"forester/cmd/cli/edit"
	"forester/internal/api/ctl"
	"io"
	"net/http"
	"os"
)

type snippetCreateCmd struct {
	Name  string `arg:"-n,required" help:"unique snippet name"`
	Kind  string `arg:"-k,required" help:"snippet type ('disk' or 'post')"`
	Stdin bool   `arg:"-i" help:"snippet contents from stdin"`
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
	kind := ctl.SnippetKindToInt(cmdArgs.Kind)

	var contents string
	if cmdArgs.Stdin {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("cannot read snippet contents from stdin: %w", err)
		}
		contents = string(stdin)
	} else {
		session := edit.Session{Input: snippetTemplate}
		err := session.Edit()
		if err != nil {
			return fmt.Errorf("snippet edit error: %w", err)
		}
		contents = session.Output
	}

	err := client.Create(ctx, cmdArgs.Name, kind, contents)
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
		fmt.Fprintf(w, "%d\t%s\t%s\n", a.ID, a.Name, ctl.SnippetIntToKind(a.Kind))
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
