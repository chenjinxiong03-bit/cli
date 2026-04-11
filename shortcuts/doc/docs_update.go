// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"fmt"

	"github.com/larksuite/cli/shortcuts/common"
)

var validCommands = []string{
	"str_replace",
	"str_delete",
	"block_delete",
	"block_insert_after",
	"block_replace",
	"overwrite",
	"append",
}

var DocsUpdate = common.Shortcut{
	Service:     "docs",
	Command:     "+update",
	Description: "Update a Lark document",
	Risk:        "write",
	Scopes:      []string{"docx:document:write_only", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "command", Desc: "operation: str_replace | str_delete | block_delete | block_insert_after | block_replace | overwrite | append", Required: true, Enum: validCommands},
		{Name: "doc-format", Desc: "content format（prefer XML）", Default: "xml", Enum: []string{"xml", "markdown"}},
		{Name: "content", Desc: "new content (XML or Markdown)", Input: []string{common.File, common.Stdin}},
		{Name: "pattern", Desc: "regex pattern for str_replace / str_delete"},
		{Name: "block-id", Desc: "target block ID for block_* operations"},
		{Name: "revision-id", Desc: "base revision (-1 = latest)", Type: "int", Default: "-1"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		cmd := runtime.Str("command")
		content := runtime.Str("content")
		pattern := runtime.Str("pattern")
		blockID := runtime.Str("block-id")

		switch cmd {
		case "str_replace":
			if pattern == "" {
				return common.FlagErrorf("--command str_replace requires --pattern")
			}
			if content == "" {
				return common.FlagErrorf("--command str_replace requires --content")
			}
		case "str_delete":
			if pattern == "" {
				return common.FlagErrorf("--command str_delete requires --pattern")
			}
		case "block_delete":
			if blockID == "" {
				return common.FlagErrorf("--command block_delete requires --block-id")
			}
		case "block_insert_after":
			if blockID == "" {
				return common.FlagErrorf("--command block_insert_after requires --block-id")
			}
			if content == "" {
				return common.FlagErrorf("--command block_insert_after requires --content")
			}
		case "block_replace":
			if blockID == "" {
				return common.FlagErrorf("--command block_replace requires --block-id")
			}
			if content == "" {
				return common.FlagErrorf("--command block_replace requires --content")
			}
		case "overwrite":
			if content == "" {
				return common.FlagErrorf("--command overwrite requires --content")
			}
		case "append":
			if content == "" {
				return common.FlagErrorf("--command append requires --content")
			}
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		ref, err := parseDocumentRef(runtime.Str("doc"))
		if err != nil {
			return common.NewDryRunAPI().Desc(fmt.Sprintf("error: %v", err))
		}
		body := buildUpdateBody(runtime)
		apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s", ref.Token)
		return common.NewDryRunAPI().
			PUT(apiPath).
			Desc("OpenAPI: update document").
			Body(body).
			Set("document_id", ref.Token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		ref, err := parseDocumentRef(runtime.Str("doc"))
		if err != nil {
			return err
		}

		apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s", ref.Token)
		body := buildUpdateBody(runtime)

		data, err := doDocAPI(runtime, "PUT", apiPath, body)
		if err != nil {
			return err
		}

		runtime.OutRaw(data, nil)
		return nil
	},
}

func buildUpdateBody(runtime *common.RuntimeContext) map[string]interface{} {
	cmd := runtime.Str("command")

	// append is a shorthand for block_insert_after with block_id "-1" (end of document)
	blockID := runtime.Str("block-id")
	if cmd == "append" {
		cmd = "block_insert_after"
		blockID = "-1"
	}

	body := map[string]interface{}{
		"format":  runtime.Str("doc-format"),
		"command": cmd,
	}
	if v := runtime.Int("revision-id"); v != 0 {
		body["revision_id"] = v
	}
	if v := runtime.Str("content"); v != "" {
		body["content"] = v
	}
	if v := runtime.Str("pattern"); v != "" {
		body["pattern"] = v
	}
	if blockID != "" {
		body["block_id"] = blockID
	}
	return body
}
