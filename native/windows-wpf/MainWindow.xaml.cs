using System.Windows;

namespace TrucoWPF;

public partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
    }

    protected override void OnClosed(System.EventArgs e)
    {
        base.OnClosed(e);
        if (DataContext is IDisposable vm)
        {
            vm.Dispose();
        }
    }
}
