using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

namespace TrucoWinUI.Services;

public static class NativeDependencyValidator
{
    private static readonly string[] RequiredFiles =
    [
        "truco-core-ffi.dll",
        "libgcc_s_seh-1.dll",
        "libstdc++-6.dll",
        "libwinpthread-1.dll",
    ];

    public static void EnsurePresent()
    {
        string baseDir = AppContext.BaseDirectory;
        List<string> missing = RequiredFiles
            .Where(file => !File.Exists(Path.Combine(baseDir, file)))
            .ToList();

        if (missing.Count == 0)
        {
            return;
        }

        throw new DllNotFoundException(
            "Windows native dependencies are missing from the application output: " +
            string.Join(", ", missing));
    }
}
