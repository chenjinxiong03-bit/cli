// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/keychain"
	"github.com/larksuite/cli/internal/output"
)

type noopConfigKeychain struct{}

func (n *noopConfigKeychain) Get(service, account string) (string, error) { return "", nil }
func (n *noopConfigKeychain) Set(service, account, value string) error    { return nil }
func (n *noopConfigKeychain) Remove(service, account string) error        { return nil }

func TestConfigInitCmd_FlagParsing(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, nil)
	f.IOStreams.In = strings.NewReader("secret123\n")

	var gotOpts *ConfigInitOptions
	cmd := NewCmdConfigInit(f, func(opts *ConfigInitOptions) error {
		gotOpts = opts
		return nil
	})
	cmd.SetArgs([]string{"--app-id", "cli_test", "--app-secret-stdin", "--brand", "lark"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOpts.AppID != "cli_test" {
		t.Errorf("expected AppID cli_test, got %s", gotOpts.AppID)
	}
	if !gotOpts.AppSecretStdin {
		t.Error("expected AppSecretStdin=true")
	}
	if gotOpts.Brand != "lark" {
		t.Errorf("expected Brand lark, got %s", gotOpts.Brand)
	}
}

func TestConfigShowCmd_FlagParsing(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{
		AppID: "test-app", AppSecret: "test-secret", Brand: core.BrandFeishu,
	})

	var gotOpts *ConfigShowOptions
	cmd := NewCmdConfigShow(f, func(opts *ConfigShowOptions) error {
		gotOpts = opts
		return nil
	})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOpts == nil {
		t.Error("expected opts to be set")
	}
}

func TestConfigShowRun_NotConfiguredReturnsStructuredError(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	f, _, _, _ := cmdutil.TestFactory(t, nil)
	err := configShowRun(&ConfigShowOptions{Factory: f})
	if err == nil {
		t.Fatal("expected error")
	}

	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error type = %T, want *output.ExitError", err)
	}
	if exitErr.Code != output.ExitValidation {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, output.ExitValidation)
	}
	if exitErr.Detail == nil || exitErr.Detail.Type != "config" || exitErr.Detail.Message != "not configured" {
		t.Fatalf("detail = %#v, want config/not configured", exitErr.Detail)
	}
}

func TestConfigShowRun_NoActiveProfileReturnsStructuredError(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	multi := &core.MultiAppConfig{
		CurrentApp: "missing",
		Apps: []core.AppConfig{{
			Name:      "default",
			AppId:     "app-default",
			AppSecret: core.PlainSecret("secret-default"),
			Brand:     core.BrandFeishu,
		}},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatalf("SaveMultiAppConfig() error = %v", err)
	}

	f, _, _, _ := cmdutil.TestFactory(t, nil)
	err := configShowRun(&ConfigShowOptions{Factory: f})
	if err == nil {
		t.Fatal("expected error")
	}

	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error type = %T, want *output.ExitError", err)
	}
	if exitErr.Code != output.ExitValidation {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, output.ExitValidation)
	}
	if exitErr.Detail == nil || exitErr.Detail.Type != "config" || exitErr.Detail.Message != "no active profile" {
		t.Fatalf("detail = %#v, want config/no active profile", exitErr.Detail)
	}
}

