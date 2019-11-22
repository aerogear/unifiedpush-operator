package unifiedpushserver

import (
	"reflect"
	"testing"
)

func TestResourcesAdd(t *testing.T) {

	scenarios := []struct {
		name     string
		given    resources
		expect   resources
		mutation func(input resources)
	}{
		{
			name:     "when struct is nil",
			given:    nil,
			expect:   nil,
			mutation: func(input resources) { input.add("a", "b") },
		},
		{
			name:     "when key is missing",
			given:    resources{},
			expect:   resources{"a": []string{"b"}},
			mutation: func(input resources) { input.add("a", "b") },
		},
		{
			name:     "when key is nil",
			given:    resources{"a": nil},
			expect:   resources{"a": []string{"b"}},
			mutation: func(input resources) { input.add("a", "b") },
		},
		{
			name:     "when key exists",
			given:    resources{"a": []string{"b", "c", "d"}},
			expect:   resources{"a": []string{"b", "c", "d", "e"}},
			mutation: func(input resources) { input.add("a", "e") },
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.mutation(scenario.given)
			if !reflect.DeepEqual(scenario.given, scenario.expect) {
				t.Fatalf(
					"Actual vs. expected differs:\nActual: %v\nExpected: %v",
					scenario.given,
					scenario.expect,
				)
			}
		})
	}

}

func TestResourcesRemove(t *testing.T) {

	scenarios := []struct {
		name     string
		given    resources
		expect   resources
		mutation func(input resources)
	}{
		{
			name:     "when struct is nil",
			given:    nil,
			expect:   nil,
			mutation: func(input resources) { input.remove("a", "b") },
		},
		{
			name:     "when key is missing",
			given:    resources{},
			expect:   resources{},
			mutation: func(input resources) { input.remove("a", "b") },
		},
		{
			name:     "when key is nil",
			given:    resources{"a": nil},
			expect:   resources{"a": nil},
			mutation: func(input resources) { input.remove("a", "b") },
		},
		{
			name:     "when removing from end of list",
			given:    resources{"a": []string{"b", "c", "d", "e"}},
			expect:   resources{"a": []string{"b", "c", "d"}},
			mutation: func(input resources) { input.remove("a", "e") },
		},
		{
			name:     "when removing from head of list, tail element is moved",
			given:    resources{"a": []string{"b", "c", "d", "e"}},
			expect:   resources{"a": []string{"e", "c", "d"}},
			mutation: func(input resources) { input.remove("a", "b") },
		},
		{
			name:     "when removing from middle of list, tail element is moved",
			given:    resources{"a": []string{"b", "c", "d", "e"}},
			expect:   resources{"a": []string{"b", "e", "d"}},
			mutation: func(input resources) { input.remove("a", "c") },
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.mutation(scenario.given)
			if !reflect.DeepEqual(scenario.given, scenario.expect) {
				t.Fatalf(
					"Actual vs. expected differs:\nActual: %v\nExpected: %v",
					scenario.given,
					scenario.expect,
				)
			}
		})
	}

}
