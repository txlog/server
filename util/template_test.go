package util

import (
	"html/template"
	"regexp"
	"testing"
)

func TestText2HTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected template.HTML
	}{
		{
			name:     "empty string",
			input:    "",
			expected: template.HTML(""),
		},
		{
			name:     "string with newlines",
			input:    "hello\nworld",
			expected: template.HTML("hello<br>world"),
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: template.HTML("hello&nbsp;world"),
		},
		{
			name:     "string with spaces and newlines",
			input:    "hello world\nfoo bar",
			expected: template.HTML("hello&nbsp;world<br>foo&nbsp;bar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Text2HTML(tt.input); got != tt.expected {
				t.Errorf("Text2HTML() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "zero",
			input:    0.0,
			expected: "0,00",
		},
		{
			name:     "positive small number",
			input:    12.34,
			expected: "12,34",
		},
		{
			name:     "negative small number",
			input:    -12.34,
			expected: "-12,34",
		},
		{
			name:     "large positive number",
			input:    1234567.89,
			expected: "1.234.567,89",
		},
		{
			name:     "large negative number",
			input:    -1234567.89,
			expected: "-1.234.567,89",
		},
		{
			name:     "small decimal",
			input:    0.01,
			expected: "0,01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPercentage(tt.input); got != tt.expected {
				t.Errorf("FormatPercentage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestFormatInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "small positive number",
			input:    123,
			expected: "123",
		},
		{
			name:     "small negative number",
			input:    -123,
			expected: "-123",
		},
		{
			name:     "medium positive number",
			input:    1234,
			expected: "1.234",
		},
		{
			name:     "medium negative number",
			input:    -1234,
			expected: "-1.234",
		},
		{
			name:     "large positive number",
			input:    1234567,
			expected: "1.234.567",
		},
		{
			name:     "large negative number",
			input:    -1234567,
			expected: "-1.234.567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatInteger(tt.input); got != tt.expected {
				t.Errorf("FormatInteger() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestIterate(t *testing.T) {
	tests := []struct {
		name     string
		start    int
		count    int
		expected []int
	}{
		{
			name:     "zero range",
			start:    0,
			count:    0,
			expected: []int{0},
		},
		{
			name:     "positive range",
			start:    1,
			count:    5,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "negative range",
			start:    -2,
			count:    2,
			expected: []int{-2, -1, 0, 1, 2},
		},
		{
			name:     "single number",
			start:    3,
			count:    3,
			expected: []int{3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Iterate(tt.start, tt.count)
			if len(got) != len(tt.expected) {
				t.Errorf("Iterate() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("Iterate()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        2,
			b:        3,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -2,
			b:        -3,
			expected: -5,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 5,
		},
		{
			name:     "positive and negative",
			a:        5,
			b:        -3,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.a, tt.b); got != tt.expected {
				t.Errorf("Add() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a less than b",
			a:        2,
			b:        3,
			expected: 2,
		},
		{
			name:     "b less than a",
			a:        5,
			b:        1,
			expected: 1,
		},
		{
			name:     "equal values",
			a:        4,
			b:        4,
			expected: 4,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        -2,
			expected: -5,
		},
		{
			name:     "positive and negative",
			a:        3,
			b:        -1,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.a, tt.b); got != tt.expected {
				t.Errorf("Min() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestVersion(t *testing.T) {
	got := Version()
	if got == "" {
		t.Error("Version() should not return empty string")
	}
	// Version should follow semantic versioning format (x.y.z)
	if matched := regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(got); !matched {
		t.Errorf("Version() = %v, want format x.y.z", got)
	}
}
func TestDnfUser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "Unknown",
		},
		{
			name:     "simple username",
			input:    "simple.user",
			expected: "simple.user",
		},
		{
			name:     "dnf format",
			input:    "rodrigo avila <rodrigo.avila>",
			expected: "rodrigo.avila",
		},
		{
			name:     "incomplete brackets",
			input:    "rodrigo avila <rodrigo.avila",
			expected: "rodrigo avila <rodrigo.avila",
		},
		{
			name:     "incomplete brackets 2",
			input:    "rodrigo avila rodrigo.avila>",
			expected: "rodrigo avila rodrigo.avila>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DnfUser(tt.input); got != tt.expected {
				t.Errorf("DnfUser() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestBrand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "brand-linux.svg",
		},
		{
			name:     "almalinux lowercase",
			input:    "almalinux",
			expected: "brand-almalinux.svg",
		},
		{
			name:     "almalinux mixed case",
			input:    "AlmaLinux",
			expected: "brand-almalinux.svg",
		},
		{
			name:     "centos",
			input:    "CentOS Linux",
			expected: "brand-centos.svg",
		},
		{
			name:     "fedora",
			input:    "Fedora Linux",
			expected: "brand-fedora.svg",
		},
		{
			name:     "oracle",
			input:    "Oracle Linux",
			expected: "brand-oracle.svg",
		},
		{
			name:     "red hat",
			input:    "Red Hat Enterprise Linux",
			expected: "brand-redhat.svg",
		},
		{
			name:     "rocky",
			input:    "Rocky Linux",
			expected: "brand-rocky.svg",
		},
		{
			name:     "unknown distribution",
			input:    "Unknown Linux",
			expected: "brand-linux.svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Brand(tt.input); got != tt.expected {
				t.Errorf("Brand() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestHasAction(t *testing.T) {
	tests := []struct {
		name     string
		actions  string
		action   string
		expected bool
	}{
		{
			name:     "single install action code",
			actions:  "I",
			action:   "Install",
			expected: true,
		},
		{
			name:     "multiple action codes with match",
			actions:  "I,D,O,U,E,R,C",
			action:   "Downgrade",
			expected: true,
		},
		{
			name:     "multiple action codes without match",
			actions:  "I,D,O",
			action:   "Reinstall",
			expected: false,
		},
		{
			name:     "direct word match",
			actions:  "Install",
			action:   "Install",
			expected: true,
		},
		{
			name:     "empty actions",
			actions:  "",
			action:   "Install",
			expected: false,
		},
		{
			name:     "case sensitive match",
			actions:  "I",
			action:   "install",
			expected: false,
		},
		{
			name:     "reason change action",
			actions:  "C",
			action:   "Reason Change",
			expected: true,
		},
		{
			name:     "spaces in action codes",
			actions:  "I, D, O",
			action:   "Downgrade",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAction(tt.actions, tt.action); got != tt.expected {
				t.Errorf("HasAction() = %v, want %v", got, tt.expected)
			}
		})
	}
}
