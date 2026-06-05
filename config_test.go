package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvForExecutableLoadsSkillDirectoryEnv(t *testing.T) {
	unsetEnv(t, "WECOM_GATEWAY_BASE_URL")
	unsetEnv(t, "CLI_IDENTITY_FILE")

	tempDir := t.TempDir()
	skillDir := filepath.Join(tempDir, "wecom-agw-operations")
	exeDir := filepath.Join(skillDir, "scripts", "windows")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatalf("mkdir executable dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: wecom-agw-operations\n---\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "identity.env"), []byte(`{"ACCESS_TOKEN":"skill-token"}`), 0o600); err != nil {
		t.Fatalf("write identity.env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, ".env"), []byte("WECOM_GATEWAY_BASE_URL=https://skill-env.example.test\nCLI_IDENTITY_FILE=identity.env\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	otherDir := filepath.Join(tempDir, "other")
	if err := os.Mkdir(otherDir, 0o755); err != nil {
		t.Fatalf("mkdir other dir: %v", err)
	}
	t.Chdir(otherDir)

	if err := loadDotEnvForExecutable(filepath.Join(exeDir, "wecom-cli.exe")); err != nil {
		t.Fatalf("loadDotEnvForExecutable returned error: %v", err)
	}
	if got, want := os.Getenv("WECOM_GATEWAY_BASE_URL"), "https://skill-env.example.test"; got != want {
		t.Fatalf("WECOM_GATEWAY_BASE_URL = %q, want %q", got, want)
	}
	if got, want := os.Getenv("CLI_IDENTITY_FILE"), filepath.Join(skillDir, "identity.env"); got != want {
		t.Fatalf("CLI_IDENTITY_FILE = %q, want %q", got, want)
	}
}

func TestParseGlobalFlagsReadsAccessTokenFromIdentityFile(t *testing.T) {
	unsetEnv(t, "WECOM_GATEWAY_BASE_URL")
	unsetEnv(t, "CLI_IDENTITY_FILE")

	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "identity.env")
	if err := os.WriteFile(identityFile, []byte(`{"ACCESS_TOKEN":"identity-token"}`), 0o600); err != nil {
		t.Fatalf("write identity file: %v", err)
	}
	t.Setenv("WECOM_GATEWAY_BASE_URL", "https://relay.example.test")
	t.Setenv("CLI_IDENTITY_FILE", identityFile)

	cfg, _, err := parseGlobalFlags([]string{"calendar", "create"})
	if err != nil {
		t.Fatalf("parseGlobalFlags returned error: %v", err)
	}
	client := &wecomClient{cfg: cfg}
	if err := client.requireCredentials(); err != nil {
		t.Fatalf("requireCredentials returned error: %v", err)
	}
	if client.cfg.GatewayToken != "identity-token" {
		t.Fatalf("GatewayToken = %q, want identity token", client.cfg.GatewayToken)
	}
}

func TestParseGlobalFlagsPrefersSystemIdentityFileOverDotEnv(t *testing.T) {
	unsetEnv(t, "WECOM_GATEWAY_BASE_URL")
	unsetEnv(t, "CLI_IDENTITY_FILE")

	tempDir := t.TempDir()
	dotEnvIdentity := filepath.Join(tempDir, "dotenv-identity.env")
	systemIdentity := filepath.Join(tempDir, "system-identity.env")
	if err := os.WriteFile(dotEnvIdentity, []byte(`{"ACCESS_TOKEN":"dotenv-token"}`), 0o600); err != nil {
		t.Fatalf("write dotenv identity file: %v", err)
	}
	if err := os.WriteFile(systemIdentity, []byte(`{"ACCESS_TOKEN":"system-token"}`), 0o600); err != nil {
		t.Fatalf("write system identity file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte("WECOM_GATEWAY_BASE_URL=https://relay.example.test\nCLI_IDENTITY_FILE="+dotEnvIdentity+"\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Chdir(tempDir)
	t.Setenv("CLI_IDENTITY_FILE", systemIdentity)

	cfg, _, err := parseGlobalFlags([]string{"calendar", "create"})
	if err != nil {
		t.Fatalf("parseGlobalFlags returned error: %v", err)
	}
	if cfg.IdentityFile != systemIdentity {
		t.Fatalf("IdentityFile = %q, want system variable path", cfg.IdentityFile)
	}
	client := &wecomClient{cfg: cfg}
	if err := client.requireCredentials(); err != nil {
		t.Fatalf("requireCredentials returned error: %v", err)
	}
	if client.cfg.GatewayToken != "system-token" {
		t.Fatalf("GatewayToken = %q, want system token", client.cfg.GatewayToken)
	}
}

func TestParseGlobalFlagsFallsBackToDotEnvWhenSystemIdentityFileBlank(t *testing.T) {
	unsetEnv(t, "WECOM_GATEWAY_BASE_URL")
	unsetEnv(t, "CLI_IDENTITY_FILE")

	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "dotenv-identity.env")
	if err := os.WriteFile(identityFile, []byte(`{"ACCESS_TOKEN":"dotenv-token"}`), 0o600); err != nil {
		t.Fatalf("write identity file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte("WECOM_GATEWAY_BASE_URL=https://relay.example.test\nCLI_IDENTITY_FILE="+identityFile+"\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Chdir(tempDir)
	t.Setenv("CLI_IDENTITY_FILE", " ")

	cfg, _, err := parseGlobalFlags([]string{"calendar", "create"})
	if err != nil {
		t.Fatalf("parseGlobalFlags returned error: %v", err)
	}
	if cfg.IdentityFile != identityFile {
		t.Fatalf("IdentityFile = %q, want .env fallback path", cfg.IdentityFile)
	}
}

func TestHelpDoesNotReadIdentityFile(t *testing.T) {
	t.Setenv("WECOM_GATEWAY_BASE_URL", "")
	t.Setenv("CLI_IDENTITY_FILE", filepath.Join(t.TempDir(), "missing.env"))

	if err := run([]string{"help"}); err != nil {
		t.Fatalf("run help returned error: %v", err)
	}
}

func TestGlobalFlagsRejectRemovedLocalStoreFlag(t *testing.T) {
	_, _, err := parseGlobalFlags([]string{"--resource-table", "resources.json", "calendar", "help"})
	if err == nil {
		t.Fatal("expected --resource-table to be unsupported")
	}
}

func TestResourcesHelpDoesNotReadIdentityFile(t *testing.T) {
	t.Setenv("WECOM_GATEWAY_BASE_URL", "")
	t.Setenv("CLI_IDENTITY_FILE", filepath.Join(t.TempDir(), "missing.env"))

	if err := run([]string{"resources", "help"}); err != nil {
		t.Fatalf("run resources help returned error: %v", err)
	}
}

func TestParseGlobalFlagsDerivesAGWBaseURLFromWeComGateway(t *testing.T) {
	unsetEnv(t, "AGW_GATEWAY_BASE_URL")
	unsetEnv(t, "WECOM_AGW_BASE_URL")
	t.Setenv("WECOM_GATEWAY_BASE_URL", "https://gateway.example.test/wecom")

	cfg, _, err := parseGlobalFlags([]string{"resources", "list"})
	if err != nil {
		t.Fatalf("parseGlobalFlags returned error: %v", err)
	}
	if got, want := cfg.AGWBaseURL, "https://gateway.example.test"; got != want {
		t.Fatalf("AGWBaseURL = %q, want %q", got, want)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, hadOld := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if hadOld {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
