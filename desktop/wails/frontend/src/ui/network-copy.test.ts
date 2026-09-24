import test from "node:test";
import assert from "node:assert/strict";

import {
  activeTransportLabel,
  connectionPathLabel,
  fallbackReasonLabel,
  requestedTransportLabel,
  seatProtocolLabel,
} from "./network-copy";
import type { NetworkSnapshot } from "../types";

const t = (key: string): string => key;

test("network copy helpers expose readable transport and path facts", () => {
  const network: NetworkSnapshot = {
    transport: "relay_quic_v2",
    requested_transport: "tcp_tls",
    relay_fallback: true,
    fallback_reason: "no direct route",
    seat_protocol_versions: { "0": 2, "3": 1 },
  };

  assert.equal(activeTransportLabel(network, t), "transport_relay");
  assert.equal(requestedTransportLabel(network, t), "transport_direct");
  assert.equal(connectionPathLabel(network, t), "connection_path_relay");
  assert.equal(fallbackReasonLabel(network, t), "no direct route");
  assert.equal(seatProtocolLabel(network), "#1: v2 · #4: v1");
});

test("network copy helpers gracefully handle tailnet and auto defaults", () => {
  const network: NetworkSnapshot = {
    transport: "tailnet_tsnet_v1",
    tailnet_authority: "tail.example.ts.net",
  };

  assert.equal(activeTransportLabel(network, t), "Tailnet");
  assert.equal(requestedTransportLabel(network, t), "transport_auto");
  assert.equal(connectionPathLabel(network, t), "connection_path_tailnet");
  assert.equal(fallbackReasonLabel(network, t), "");
  assert.equal(seatProtocolLabel(network), "");
});
