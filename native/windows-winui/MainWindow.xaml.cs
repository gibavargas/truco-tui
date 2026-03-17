using Microsoft.UI.Xaml;
using TrucoWinUI.ViewModels;

namespace TrucoWinUI;

public sealed partial class MainWindow : Window
{
    private readonly AppShellViewModel _viewModel;

    public MainWindow(AppShellViewModel viewModel)
    {
        _viewModel = viewModel ?? throw new ArgumentNullException(nameof(viewModel));
        InitializeComponent();
        RootPanel.DataContext = _viewModel;
        Closed += OnClosed;
    }

    public MainWindow() : this(new AppShellViewModel())
    {
    }

    private void OnClosed(object sender, WindowEventArgs args)
    {
        Closed -= OnClosed;
        _viewModel.Dispose();
    }
}
