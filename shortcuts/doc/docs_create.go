// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"

	"github.com/larksuite/cli/shortcuts/common"
)

var DocsCreate = common.Shortcut{
	Service:     "docs",
	Command:     "+create",
	Description: "Create a Lark document",
	Risk:        "write",
	AuthTypes:   []string{"user", "bot"},
	Scopes:      []string{"docx:document:create"},
	Flags: []common.Flag{
		{Name: "content", Desc: "document content (XML or Markdown)", Required: true, Input: []string{common.File, common.Stdin}},
		{Name: "doc-format", Desc: "content format（prefer XML）", Default: "xml", Enum: []string{"xml", "markdown"}},
		{Name: "parent-token", Desc: "parent folder or wiki-node token"},
		{Name: "parent-position", Desc: "parent position (e.g. my_library)"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if runtime.Str("parent-token") != "" && runtime.Str("parent-position") != "" {
			return common.FlagErrorf("--parent-token and --parent-position are mutually exclusive")
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		body := buildCreateBody(runtime)
		return common.NewDryRunAPI().
			POST("/open-apis/docs_ai/v1/documents").
			Desc("OpenAPI: create document").
			Body(body)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		body := buildCreateBody(runtime)

		data, err := doDocAPI(runtime, "POST", "/open-apis/docs_ai/v1/documents", body)
		if err != nil {
			return err
		}

		stripBlockIDs(data)
		runtime.OutRaw(data, nil)
		return nil
	},
}

func buildCreateBody(runtime *common.RuntimeContext) map[string]interface{} {
	body := map[string]interface{}{
		"format":  runtime.Str("doc-format"),
		"content": runtime.Str("content"),
	}
	if v := runtime.Str("parent-token"); v != "" {
		body["parent_token"] = v
	}
	if v := runtime.Str("parent-position"); v != "" {
		body["parent_position"] = v
	}
	return body
}
