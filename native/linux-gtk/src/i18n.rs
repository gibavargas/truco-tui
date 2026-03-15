#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Locale {
    PtBr,
    EnUs,
}

impl Locale {
    pub fn from_code(code: &str) -> Self {
        match code {
            "en-US" => Self::EnUs,
            _ => Self::PtBr,
        }
    }

    pub fn code(self) -> &'static str {
        match self {
            Self::PtBr => "pt-BR",
            Self::EnUs => "en-US",
        }
    }
}

pub fn text(locale: Locale, key: &str) -> &'static str {
    match (locale, key) {
        (Locale::PtBr, "app-title") => "Truco Paulista",
        (Locale::EnUs, "app-title") => "Truco Paulista",
        (Locale::PtBr, "app-subtitle") => "Mesa nativa Linux em libadwaita",
        (Locale::EnUs, "app-subtitle") => "Native Linux table in libadwaita",
        (Locale::PtBr, "play-offline") => "Jogar Offline",
        (Locale::EnUs, "play-offline") => "Play Offline",
        (Locale::PtBr, "create-room") => "Criar Sala",
        (Locale::EnUs, "create-room") => "Create Room",
        (Locale::PtBr, "join-room") => "Entrar",
        (Locale::EnUs, "join-room") => "Join",
        (Locale::PtBr, "copy-key") => "Copiar chave",
        (Locale::EnUs, "copy-key") => "Copy key",
        (Locale::PtBr, "copied") => "Chave copiada",
        (Locale::EnUs, "copied") => "Invite key copied",
        (Locale::PtBr, "missing-core") => "Biblioteca Linux não encontrada. Gere `bin/libtruco_core.so` antes de abrir o app.",
        (Locale::EnUs, "missing-core") => "Linux runtime library not found. Build `bin/libtruco_core.so` before launching the app.",
        (Locale::PtBr, "bad-core") => "A biblioteca carregada é incompatível com este app.",
        (Locale::EnUs, "bad-core") => "The loaded runtime library is incompatible with this app.",
        (Locale::PtBr, "connection-error") => "Falha de conexão",
        (Locale::EnUs, "connection-error") => "Connection failed",
        (Locale::PtBr, "table-ready") => "Mesa pronta",
        (Locale::EnUs, "table-ready") => "Table ready",
        (Locale::PtBr, "online-room") => "Sala Online",
        (Locale::EnUs, "online-room") => "Online Room",
        (Locale::PtBr, "leave-room") => "Sair da sala",
        (Locale::EnUs, "leave-room") => "Leave room",
        (Locale::PtBr, "start-match") => "Iniciar partida",
        (Locale::EnUs, "start-match") => "Start match",
        (Locale::PtBr, "send") => "Enviar",
        (Locale::EnUs, "send") => "Send",
        (Locale::PtBr, "you") => "você",
        (Locale::EnUs, "you") => "you",
        _ => "",
    }
}
