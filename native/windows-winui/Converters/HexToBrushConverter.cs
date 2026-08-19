using Microsoft.UI.Xaml.Data;
using Microsoft.UI.Xaml.Media;
using System;
using Windows.UI;

namespace TrucoWinUI.Converters;

public sealed class HexToBrushConverter : IValueConverter
{
    public object Convert(object value, Type targetType, object parameter, string language)
    {
        if (value is string text && TryParseHexColor(text, out Color color))
        {
            return new SolidColorBrush(color);
        }

        return new SolidColorBrush(Color.FromArgb(255, 224, 208, 190));
    }

    public object ConvertBack(object value, Type targetType, object parameter, string language)
        => throw new NotSupportedException();

    private static bool TryParseHexColor(string value, out Color color)
    {
        color = default;
        string text = value.Trim().TrimStart('#');
        if (text.Length is not (6 or 8))
        {
            return false;
        }

        try
        {
            byte a = 255;
            int offset = 0;
            if (text.Length == 8)
            {
                a = System.Convert.ToByte(text.Substring(0, 2), 16);
                offset = 2;
            }

            byte r = System.Convert.ToByte(text.Substring(offset, 2), 16);
            byte g = System.Convert.ToByte(text.Substring(offset + 2, 2), 16);
            byte b = System.Convert.ToByte(text.Substring(offset + 4, 2), 16);
            color = Color.FromArgb(a, r, g, b);
            return true;
        }
        catch
        {
            return false;
        }
    }
}
