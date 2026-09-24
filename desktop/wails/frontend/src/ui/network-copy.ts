import type { NetworkSnapshot } from "../types";

type Translate = (key: string, ...args: Array<string | number>) => string;

export function activeTransportLabel(network: NetworkSnapshot | undefined, t: Translate): string {
  return transportLabel(network?.transport, t) || "-";
}

export function requestedTransportLabel(network: NetworkSnapshot | undefined, t: Translate): string {
  const value = transportLabel(network?.requested_transport, t);
  return value || t("transport_auto");
}

export function connectionPathLabel(network: NetworkSnapshot | undefined, t: Translate): string {
  if (!network) {
    return "-";
  }
  if (network.transport === "tailnet_tsnet_v1" || network.tailnet_authority || network.tailnet_node) {
    return t("connection_path_tailnet");
  }
  if (network.direct_path === true) {
    return t("connection_path_direct");
  }
  if (network.relay_fallback) {
    return t("connection_path_relay");
  }
  return t("connection_path_unknown");
}

export function fallbackReasonLabel(network: NetworkSnapshot | undefined, t: Translate): string {
  if (!network?.relay_fallback) {
    return "";
  }
  return network.fallback_reason || t("connection_fallback_active");
}

export function seatProtocolLabel(network: NetworkSnapshot | undefined): string {
  const versions = network?.seat_protocol_versions;
  if (!versions) {
    return "";
  }

  const entries = Object.entries(versions)
    .map(([seat, version]) => [Number.parseInt(seat, 10), version] as const)
    .filter(([seat, version]) => Number.isFinite(seat) && typeof version === "number")
    .sort((left, right) => left[0] - right[0]);

  if (entries.length === 0) {
    return "";
  }

  return entries
    .map(([seat, version]) => `#${seat + 1}: v${version}`)
    .join(" · ");
}

function transportLabel(value: string | undefined, t: Translate): string {
  switch (value) {
    case "auto":
    case "":
    case undefined:
      return "";
    case "tcp_tls":
      return t("transport_direct");
    case "relay_quic_v2":
      return t("transport_relay");
    case "tailnet_tsnet_v1":
      return "Tailnet";
    default:
      return value;
  }
}
