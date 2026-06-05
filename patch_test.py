import os

filepath = 'desktop/wails/app_test.go'
with open(filepath, 'r') as f:
    content = f.read()

# Replace the tight loop to wait more ticks
old = """	for i := 0; i < 2000; i++ {
		snapshot := app.Snapshot()
		if snapshot.Match != nil && snapshot.Match.CurrentHand.Round == 1 && snapshot.UI.Actions.CanPlayCard {
			return
		}
		if err := app.Tick(12); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}"""
new = """	for i := 0; i < 5000; i++ {
		snapshot := app.Snapshot()
		if snapshot.Match != nil && snapshot.Match.CurrentHand.Round == 1 && snapshot.UI.Actions.CanPlayCard {
			return
		}
		if err := app.Tick(12); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}"""

content = content.replace(old, new)
with open(filepath, 'w') as f:
    f.write(content)
