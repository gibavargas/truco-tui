using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Data;
using System;

namespace TrucoWinUI.Converters;

public sealed class BoolToVisibilityConverter : IValueConverter
{
    public object Convert(object value, Type targetType, object parameter, string language)
    {
        bool isTrue = value is bool flag && flag;
        bool invert = parameter is string text && string.Equals(text, "invert", StringComparison.OrdinalIgnoreCase);
        if (invert)
        {
            isTrue = !isTrue;
        }

        return isTrue ? Visibility.Visible : Visibility.Collapsed;
    }

    public object ConvertBack(object value, Type targetType, object parameter, string language)
    {
        bool isVisible = value is Visibility visibility && visibility == Visibility.Visible;
        bool invert = parameter is string text && string.Equals(text, "invert", StringComparison.OrdinalIgnoreCase);
        if (invert)
        {
            isVisible = !isVisible;
        }

        return isVisible;
    }
}
