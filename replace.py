import re

with open("desktop/wails/app_test.go", "r") as f:
    content = f.read()

# Replace dispatch calls with correct struct fields
content = content.replace("""app.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerName: "Mesa",
		NumPlayers: 2,
		SeedLo:     123,
		SeedHi:     456,
	})""", """app.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerNames: []string{"Mesa", "CPU 1"},
		CPUFlags:    []bool{false, true},
		SeedLo:      123,
		SeedHi:      456,
	})""")

content = content.replace("""app.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerName: "Mesa",
		NumPlayers: 4,
		SeedLo:     123,
		SeedHi:     456,
	})""", """app.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerNames: []string{"Mesa", "CPU 1", "CPU 2", "CPU 3"},
		CPUFlags:    []bool{false, true, true, true},
		SeedLo:      123,
		SeedHi:      456,
	})""")

with open("desktop/wails/app_test.go", "w") as f:
    f.write(content)
