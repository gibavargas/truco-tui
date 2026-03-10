# Windows WinUI 3 shell

This scaffold assumes a packaged WinUI 3 desktop app using:

- CommunityToolkit.Mvvm for view models and commands
- a `TrucoCoreService` wrapping the generated C ABI
- UI updates marshaled through `DispatcherQueue`

The sample code is intentionally minimal and should be opened in Visual Studio
on Windows after adding the generated `truco_core.dll` bridge artifacts.