func TestConfigInitCmd_LangFlag(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, nil)

	var gotOpts *ConfigInitOptions
	cmd := NewCmdConfigInit(f, func(opts *ConfigInitOptions) error {
		gotOpts = opts
		return nil
	})
	f.IOStreams.In = strings.NewReader("y\n")
	cmd.SetArgs([]string{"--app-id", "x", "--app-secret-stdin", "--lang", "en"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOpts.Lang != "en" {
		t.Errorf("expected Lang en, got %s", gotOpts.Lang)
	}
	if !gotOpts.langExplicit {
		t.Error("expected langExplicit=true when --lang is passed")
	}
}

func TestConfigInitCmd_LangDefault(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, nil)

	var gotOpts *ConfigInitOptions
	cmd := NewCmdConfigInit(f, func(opts *ConfigInitOptions) error {
		gotOpts = opts
		return nil
	})
	f.IOStreams.In = strings.NewReader("y\n")
	cmd.SetArgs([]string{"--app-id", "x", "--app-secret-stdin"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOpts.Lang != "zh" {
		t.Errorf("expected default Lang zh, got %s", gotOpts.Lang)
	}
	if gotOpts.langExplicit {
		t.Error("expected langExplicit=false when --lang is not passed")
	}
}

func TestHasAnyNonInteractiveFlag(t *testing.T) {
	tests := []struct {
		name string
		opts ConfigInitOptions
		want bool
	}{
		{"empty", ConfigInitOptions{}, false},
		{"new", ConfigInitOptions{New: true}, true},
		{"app-id", ConfigInitOptions{AppID: "x"}, true},
		{"app-secret-stdin", ConfigInitOptions{AppSecretStdin: true}, true},
		{"app-id+secret-stdin", ConfigInitOptions{AppID: "x", AppSecretStdin: true}, true},
		{"lang-only", ConfigInitOptions{Lang: "en"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.hasAnyNonInteractiveFlag()
			if got != tt.want {
				t.Errorf("hasAnyNonInteractiveFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigInitRun_NonTerminal_NoFlags_RejectsWithHint(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, nil)
	// TestFactory has IsTerminal=false by default
	opts := &ConfigInitOptions{Factory: f, Ctx: context.Background(), Lang: "zh"}
	err := configInitRun(opts)
	if err == nil {
		t.Fatal("expected error for non-terminal without flags")
	}
	msg := err.Error()
	if !strings.Contains(msg, "--new") {
		t.Errorf("expected error to mention --new, got: %s", msg)
	}
	if !strings.Contains(msg, "terminal") {
		t.Errorf("expected error to mention terminal, got: %s", msg)
	}
}

func TestConfigRemoveCmd_FlagParsing(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, nil)

	var gotOpts *ConfigRemoveOptions
	cmd := NewCmdConfigRemove(f, func(opts *ConfigRemoveOptions) error {
		gotOpts = opts
		return nil
	})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOpts == nil {
		t.Fatal("expected opts to be set")
	}
	if gotOpts.Factory != f {
		t.Fatal("expected factory to be preserved in options")
	}
}

func TestConfigRemoveRun_NotConfiguredReturnsValidationError(t *testing.T) {
	// GIVEN: no config file exists
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	f, _, _, _ := cmdutil.TestFactory(t, nil)

	// WHEN: configRemoveRun is called
	err := configRemoveRun(&ConfigRemoveOptions{Factory: f})

	// THEN: returns a validation error "not configured yet"
	if err == nil {
		t.Fatal("expected error when not configured")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("error = %v, want 'not configured'", err)
	}
}

type recordingConfigKeychain struct {
	removed []string
}

func (r *recordingConfigKeychain) Get(service, account string) (string, error) { return "", nil }
func (r *recordingConfigKeychain) Set(service, account, value string) error    { return nil }
func (r *recordingConfigKeychain) Remove(service, account string) error {
	r.removed = append(r.removed, service+":"+account)
	return nil
}

func TestConfigRemoveRun_CleansKeychainBeforeSave(t *testing.T) {
	// GIVEN: a config with a single app that has a keychain secret
	configDir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", configDir)

	multi := &core.MultiAppConfig{
		Apps: []core.AppConfig{{
			AppId: "app-test",
			AppSecret: core.SecretInput{
				Ref: &core.SecretRef{Source: "keychain", ID: "appsecret:app-test"},
			},
			Brand: core.BrandFeishu,
		}},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatalf("SaveMultiAppConfig() error = %v", err)
	}

	kc := &recordingConfigKeychain{}
	f, _, stderr, _ := cmdutil.TestFactory(t, nil)
	f.Keychain = kc

	// WHEN: configRemoveRun is called
	err := configRemoveRun(&ConfigRemoveOptions{Factory: f})

	// THEN: no error
	if err != nil {
		t.Fatalf("configRemoveRun() error = %v", err)
	}

	// THEN: keychain entries were cleaned (verifies cleanup happens)
	if len(kc.removed) == 0 {
		t.Error("expected keychain entries to be removed")
	}

	// THEN: success message printed
	if !strings.Contains(stderr.String(), "Configuration removed") {
		t.Errorf("expected success message in stderr, got: %s", stderr.String())
	}
}

func TestConfigRemoveRun_SavesEmptyConfigAfterCleanup(t *testing.T) {
	// GIVEN: a config with apps and users
	configDir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", configDir)

	multi := &core.MultiAppConfig{
		Apps: []core.AppConfig{
			{
				AppId:     "app1",
				AppSecret: core.PlainSecret("secret1"),
				Brand:     core.BrandFeishu,
				Users:     []core.AppUser{{UserOpenId: "ou_user1", UserName: "User1"}},
			},
			{
				AppId:     "app2",
				AppSecret: core.PlainSecret("secret2"),
				Brand:     core.BrandFeishu,
			},
		},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatalf("SaveMultiAppConfig() error = %v", err)
	}

	f, _, stderr, _ := cmdutil.TestFactory(t, nil)
	f.Keychain = &noopConfigKeychain{}

	// WHEN: configRemoveRun is called
	err := configRemoveRun(&ConfigRemoveOptions{Factory: f})

	// THEN: no error
	if err != nil {
		t.Fatalf("configRemoveRun() error = %v", err)
	}

	// THEN: config is now empty (LoadMultiAppConfig returns "no apps" error for empty config)
	saved, loadErr := core.LoadMultiAppConfig()
	if loadErr == nil && saved != nil && len(saved.Apps) > 0 {
		t.Fatalf("expected empty config after remove, got %d apps", len(saved.Apps))
	}
	// Either a "no apps" error or nil saved config is acceptable - both indicate successful removal
	if loadErr != nil && !strings.Contains(loadErr.Error(), "no apps") {
		t.Fatalf("unexpected LoadMultiAppConfig() error = %v", loadErr)
	}

	// THEN: stderr mentions user count from the cleared apps
	if !strings.Contains(stderr.String(), "1 users") {
		t.Errorf("expected '1 users' message (1 user from app1), got: %s", stderr.String())
	}
}

func TestSaveAsProfile_RejectsProfileNameCollisionWithExistingAppID(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	existing := &core.MultiAppConfig{
		Apps: []core.AppConfig{
			{
				Name:      "prod",
				AppId:     "cli_prod",
				AppSecret: core.PlainSecret("secret"),
				Brand:     core.BrandFeishu,
			},
		},
	}

	err := saveAsProfile(existing, keychain.KeychainAccess(&noopConfigKeychain{}), "cli_prod", "app-new", core.PlainSecret("new-secret"), core.BrandLark, "en")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "conflicts with existing appId") {
		t.Fatalf("error = %v, want conflict with existing appId", err)
	}
}

func TestUpdateExistingProfileWithoutSecret_RejectsAppIDChange(t *testing.T) {
	multi := &core.MultiAppConfig{
		CurrentApp: "prod",
		Apps: []core.AppConfig{
			{
				Name:      "prod",
				AppId:     "app-old",
				AppSecret: core.SecretInput{Ref: &core.SecretRef{Source: "keychain", ID: "appsecret:app-old"}},
				Brand:     core.BrandFeishu,
				Lang:      "zh",
				Users:     []core.AppUser{{UserOpenId: "ou_1", UserName: "User"}},
			},
		},
	}

	err := updateExistingProfileWithoutSecret(multi, "", "app-new", core.BrandLark, "en")
	if err == nil {
		t.Fatal("expected error when changing app ID without a new secret")
	}
	if !strings.Contains(err.Error(), "App Secret") {
		t.Fatalf("error = %v, want mention of App Secret", err)
	}
}