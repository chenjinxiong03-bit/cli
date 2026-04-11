// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"fmt"
	"io"

	"github.com/larksuite/cli/shortcuts/common"
)

var DocsFetch = common.Shortcut{
	Service:     "docs",
	Command:     "+fetch",
	Description: "Fetch Lark document content",
	Risk:        "read",
	Scopes:      []string{"docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "doc-format", Desc: "content format", Default: "xml", Enum: []string{"xml", "markdown", "text"}},
		{Name: "detail", Desc: "export detail level: simple (read-only) | with-ids (block IDs for cross-referencing) | full (all attrs for editing)", Default: "simple", Enum: []string{"simple", "with-ids", "full"}},
		{Name: "revision-id", Desc: "document revision (-1 = latest)", Type: "int", Default: "-1"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		ref, err := parseDocumentRef(runtime.Str("doc"))
		if err != nil {
			return common.NewDryRunAPI().Desc(fmt.Sprintf("error: %v", err))
		}
		body := buildFetchBody(runtime)
		apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s/fetch", ref.Token)
		return common.NewDryRunAPI().
			POST(apiPath).
			Desc("OpenAPI: fetch document").
			Body(body).
			Set("document_id", ref.Token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		ref, err := parseDocumentRef(runtime.Str("doc"))
		if err != nil {
			return err
		}

		apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s/fetch", ref.Token)
		body := buildFetchBody(runtime)

		data, err := doDocAPI(runtime, "POST", apiPath, body)
		if err != nil {
			return err
		}

		runtime.OutFormatRaw(data, nil, func(w io.Writer) {
			if doc, ok := data["document"].(map[string]interface{}); ok {
				if content, ok := doc["content"].(string); ok {
					fmt.Fprintln(w, content)
				}
			}
		})
		return nil
	},
}

func buildFetchBody(runtime *common.RuntimeContext) map[string]interface{} {
	body := map[string]interface{}{
		"format": runtime.Str("doc-format"),
	}
	if v := runtime.Int("revision-id"); v != 0 {
		body["revision_id"] = v
	}

	detail := runtime.Str("detail")
	switch detail {
	case "with-ids":
		body["export_option"] = map[string]interface{}{
			"export_block_id": true,
		}
	case "full":
		body["export_option"] = map[string]interface{}{
			"export_block_id":        true,
			"export_style_attrs":     true,
			"export_cite_extra_data": true,
		}
	}
	// "simple": no export_option (server defaults)

	return body
}
