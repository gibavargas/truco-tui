using System;
using System.IO;
using System.Runtime.InteropServices;
using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Input;
using WinRT.Interop;
using TrucoWinUI.Models;
using TrucoWinUI.ViewModels;
using Windows.System;

namespace TrucoWinUI;

public sealed partial class MainWindow : Window
{
    private const double WideLayoutThreshold = 1320;
    private const double CompactLayoutThreshold = 980;

    public AppShellViewModel ViewModel { get; } = new();
    private SUBCLASSPROC _subclassProcDelegate;

    public MainWindow()
    {
        InitializeComponent();
        RootPanel.DataContext = ViewModel;
        Closed += OnClosed;
        ConfigureWindow(1400, 900);
    }

    private void ConfigureWindow(int width, int height)
    {
        IntPtr hwnd = WindowNative.GetWindowHandle(this);
        WindowId windowId = Microsoft.UI.Win32Interop.GetWindowIdFromWindow(hwnd);
        AppWindow appWindow = AppWindow.GetFromWindowId(windowId);
        string iconPath = Path.Combine(AppContext.BaseDirectory, "Assets", "truco.ico");
        if (File.Exists(iconPath))
        {
            appWindow.SetIcon(iconPath);
        }

        appWindow.Resize(new Windows.Graphics.SizeInt32(width, height));

        // Subclass window to enforce minimum window dimensions
        _subclassProcDelegate = new SUBCLASSPROC(WindowSubclassProc);
        SetWindowSubclass(hwnd, _subclassProcDelegate, (IntPtr)101, IntPtr.Zero);
    }

    #region Win32 Window Sizing Subclassing

    private delegate IntPtr SUBCLASSPROC(IntPtr hWnd, uint uMsg, IntPtr wParam, IntPtr lParam, IntPtr uIdSubclass, IntPtr dwRefData);

    [DllImport("comctl32.dll", CharSet = CharSet.Auto, SetLastError = true)]
    private static extern bool SetWindowSubclass(IntPtr hWnd, SUBCLASSPROC pfnSubclass, IntPtr uIdSubclass, IntPtr dwRefData);

    [DllImport("comctl32.dll", CharSet = CharSet.Auto, SetLastError = true)]
    private static extern IntPtr DefSubclassProc(IntPtr hWnd, uint uMsg, IntPtr wParam, IntPtr lParam);

    [DllImport("user32.dll")]
    private static extern uint GetDpiForWindow(IntPtr hwnd);

    private const uint WM_GETMINMAXINFO = 0x0024;

    [StructLayout(LayoutKind.Sequential)]
    private struct POINT
    {
        public int x;
        public int y;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct MINMAXINFO
    {
        public POINT ptReserved;
        public POINT ptMaxSize;
        public POINT ptMaxPosition;
        public POINT ptMinTrackSize;
        public POINT ptMaxTrackSize;
    }

    private IntPtr WindowSubclassProc(IntPtr hWnd, uint uMsg, IntPtr wParam, IntPtr lParam, IntPtr uIdSubclass, IntPtr dwRefData)
    {
        if (uMsg == WM_GETMINMAXINFO)
        {
            MINMAXINFO mmi = Marshal.PtrToStructure<MINMAXINFO>(lParam);

            uint dpi = GetDpiForWindow(hWnd);
            float scalingFactor = (dpi == 0) ? 1.0f : (dpi / 96.0f);

            mmi.ptMinTrackSize.x = (int)(960 * scalingFactor);
            mmi.ptMinTrackSize.y = (int)(640 * scalingFactor);

            Marshal.StructureToPtr(mmi, lParam, false);
        }
        return DefSubclassProc(hWnd, uMsg, wParam, lParam);
    }

    #endregion

    private void RootPanel_Loaded(object sender, RoutedEventArgs e)
    {
        UpdateResponsiveLayout();
    }

    private void RootPanel_SizeChanged(object sender, SizeChangedEventArgs e)
    {
        UpdateResponsiveLayout();
    }

    private void RootPanel_KeyDown(object sender, KeyRoutedEventArgs e)
    {
        if (e.Key == VirtualKey.Escape && ViewModel.IsDiagnosticsOpen && ViewModel.CloseDiagnosticsCommand.CanExecute(null))
        {
            ViewModel.CloseDiagnosticsCommand.Execute(null);
            e.Handled = true;
        }
    }

