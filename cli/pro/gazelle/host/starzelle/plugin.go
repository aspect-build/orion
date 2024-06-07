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
	"github.com/bazelbuild/bazel-gazelle/rule"
	"go.starlark.net/starlark"
)

var proxyStateKey = "$starzelleState$"

var EmptyPrepareResult = plugin.PrepareResult{
	Sources: make([]plugin.SourceFilter, 0),
	Queries: plugin.NamedQueries{},
}

var EmptyDeclareTargetsResult = plugin.DeclareTargetsResult{}

// A go interface into a single starzelle plugin file.
// A starzelle file may include multiple plugins, hooks etc.
type StarzelleProxy interface {
	Plugins() []plugin.Plugin
	Kinds() []plugin.RuleKind
}

var _ StarzelleProxy = (*starzelleState)(nil)

type starzelleState struct {
	pluginPath string
	plugins    []plugin.Plugin
	kinds      []plugin.RuleKind
}

// Plugins implements StarzelleProxy.
func (s *starzelleState) Plugins() []plugin.Plugin {
	return s.plugins
}
func (s *starzelleState) Kinds() []plugin.RuleKind {
	return s.kinds
}

func LoadProxy(pluginPath string) (StarzelleProxy, error) {
	BazelLog.Infof("Load configure plugin %q", pluginPath)

	state := starzelleState{
		pluginPath: pluginPath,
		plugins:    make([]plugin.Plugin, 0),
		kinds:      make([]plugin.RuleKind, 0),
	}
	evalState := make(map[string]interface{})
	evalState[proxyStateKey] = &state

	libs := starlark.StringDict{
		"starzelle": starzelleModule,
	}

	_, err := starEval.Eval(pluginPath, libs, evalState)
	if err != nil {
		BazelLog.Errorf("Failed to load configure plugin %q: %v", pluginPath, err)
		fmt.Printf("Failed to load configure plugin %q: %v", pluginPath, err)
		return nil, err
	}

	return &state, nil
}

func (s *starzelleState) AddKind(name starlark.String, attributes *starlark.Dict) {
	s.kinds = append(s.kinds, readRuleKind(name, attributes))
}

func (s *starzelleState) AddPlugin(pluginId starlark.String, properties *starlark.Dict, prepare, analyze, declare *starlark.Function) {
	var pluginProperties map[string]plugin.Property

	if properties != nil {
		pluginProperties = starUtils.ReadMap(properties, readProperty)
	}

	s.plugins = append(s.plugins, starzellePluginProxy{
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

var _ plugin.Plugin = (*starzellePluginProxy)(nil)

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

	BazelLog.Debugf("Invoked plugin %s:prepare(): %v\n", p.name, v)
	return v.(plugin.PrepareResult)
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

	BazelLog.Debugf("Invoked plugin %s:DeclareTargets(): %v\n", p.name, v)
	return readDeclareTargetsResult(v)
}

func readDeclareTargetsResult(_ starlark.Value) plugin.DeclareTargetsResult {
	return plugin.DeclareTargetsResult{}
}

func readRuleKind(n starlark.String, v starlark.Value) plugin.RuleKind {
	return plugin.RuleKind{
		Name: n.GoString(),
		From: starUtils.ReadMapStringEntry(v, "From"),
		KindInfo: rule.KindInfo{
			MatchAny:        starUtils.ReadOptionalMapEntry(v, "MatchAny", starUtils.ReadBool, false),
			MatchAttrs:      starUtils.ReadOptionalMapEntry(v, "MatchAttrs", starUtils.ReadStringList, starUtils.EmptyStrings),
			NonEmptyAttrs:   starUtils.ReadOptionalMapEntry(v, "NonEmptyAttrs", starUtils.ReadBoolMap, starUtils.EmptyStringBoolMap),
			SubstituteAttrs: starUtils.ReadOptionalMapEntry(v, "SubstituteAttrs", starUtils.ReadBoolMap, starUtils.EmptyStringBoolMap),
			MergeableAttrs:  starUtils.ReadOptionalMapEntry(v, "MergeableAttrs", starUtils.ReadBoolMap, starUtils.EmptyStringBoolMap),
			ResolveAttrs:    starUtils.ReadOptionalMapEntry(v, "ResolveAttrs", starUtils.ReadBoolMap, starUtils.EmptyStringBoolMap),
		},
	}
}

func readProperty(k string, v starlark.Value) plugin.Property {
	return v.(plugin.Property)
}
