const bridge = () => window.go?.main?.App;

const state = {
  locale: "pt-BR",
  snapshot: null,
};

const els = {
  playerName: document.getElementById("player-name"),
  numPlayers: document.getElementById("num-players"),
  inviteKey: document.getElementById("invite-key"),
  summary: document.getElementById("summary"),
  cards: document.getElementById("cards"),
  events: document.getElementById("events"),
  chatText: document.getElementById("chat-text"),
};

async function call(method, ...args) {
  const api = bridge();
  if (!api || typeof api[method] !== "function") {
    throw new Error("Wails bridge unavailable");
  }
  const result = await api[method](...args);
  if (result && result.message) {
    appendEvent(`error: ${result.message}`);
  }
  await refresh();
  return result;
}

function appendEvent(text) {
  els.events.textContent = `${text}\n${els.events.textContent}`.trim();
}

function cardLabel(card) {
  if (!card) return "";
  const suitMap = { Espadas: "♠", Copas: "♥", Ouros: "♦", Paus: "♣" };
  return `${card.Rank || card.rank || ""}${suitMap[card.Suit || card.suit] || ""}`;
}

function render(snapshot) {
  state.snapshot = snapshot;
  const match = snapshot?.match;
  const lobby = snapshot?.lobby;
  if (!snapshot) {
    els.summary.textContent = "No runtime snapshot.";
    els.cards.innerHTML = "";
    return;
  }

  const lines = [
    `Mode: ${snapshot.mode}`,
    `Locale: ${snapshot.locale}`,
  ];
  if (lobby) {
    lines.push(`Invite: ${lobby.invite_key || "-"}`);
    lines.push(`Seats: ${(lobby.slots || []).join(" | ")}`);
  }
  if (match) {
    const score = match.MatchPoints || {};
    lines.push(`Score: ${score["0"] ?? score[0] ?? 0} x ${score["1"] ?? score[1] ?? 0}`);
    lines.push(`Turn: ${match.TurnPlayer}`);
  }
  els.summary.textContent = lines.join("\n");

  els.cards.innerHTML = "";
  const me = (match?.Players || []).find((player) => (player.ID ?? player.id) === 0);
  for (const card of me?.Hand || me?.hand || []) {
    const node = document.createElement("button");
    node.className = "card";
    node.textContent = cardLabel(card);
    node.onclick = async () => {
      const hand = me?.Hand || me?.hand || [];
      const idx = hand.findIndex((item) => cardLabel(item) === cardLabel(card));
      if (idx >= 0) {
        await call("PlayCard", idx);
      }
    };
    els.cards.appendChild(node);
  }
}

async function refresh() {
  const api = bridge();
  if (!api) return;
  const snapshot = await api.Snapshot();
  render(snapshot);
  const events = await api.PollEvents();
  for (const ev of events || []) {
    appendEvent(`${ev.kind}: ${JSON.stringify(ev.payload ?? {})}`);
  }
}

document.getElementById("start-offline").onclick = async () => {
  await call("StartOfflineGame", els.playerName.value, Number(els.numPlayers.value));
  await call("Tick", 12);
};
document.getElementById("create-host").onclick = async () => {
  await call("CreateHostSession", els.playerName.value, Number(els.numPlayers.value), "", "", "");
};
document.getElementById("join-session").onclick = async () => {
  await call("JoinSession", els.inviteKey.value, els.playerName.value, "auto");
};
document.getElementById("start-match").onclick = async () => {
  await call("StartHostedMatch");
};
document.getElementById("new-hand").onclick = async () => {
  await call("NewHand");
};
document.getElementById("tick").onclick = async () => {
  await call("Tick", 12);
};
document.getElementById("truco").onclick = async () => call("RequestTruco");
document.getElementById("accept").onclick = async () => call("AcceptTruco");
document.getElementById("refuse").onclick = async () => call("RefuseTruco");
document.getElementById("send-chat").onclick = async () => {
  await call("SendChat", els.chatText.value);
  els.chatText.value = "";
};
document.getElementById("vote-host").onclick = async () => call("VoteHost", 0);
document.getElementById("replacement").onclick = async () => call("RequestReplacementInvite", 1);
document.getElementById("close-session").onclick = async () => call("CloseSession");
document.getElementById("locale-toggle").onclick = async () => {
  state.locale = state.locale === "pt-BR" ? "en-US" : "pt-BR";
  await call("SetLocale", state.locale);
};

setInterval(refresh, 1200);
refresh().catch((err) => appendEvent(err.message));