    private void ChatInputBox_KeyDown(object sender, KeyRoutedEventArgs e)
    {
        if (e.Key == VirtualKey.Enter && ViewModel.SendChatCommand.CanExecute(null))
        {
            ViewModel.SendChatCommand.Execute(null);
            e.Handled = true;
        }
    }

    private void UpdateResponsiveLayout()
    {
        if (GameLayoutGrid is null || MainBoardBorder is null || SidebarBorder is null || HeaderActionPanel is null || HeaderBorder is null)
        {
            return;
        }

        double width = RootPanel.ActualWidth;
        bool wideLayout = width >= WideLayoutThreshold;
        bool compactLayout = width < CompactLayoutThreshold;

        GameLayoutGrid.ColumnSpacing = compactLayout ? 12 : 16;
        GameLayoutGrid.RowSpacing = compactLayout ? 12 : 16;
        GameLayoutGrid.Margin = compactLayout ? new Thickness(12) : new Thickness(20);
        MainBoardBorder.Padding = compactLayout ? new Thickness(12) : new Thickness(16);
        SidebarBorder.Padding = compactLayout ? new Thickness(12) : new Thickness(16);
        MainBoardBorder.Margin = compactLayout ? new Thickness(0, 8, 0, 0) : new Thickness(16);
        SidebarBorder.Margin = compactLayout ? new Thickness(0, 8, 0, 0) : new Thickness(0, 16, 0, 16);
        MainBoardBorder.CornerRadius = compactLayout ? new CornerRadius(64) : new CornerRadius(120);
        HeaderBorder.Padding = compactLayout ? new Thickness(12, 10) : new Thickness(16, 12);
        MainBoardBorder.BorderThickness = new Thickness(2);
        SidebarBorder.BorderThickness = compactLayout ? new Thickness(1) : new Thickness(1);
        HeaderActionPanel.Orientation = compactLayout ? Orientation.Vertical : Orientation.Horizontal;
        HeaderActionPanel.HorizontalAlignment = compactLayout ? HorizontalAlignment.Stretch : HorizontalAlignment.Right;
        HeaderActionPanel.Spacing = compactLayout ? 6 : 8;

        if (wideLayout)
        {
            GameMainColumn.Width = new GridLength(1, GridUnitType.Star);
            GameSidebarColumn.Width = new GridLength(compactLayout ? 320 : 360);
            GameMainRow.Height = new GridLength(1, GridUnitType.Star);
            GameSidebarRow.Height = GridLength.Auto;

            Grid.SetRow(MainBoardBorder, 1);
            Grid.SetColumn(MainBoardBorder, 0);
            Grid.SetRowSpan(MainBoardBorder, 1);
            Grid.SetColumnSpan(MainBoardBorder, 1);

            Grid.SetRow(SidebarBorder, 1);
            Grid.SetColumn(SidebarBorder, 1);
            Grid.SetRowSpan(SidebarBorder, 1);
            Grid.SetColumnSpan(SidebarBorder, 1);
        }
        else
        {
            GameMainColumn.Width = new GridLength(1, GridUnitType.Star);
            GameSidebarColumn.Width = new GridLength(0);
            GameMainRow.Height = new GridLength(1, GridUnitType.Star);
            GameSidebarRow.Height = GridLength.Auto;

            Grid.SetRow(MainBoardBorder, 1);
            Grid.SetColumn(MainBoardBorder, 0);
            Grid.SetRowSpan(MainBoardBorder, 1);
            Grid.SetColumnSpan(MainBoardBorder, 1);

            Grid.SetRow(SidebarBorder, 2);
            Grid.SetColumn(SidebarBorder, 0);
            Grid.SetRowSpan(SidebarBorder, 1);
            Grid.SetColumnSpan(SidebarBorder, 1);
        }
    }

    private void PlayCardButton_Click(object sender, RoutedEventArgs e)
    {
        if (sender is not Button button || button.Tag is not CardState card)
        {
            return;
        }

        ViewModel.PlayCardCommand.Execute(card);
    }

    private void OnClosed(object sender, WindowEventArgs args)
    {
        ViewModel.Dispose();
    }

}
