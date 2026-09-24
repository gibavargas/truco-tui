using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

namespace TrucoWinUI.Services;

public static class NativeDependencyValidator
{
    private const string CoreLibrary = "truco-core-ffi.dll";

    private static readonly string[] RequiredFiles =
    [
        CoreLibrary,
        "libgcc_s_seh-1.dll",
        "libstdc++-6.dll",
        "libwinpthread-1.dll",
    ];

    public static void EnsurePresent(string? coreLibraryPath = null)
    {
        string baseDir = string.IsNullOrWhiteSpace(coreLibraryPath)
            ? AppContext.BaseDirectory
            : Path.GetDirectoryName(Path.GetFullPath(coreLibraryPath)) ?? AppContext.BaseDirectory;

        List<string> missing = RequiredFiles
            .Where(file => !File.Exists(Path.Combine(baseDir, file)))
            .ToList();

        if (missing.Count == 0)
        {
            return;
        }

        throw new DllNotFoundException(
            $"Windows native dependencies are missing from {baseDir}: " +
            string.Join(", ", missing) +
            ". Build the portable bundle so the Go FFI DLL and MinGW runtime DLLs are copied together.");
    }
}
