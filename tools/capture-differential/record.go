package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// recordCommand writes every root fixture case's expected value from this
// build's observation. Only the baseline build records, so expectations never
// come from the code under test.
func recordCommand(args []string) error {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	root := fs.String("root", "", "root capture fixture to rewrite")
	_ = fs.Parse(args)
	if adapterName != "baseline" {
		return fmt.Errorf("record refused: expectations are recorded only from the baseline build; rerun with -tags capture_baseline and GOWORK=off")
	}
	raw, err := os.ReadFile(*root)
	if err != nil {
		return err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("decode %s: %w", *root, err)
	}
	top := doc.Content[0]
	recorded := 0
	for i := 0; i+1 < len(top.Content); i += 2 {
		arm := top.Content[i+1]
		cases := mappingValue(arm, "cases")
		if cases == nil {
			return fmt.Errorf("arm %s has no cases", top.Content[i].Value)
		}
		for _, c := range cases.Content {
			var rc rootCase
			if err := c.Decode(&rc); err != nil {
				return err
			}
			in, err := rc.Input.document()
			if err != nil {
				return fmt.Errorf("case %s: %w", rc.Name, err)
			}
			o, err := observe(in)
			if err != nil {
				return fmt.Errorf("case %s: %w", rc.Name, err)
			}
			var node yaml.Node
			if err := node.Encode(o); err != nil {
				return err
			}
			expected := mappingValue(c, "expected")
			if expected == nil {
				return fmt.Errorf("case %s has no expected member", rc.Name)
			}
			*expected = node
			recorded++
		}
	}
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(*root, out.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "recorded %d cases into %s\n", recorded, *root)
	return nil
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}
