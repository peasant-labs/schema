package intent

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// configBotKinds is the closed set of keys a bot_branches map may carry: the
// bot PR kinds, derived once from AllPRKinds so the closed set has one home.
var configBotKinds = func() map[PRKind]bool {
	out := map[PRKind]bool{}
	for _, kind := range botKinds {
		out[kind] = true
	}
	return out
}()

// LoadConfig reads and validates one repo's release-intent file. Unknown
// keys fail closed so a typo'd key disables nothing silently. A repo that
// does not release names no surfaces, because nothing there may bump.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, inputError(
			"cannot read release config "+strconv.Quote(path),
			"the file holds the tag prefix, the protected paths, and the bot branches every check compares against",
			"intent.LoadConfig",
			"loading the release config",
			"no title, plan, or gate can be judged without the repo's contract",
			"pass --config for a readable .github/release-intent.yaml, or add the file from the plan's shape",
		)
	}
	return ParseConfig(data)
}

// ParseConfig parses and validates one repo's release-intent file from its
// bytes, with the same strict decoding and value checks LoadConfig applies
// to a path. LoadConfig reads the file and calls this; callers holding the
// file's bytes from another revision, such as a git object, call this
// directly.
func ParseConfig(data []byte) (Config, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, inputError(
			"cannot parse release config bytes: "+err.Error(),
			"the config is strict YAML with a fixed key set",
			"intent.ParseConfig",
			"loading the release config",
			"no title, plan, or gate can be judged without the repo's contract",
			"fix the YAML syntax or the unknown key the message names",
		)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return Config{}, inputError(
				"cannot parse release config bytes: "+err.Error(),
				"the config is one YAML document",
				"intent.ParseConfig",
				"loading the release config",
				"no title, plan, or gate can be judged without the repo's contract",
				"keep a single document in the config",
			)
		}
		return Config{}, inputError(
			"release config bytes hold more than one document",
			"the config is one YAML document",
			"intent.ParseConfig",
			"loading the release config",
			"no title, plan, or gate can be judged without the repo's contract",
			"keep a single document in the config",
		)
	}
	if err := checkConfig(cfg, "release config bytes, parsed by intent.ParseConfig"); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// checkConfig validates the loaded values beyond their YAML shape. where
// names the config under check for the error message.
func checkConfig(cfg Config, where string) error {
	fail := func(what, why, fix string) error {
		return inputError(what, why, where, "validating the release config",
			"no title, plan, or gate can be judged without a valid contract", fix)
	}
	if cfg.TagPrefix == "" {
		return fail("the tag prefix is empty",
			"the prefix tells release tags apart from other tags that share the repository",
			"set tag_prefix, for example \"v\" or \"fairtrade-v\"")
	}
	if cfg.AutomationAuthor == "" {
		return fail("the automation author is empty",
			"the author tells the release PR apart from author PRs that may not touch protected paths",
			"set automation_author to the bot login, for example \"peasant-labs-releaser[bot]\"")
	}
	if cfg.Changelog == "" {
		return fail("the changelog path is empty",
			"the changelog is protected as a whole file and only the release PR may change it",
			"set changelog, for example \"CHANGELOG.md\"")
	}
	for i, vf := range cfg.VersionFiles {
		if vf.Path == "" || vf.Field == "" {
			return fail(fmt.Sprintf("version_files[%d] names path %q and field %q", i, vf.Path, vf.Field),
				"a version file entry needs both the file and the one protected top-level field the release PR writes",
				"set both path and field, for example \"{ path: package.json, field: version }\"")
		}
	}
	for name, stem := range cfg.Surfaces {
		if name == "" || stem == "" {
			return fail("a surfaces entry names an empty surface or stem",
				"each surface maps a closed scope name to the file stem it versions",
				"drop the empty entry or give both the surface name and its stem")
		}
	}
	if !cfg.Releases && len(cfg.Surfaces) > 0 {
		return fail("a repo that does not release names surfaces",
			"surfaces drive bumps, and nothing bumps where releases are off",
			"either set releases to true or drop the surfaces map")
	}
	for kind, branch := range cfg.BotBranches {
		if !configBotKinds[kind] {
			return fail("bot_branches names unknown kind "+strconv.Quote(string(kind)),
				"the bot PR kinds are closed so every workflow compares against the same names",
				"use one of release, vendor-hash or cut-release as the key")
		}
		if branch == "" {
			return fail("bot_branches entry "+strconv.Quote(string(kind))+" names an empty branch",
				"each bot PR kind needs the one head branch its PR uses",
				"set the branch, for example \"automation/release\" for the release kind")
		}
	}
	return nil
}
