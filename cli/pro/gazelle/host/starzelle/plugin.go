package starzelle

/**
 * A proxy into a starzelle plugin file.
 */

import (
	"fmt"

	starEval "github.com/aspect-build/silo/cli/core/gazelle/common/starlark"
	starUtils "github.com/aspect-build/silo/cli/core/gazelle/common/starlark/utils"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"go.starlark.net/starlark"
)

var proxyStateKey = "$starzelleState$"

var EmptyPrepareResult = plugin.PrepareResult{
	Sources: make([]plugin.SourceFilter, 0),
	Queries: plugin.NamedQueries{},
}

var EmptyDeclareTargetsResult = plugin.DeclareTargetsResult{}

type starzelleState struct {
	pluginPath string
	host       plugin.PluginHost
}

func LoadProxy(host plugin.PluginHost, pluginPath string) error {
	BazelLog.Infof("Load configure plugin %q", pluginPath)

	state := starzelleState{
		pluginPath: pluginPath,
		host:       host,
	}
	evalState := make(map[string]interface{})
	evalState[proxyStateKey] = &state

	libs := starlark.StringDict{
		"aspect": aspectModule,
	}

	_, err := starEval.Eval(pluginPath, libs, evalState)
	if err != nil {
		BazelLog.Errorf("Failed to load configure plugin %q: %v", pluginPath, err)
		fmt.Printf("Failed to load configure plugin %q: %v", pluginPath, err)
		return err
	}

	return nil
}

func (s *starzelleState) AddKind(name starlark.String, attributes *starlark.Dict) {
	s.host.AddKind(readRuleKind(name, attributes))
}

func (s *starzelleState) AddPlugin(pluginId starlark.String, properties *starlark.Dict, prepare, analyze, declare *starlark.Function) {
	var pluginProperties map[string]plugin.Property

	if properties != nil {
		pluginProperties = starUtils.ReadMap(properties, readProperty)
	}

	s.host.AddPlugin(starzellePluginProxy{
		name:       pluginId.GoString(),
		pluginPath: s.pluginPath,
		properties: pluginProperties,
		prepare:    prepare,
		analyze:    analyze,
		declare:    declare,
	})
}

// A plugin implementation loaded via starlark and proxying
// to starlark functions.
var _ plugin.Plugin = (*starzellePluginProxy)(nil)

type starzellePluginProxy struct {
	name                      string
	pluginPath                string
	properties                map[string]plugin.Property
	prepare, analyze, declare *starlark.Function
}

func (p starzellePluginProxy) Name() string {
	return p.name
}

func (p starzellePluginProxy) Properties() map[string]plugin.Property {
	return p.properties
}

func (p starzellePluginProxy) Prepare(ctx plugin.PrepareContext) plugin.PrepareResult {
	if p.prepare == nil {
		return EmptyPrepareResult
	}

	v, err := starEval.Call(p.prepare, starlark.Tuple{ctx}, starUtils.EmptyKwArgs)
	if err != nil {
		errStr := starUtils.ErrorStr(fmt.Sprintf("Failed to invoke %s:Prepare()", p.name), err)
		BazelLog.Errorf(errStr)
		fmt.Printf(errStr)
		return EmptyPrepareResult
	}

	BazelLog.Debugf("Invoked plugin %s:prepare(%q): %v\n", p.name, ctx.Rel, v)

	pr, isPR := v.(plugin.PrepareResult)
	if !isPR {
		BazelLog.Fatalf("Prepare %v is not a PrepareResult", v)
	}

	return pr
}

// Analyze implements plugin.Plugin.
func (p starzellePluginProxy) Analyze(ctx plugin.AnalyzeContext) error {
	if p.analyze == nil {
		return nil
	}
	_, err := starEval.Call(p.analyze, starlark.Tuple{&ctx}, starUtils.EmptyKwArgs)
	if err != nil {
		return err
	}
	return nil
}

func (p starzellePluginProxy) DeclareTargets(ctx plugin.DeclareTargetsContext) plugin.DeclareTargetsResult {
	if p.declare == nil {
		return EmptyDeclareTargetsResult
	}

	v, err := starEval.Call(p.declare, starlark.Tuple{ctx}, starUtils.EmptyKwArgs)
	if err != nil {
		errStr := starUtils.ErrorStr(fmt.Sprintf("Failed to invoke %s:DeclareTargets()", p.name), err)
		BazelLog.Errorf(errStr)
		fmt.Printf(errStr)
		return EmptyDeclareTargetsResult
	}

	BazelLog.Debugf("Invoked plugin %s:DeclareTargets(%q): %v\n", p.name, ctx.Rel, v)
	return readDeclareTargetsResult(v)
}

func readDeclareTargetsResult(_ starlark.Value) plugin.DeclareTargetsResult {
	return plugin.DeclareTargetsResult{}
}

func readRuleKind(n starlark.String, v starlark.Value) plugin.RuleKind {
	return plugin.RuleKind{
		Name: n.GoString(),
		From: starUtils.ReadMapStringEntry(v, "From"),
		KindInfo: plugin.KindInfo{
			MatchAny:       starUtils.ReadOptionalMapEntry(v, "MatchAny", starUtils.ReadBool, false),
			MatchAttrs:     starUtils.ReadOptionalMapEntry(v, "MatchAttrs", starUtils.ReadStringList, starUtils.EmptyStrings),
			NonEmptyAttrs:  starUtils.ReadOptionalMapEntry(v, "NonEmptyAttrs", starUtils.ReadStringList, starUtils.EmptyStrings),
			MergeableAttrs: starUtils.ReadOptionalMapEntry(v, "MergeableAttrs", starUtils.ReadStringList, starUtils.EmptyStrings),
			ResolveAttrs:   starUtils.ReadOptionalMapEntry(v, "ResolveAttrs", starUtils.ReadStringList, starUtils.EmptyStrings),
		},
	}
}

func readProperty(k string, v starlark.Value) plugin.Property {
	p, isProp := v.(plugin.Property)

	if !isProp {
		BazelLog.Fatalf("Property %v is not a Property", k)
	}

	if p.Name != "" && p.Name != k {
		BazelLog.Errorf("Property name %q does not match key %q", p.Name, k)
	}

	p.Name = k
	return p
}
