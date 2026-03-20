using System;
using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using WinRT.Interop;

namespace TrucoWinUI;

public sealed partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
        Closed += OnClosed;
        ResizeWindow(1400, 900);
    }

    private void ResizeWindow(int width, int height)
    {
        IntPtr hwnd = WindowNative.GetWindowHandle(this);
        WindowId windowId = Microsoft.UI.Win32Interop.GetWindowIdFromWindow(hwnd);
        AppWindow appWindow = AppWindow.GetFromWindowId(windowId);
        appWindow.Resize(new Windows.Graphics.SizeInt32(width, height));
    }

    private void OnClosed(object sender, WindowEventArgs args)
    {
        if (Content is FrameworkElement element && element.DataContext is IDisposable disposable)
        {
            disposable.Dispose();
        }
    }
}
