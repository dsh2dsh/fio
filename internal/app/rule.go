package app

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"text/template/parse"
)

type SectionRule struct {
	Key     string
	Re      string
	Account string
	Vs      string
	If      string

	keyTemplate *template.Template
	reCompiled  *regexp.Regexp
	ifTemplate  *template.Template
}

func (self *SectionRule) ExtractKey(rec Record) (string, error) {
	if !self.knownAccount(rec) {
		return "", nil
	}

	if self.reCompiled != nil {
		return self.extractKeyRe(rec)
	}

	if stop, s, err := self.templatedChecks(rec); err != nil {
		return "", err
	} else if stop {
		return s, err
	}

	return self.accountKey(rec), nil
}

func (self *SectionRule) knownAccount(rec Record) bool {
	if self.Account != "" {
		if self.Account != rec.AccountId() {
			return false
		} else if self.Vs != "" && self.Vs != rec.Vs() {
			return false
		}
	}
	return true
}

func (self *SectionRule) templatedChecks(rec Record) (bool, string, error) {
	if ok, err := self.templatedIf(rec); err != nil {
		return false, "", err
	} else if !ok {
		return true, "", nil
	}

	if k, err := self.templatedKey(rec); err != nil {
		return false, "", err
	} else if k != "" {
		return true, k, nil
	}

	return false, "", nil
}

func (self *SectionRule) extractKeyRe(rec Record) (string, error) {
	s := self.reCompiled.FindStringSubmatch(rec.Note())
	if s == nil {
		return "", nil
	}

	if stop, s, err := self.templatedChecks(rec); err != nil {
		return "", err
	} else if stop {
		return s, err
	}

	switch {
	case len(s) > 1:
		return s[1], nil
	case self.Account != "":
		return self.accountKey(rec), nil
	}

	return rec.Note(), nil
}

func (self *SectionRule) accountKey(rec Record) string {
	if self.Account == "" {
		return ""
	}
	account := self.Account

	if rec.Vs() != "" {
		account += ", VS: " + rec.Vs()
	}

	if rec.Note() != "" {
		account += ", " + rec.Note()
	}

	return account
}

func (self *SectionRule) templatedKey(rec Record) (string, error) {
	if self.keyTemplate == nil {
		return self.Key, nil
	}

	var b bytes.Buffer
	if err := self.keyTemplate.Execute(&b, &rec); err != nil {
		return "", fmt.Errorf("extract rule key from %q: %w", self.Key, err)
	}
	return strings.Trim(b.String(), " "), nil
}

func (self *SectionRule) compile(sectName string, idx int) error {
	if self.Re == "" && self.Key == "" && self.Account == "" {
		return fmt.Errorf(
			"config compile: section %q, rule %d: both key and account empty",
			sectName, idx)
	}

	if err := self.compileTemplates(sectName, idx); err != nil {
		return err
	}

	if reCompiled, err := regexp.Compile(self.Re); err != nil {
		return fmt.Errorf("re compile %q: %w", self.Re, err)
	} else {
		self.reCompiled = reCompiled
	}

	return nil
}

func (self *SectionRule) compileTemplates(sectName string, idx int) error {
	keys := []struct {
		tmpl string
		prop **template.Template
	}{
		{self.Key, &self.keyTemplate},
		{self.If, &self.ifTemplate},
	}

	for _, v := range keys {
		if t, err := self.compileOneTemplate(v.tmpl); err != nil {
			return fmt.Errorf("config: section %q, rule %d: %w", sectName, idx, err)
		} else if t != nil {
			*v.prop = t
		}
	}

	return nil
}

func (self *SectionRule) compileOneTemplate(tmpl string) (*template.Template, error) {
	if tmpl == "" {
		return nil, nil
	}

	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("compile %q: %w", tmpl, err)
	}

	nodes := t.Root.Nodes
	if len(nodes) == 0 {
		return nil, nil
	}

	if nodes[0].Type() != parse.NodeText || len(nodes) > 1 {
		return t, nil
	}

	return nil, nil
}

func (self *SectionRule) templatedIf(rec Record) (bool, error) {
	if self.ifTemplate == nil {
		return true, nil
	}

	var b bytes.Buffer
	if err := self.ifTemplate.Execute(&b, &rec); err != nil {
		return false, fmt.Errorf("check rule if %q: %w", self.If, err)
	}
	s := strings.Trim(b.String(), " ")

	return len(s) > 0, nil
}
