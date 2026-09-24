namespace TrucoWinUI.Models;

public static class GameStateHelper
{
    public static string GetTrucoLabel(int? pendingRaiseTo)
    {
        return pendingRaiseTo switch
        {
            3 => "TRUCO!",
            6 => "SEIS!",
            9 => "NOVE!",
            12 => "DOZE!",
            _ => "TRUCO!"
        };
    }

    public static string GetAskTrucoLabel(int? currentStake)
    {
        return currentStake switch
        {
            1 => "TRUCO!",
            3 => "SEIS!",
            6 => "NOVE!",
            9 => "DOZE!",
            _ => "TRUCO!"
        };
    }

    public static string GetRelativeTeamLabel(int team, int localTeam)
    {
        if (team < 0)
        {
            return "-";
        }

        return team == localTeam ? "NÓS" : "ELES";
    }
}
