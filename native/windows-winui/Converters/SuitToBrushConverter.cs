using Microsoft.UI.Xaml.Data;
using Microsoft.UI.Xaml.Media;
using System;
using Windows.UI;

namespace TrucoWinUI.Converters;

public sealed class SuitToBrushConverter : IValueConverter
{
    private static readonly SolidColorBrush RedBrush = new(Color.FromArgb(255, 191, 37, 37));
    private static readonly SolidColorBrush BlackBrush = new(Color.FromArgb(255, 31, 43, 56));

    public object Convert(object value, Type targetType, object parameter, string language)
        => value is bool isRed && isRed ? RedBrush : BlackBrush;

    public object ConvertBack(object value, Type targetType, object parameter, string language)
        => throw new NotSupportedException();
}
