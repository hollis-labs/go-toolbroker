package broker

import (
	"context"
	"fmt"
)

func ExampleLocalBroker_SelectTools() {
	b := NewLocalBroker(nil, []Rule{
		{
			Name:     "exclude-blueprints",
			Intent:   "*",
			Priority: 10,
			Match:    Match{Patterns: []string{"hadron_bp_*"}},
			Action:   Action{Type: "exclude"},
		},
	})

	b.RegisterTools([]ToolDefinition{
		{Name: "volon_tasks_list", Description: "List tasks", Server: "volon"},
		{Name: "hadron_bp_build_volon", Description: "Build volon", Server: "hadron"},
	})

	result, err := b.SelectTools(context.Background(), "any", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Count)
	fmt.Println(result.Tools[0].Name)
	// Output:
	// 1
	// volon_tasks_list
}
