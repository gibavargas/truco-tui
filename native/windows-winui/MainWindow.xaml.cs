using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Input;
using Microsoft.UI.Xaml.Media.Animation;

namespace TrucoWinUI;

public sealed partial class MainWindow : Window
{
    public MainWindow()
    {
        this.InitializeComponent();
    }
    
    private void Card_PointerEntered(object sender, PointerRoutedEventArgs e)
    {
        if (sender is FrameworkElement element && element.Parent is FrameworkElement parentBtn && parentBtn.Resources.TryGetValue("PointerEnteredStoryboard", out var res) && res is Storyboard sb)
        {
            sb.Begin();
        }
    }

    private void Card_PointerExited(object sender, PointerRoutedEventArgs e)
    {
        if (sender is FrameworkElement element && element.Parent is FrameworkElement parentBtn && parentBtn.Resources.TryGetValue("PointerExitedStoryboard", out var res) && res is Storyboard sb)
        {
            sb.Begin();
        }
    }
}
