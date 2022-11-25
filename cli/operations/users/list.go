// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package users

import (
	"context"
	"encoding/json"
	"os"
	"text/template"
	"time"

	"github.com/harness/gitness/client"
	"github.com/harness/gitness/types"

	"github.com/drone/funcmap"
	"gopkg.in/alecthomas/kingpin.v2"
)

const userTmpl = `
id:    {{ .ID }}
email: {{ .Email }}
admin: {{ .Admin }}
`

type listCommand struct {
	client client.Client
	tmpl   string
	page   int
	size   int
	json   bool
}

func (c *listCommand) run(*kingpin.ParseContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	list, err := c.client.UserList(ctx, types.Params{
		Size: c.size,
		Page: c.page,
	})
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Funcs(funcmap.Funcs).Parse(c.tmpl + "\n")
	if err != nil {
		return err
	}
	if c.json {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	}
	for _, item := range list {
		if err = tmpl.Execute(os.Stdout, item); err != nil {
			return err
		}
	}
	return nil
}

// helper function registers the user list command.
func registerList(app *kingpin.CmdClause, client client.Client) {
	c := &listCommand{
		client: client,
	}

	cmd := app.Command("ls", "display a list of users").
		Action(c.run)

	cmd.Flag("page", "page number").
		IntVar(&c.page)

	cmd.Flag("per-page", "page size").
		IntVar(&c.size)

	cmd.Flag("json", "json encode the output").
		BoolVar(&c.json)

	cmd.Flag("format", "format the output using a Go template").
		Default(userTmpl).
		Hidden().
		StringVar(&c.tmpl)
}