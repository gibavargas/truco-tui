using System;
using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using WinRT.Interop;
using TrucoWinUI.Models;
using TrucoWinUI.ViewModels;

namespace TrucoWinUI;

public sealed partial class MainWindow : Window
{
    private const double WideLayoutThreshold = 1320;
    private const double CompactLayoutThreshold = 980;

    public AppShellViewModel ViewModel { get; } = new();

    public MainWindow()
    {
        InitializeComponent();
        RootPanel.DataContext = ViewModel;
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

    private void RootPanel_Loaded(object sender, RoutedEventArgs e)
    {
        UpdateResponsiveLayout();
    }

    private void RootPanel_SizeChanged(object sender, SizeChangedEventArgs e)
    {
        UpdateResponsiveLayout();
    }



    private void UpdateResponsiveLayout()
    {
        if (GameLayoutGrid is null || MainBoardBorder is null || SidebarBorder is null)
        {
            return;
        }

        double width = RootPanel.ActualWidth;
        bool wideLayout = width >= WideLayoutThreshold;
        bool compactLayout = width < CompactLayoutThreshold;

        GameLayoutGrid.ColumnSpacing = compactLayout ? 12 : 16;
        GameLayoutGrid.RowSpacing = compactLayout ? 12 : 16;
        MainBoardBorder.Padding = compactLayout ? new Thickness(12) : new Thickness(16);
        SidebarBorder.Padding = compactLayout ? new Thickness(12) : new Thickness(16);
        MainBoardBorder.BorderThickness = compactLayout ? new Thickness(10) : new Thickness(12);
        SidebarBorder.BorderThickness = compactLayout ? new Thickness(1) : new Thickness(1);

        if (wideLayout)
        {
            GameMainColumn.Width = new GridLength(1, GridUnitType.Star);
            GameSidebarColumn.Width = new GridLength(compactLayout ? 320 : 360);
            GameMainRow.Height = new GridLength(1, GridUnitType.Star);
            GameSidebarRow.Height = GridLength.Auto;

            Grid.SetRow(MainBoardBorder, 0);
            Grid.SetColumn(MainBoardBorder, 0);
            Grid.SetRowSpan(MainBoardBorder, 2);
            Grid.SetColumnSpan(MainBoardBorder, 1);

            Grid.SetRow(SidebarBorder, 0);
            Grid.SetColumn(SidebarBorder, 1);
            Grid.SetRowSpan(SidebarBorder, 2);
            Grid.SetColumnSpan(SidebarBorder, 1);
        }
        else
        {
            GameMainColumn.Width = new GridLength(1, GridUnitType.Star);
            GameSidebarColumn.Width = new GridLength(0);
            GameMainRow.Height = new GridLength(1, GridUnitType.Star);
            GameSidebarRow.Height = GridLength.Auto;

            Grid.SetRow(MainBoardBorder, 0);
            Grid.SetColumn(MainBoardBorder, 0);
            Grid.SetRowSpan(MainBoardBorder, 1);
            Grid.SetColumnSpan(MainBoardBorder, 1);

            Grid.SetRow(SidebarBorder, 1);
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
