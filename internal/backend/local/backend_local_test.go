// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package local

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/opentffoundation/opentf/internal/backend"
	"github.com/opentffoundation/opentf/internal/command/arguments"
	"github.com/opentffoundation/opentf/internal/command/clistate"
	"github.com/opentffoundation/opentf/internal/command/views"
	"github.com/opentffoundation/opentf/internal/configs/configload"
	"github.com/opentffoundation/opentf/internal/configs/configschema"
	"github.com/opentffoundation/opentf/internal/initwd"
	"github.com/opentffoundation/opentf/internal/plans"
	"github.com/opentffoundation/opentf/internal/plans/planfile"
	"github.com/opentffoundation/opentf/internal/states"
	"github.com/opentffoundation/opentf/internal/states/statefile"
	"github.com/opentffoundation/opentf/internal/states/statemgr"
	"github.com/opentffoundation/opentf/internal/terminal"
	"github.com/opentffoundation/opentf/internal/terraform"
	"github.com/opentffoundation/opentf/internal/tfdiags"
)

func TestLocalRun(t *testing.T) {
	configDir := "./testdata/empty"
	b := TestLocal(t)

	_, configLoader, configCleanup := initwd.MustLoadConfigForTests(t, configDir)
	defer configCleanup()

	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)
	stateLocker := clistate.NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))

	op := &backend.Operation{
		ConfigDir:    configDir,
		ConfigLoader: configLoader,
		Workspace:    backend.DefaultStateName,
		StateLocker:  stateLocker,
	}

	_, _, diags := b.LocalRun(op)
	if diags.HasErrors() {
		t.Fatalf("unexpected error: %s", diags.Err().Error())
	}

	// LocalRun() retains a lock on success
	assertBackendStateLocked(t, b)
}

func TestLocalRun_error(t *testing.T) {
	configDir := "./testdata/invalid"
	b := TestLocal(t)

	// This backend will return an error when asked to RefreshState, which
	// should then cause LocalRun to return with the state unlocked.
	b.Backend = backendWithStateStorageThatFailsRefresh{}

	_, configLoader, configCleanup := initwd.MustLoadConfigForTests(t, configDir)
	defer configCleanup()

	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)
	stateLocker := clistate.NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))

	op := &backend.Operation{
		ConfigDir:    configDir,
		ConfigLoader: configLoader,
		Workspace:    backend.DefaultStateName,
		StateLocker:  stateLocker,
	}

	_, _, diags := b.LocalRun(op)
	if !diags.HasErrors() {
		t.Fatal("unexpected success")
	}

	// LocalRun() unlocks the state on failure
	assertBackendStateUnlocked(t, b)
}

func TestLocalRun_stalePlan(t *testing.T) {
	configDir := "./testdata/apply"
	b := TestLocal(t)

	_, configLoader, configCleanup := initwd.MustLoadConfigForTests(t, configDir)
	defer configCleanup()

	// Write an empty state file with serial 3
	sf, err := os.Create(b.StatePath)
	if err != nil {
		t.Fatalf("unexpected error creating state file %s: %s", b.StatePath, err)
	}
	if err := statefile.Write(statefile.New(states.NewState(), "boop", 3), sf); err != nil {
		t.Fatalf("unexpected error writing state file: %s", err)
	}

	// Refresh the state
	sm, err := b.StateMgr("")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := sm.RefreshState(); err != nil {
		t.Fatalf("unexpected error refreshing state: %s", err)
	}

	// Create a minimal plan which also has state file serial 2, so is stale
	backendConfig := cty.ObjectVal(map[string]cty.Value{
		"path":          cty.NullVal(cty.String),
		"workspace_dir": cty.NullVal(cty.String),
	})
	backendConfigRaw, err := plans.NewDynamicValue(backendConfig, backendConfig.Type())
	if err != nil {
		t.Fatal(err)
	}
	plan := &plans.Plan{
		UIMode:  plans.NormalMode,
		Changes: plans.NewChanges(),
		Backend: plans.Backend{
			Type:   "local",
			Config: backendConfigRaw,
		},
		PrevRunState: states.NewState(),
		PriorState:   states.NewState(),
	}
	prevStateFile := statefile.New(plan.PrevRunState, "boop", 1)
	stateFile := statefile.New(plan.PriorState, "boop", 2)

	// Roundtrip through serialization as expected by the operation
	outDir := t.TempDir()
	defer os.RemoveAll(outDir)
	planPath := filepath.Join(outDir, "plan.tfplan")
	planfileArgs := planfile.CreateArgs{
		ConfigSnapshot:       configload.NewEmptySnapshot(),
		PreviousRunStateFile: prevStateFile,
		StateFile:            stateFile,
		Plan:                 plan,
	}
	if err := planfile.Create(planPath, planfileArgs); err != nil {
		t.Fatalf("unexpected error writing planfile: %s", err)
	}
	planFile, err := planfile.Open(planPath)
	if err != nil {
		t.Fatalf("unexpected error reading planfile: %s", err)
	}

	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)
	stateLocker := clistate.NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))

	op := &backend.Operation{
		ConfigDir:    configDir,
		ConfigLoader: configLoader,
		PlanFile:     planFile,
		Workspace:    backend.DefaultStateName,
		StateLocker:  stateLocker,
	}

	_, _, diags := b.LocalRun(op)
	if !diags.HasErrors() {
		t.Fatal("unexpected success")
	}

	// LocalRun() unlocks the state on failure
	assertBackendStateUnlocked(t, b)
}

type backendWithStateStorageThatFailsRefresh struct {
}

var _ backend.Backend = backendWithStateStorageThatFailsRefresh{}

func (b backendWithStateStorageThatFailsRefresh) StateMgr(workspace string) (statemgr.Full, error) {
	return &stateStorageThatFailsRefresh{}, nil
}

func (b backendWithStateStorageThatFailsRefresh) ConfigSchema() *configschema.Block {
	return &configschema.Block{}
}

func (b backendWithStateStorageThatFailsRefresh) PrepareConfig(in cty.Value) (cty.Value, tfdiags.Diagnostics) {
	return in, nil
}

func (b backendWithStateStorageThatFailsRefresh) Configure(cty.Value) tfdiags.Diagnostics {
	return nil
}

func (b backendWithStateStorageThatFailsRefresh) DeleteWorkspace(name string, force bool) error {
	return fmt.Errorf("unimplemented")
}

func (b backendWithStateStorageThatFailsRefresh) Workspaces() ([]string, error) {
	return []string{"default"}, nil
}

type stateStorageThatFailsRefresh struct {
	locked bool
}

func (s *stateStorageThatFailsRefresh) Lock(info *statemgr.LockInfo) (string, error) {
	if s.locked {
		return "", fmt.Errorf("already locked")
	}
	s.locked = true
	return "locked", nil
}

func (s *stateStorageThatFailsRefresh) Unlock(id string) error {
	if !s.locked {
		return fmt.Errorf("not locked")
	}
	s.locked = false
	return nil
}

func (s *stateStorageThatFailsRefresh) State() *states.State {
	return nil
}

func (s *stateStorageThatFailsRefresh) GetRootOutputValues() (map[string]*states.OutputValue, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *stateStorageThatFailsRefresh) WriteState(*states.State) error {
	return fmt.Errorf("unimplemented")
}

func (s *stateStorageThatFailsRefresh) RefreshState() error {
	return fmt.Errorf("intentionally failing for testing purposes")
}

func (s *stateStorageThatFailsRefresh) PersistState(schemas *terraform.Schemas) error {
	return fmt.Errorf("unimplemented")
}
