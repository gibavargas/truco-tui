using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Media;
using TrucoWinUI.Services;
using TrucoWinUI.ViewModels;

namespace TrucoWinUI;

public partial class App : Application
{
    private Window? _window;

    public App()
    {
        InitializeComponent();
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        try
        {
            var stringProvider = new StringProvider();
            var core = new TrucoCoreService();
            var viewModel = new AppShellViewModel(core, stringProvider);
            _window = new MainWindow(viewModel);
        }
        catch (Exception ex)
        {
            _window = BuildStartupFailureWindow(ex);
        }

        _window.Activate();
    }

    private static Window BuildStartupFailureWindow(Exception ex)
    {
        var window = new Window
        {
            Title = "Truco - Windows startup error",
            Content = new Grid
            {
                Background = new SolidColorBrush(Microsoft.UI.ColorHelper.FromArgb(255, 22, 27, 34)),
                Children =
                {
                    new Border
                    {
                        Padding = new Thickness(32),
                        Child = new ScrollViewer
                        {
                            Content = new StackPanel
                            {
                                Spacing = 16,
                                Children =
                                {
                                    new TextBlock
                                    {
                                        Text = "Unable to start the Windows client",
                                        FontSize = 28,
                                        FontWeight = Microsoft.UI.Text.FontWeights.Bold,
                                        Foreground = new SolidColorBrush(Microsoft.UI.Colors.White),
                                    },
                                    new TextBlock
                                    {
                                        Text = "Check that the portable bundle includes `truco-core-ffi.dll` and that the runtime matches the expected core API/schema version.",
                                        TextWrapping = TextWrapping.Wrap,
                                        Foreground = new SolidColorBrush(Microsoft.UI.ColorHelper.FromArgb(255, 199, 210, 254)),
                                    },
                                    new Border
                                    {
                                        CornerRadius = new CornerRadius(12),
                                        Background = new SolidColorBrush(Microsoft.UI.ColorHelper.FromArgb(255, 15, 23, 42)),
                                        Padding = new Thickness(16),
                                        Child = new TextBlock
                                        {
                                            Text = ex.ToString(),
                                            TextWrapping = TextWrapping.Wrap,
                                            FontFamily = new FontFamily("Consolas"),
                                            Foreground = new SolidColorBrush(Microsoft.UI.ColorHelper.FromArgb(255, 248, 250, 252)),
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        };

        return window;
    }
}
