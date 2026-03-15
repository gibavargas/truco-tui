Original prompt: Improve the browser edition into a world-class digital Truco Paulista UI centered on table drama, card readability, strong action clarity, compact narration, and a premium 4-player table layout.

# Documentation Progress

This file tracks the latest documentation-aligned project status captured in the repository.

## Current Browser Edition Redesign

- Reworked the match view into a table-first layout with a compact score strip, dramatic center callout, compact narration, a stronger action dock, and a premium hand tray.
- Reframed side information into lighter "Boca da mesa" notes instead of tall empty dashboard panels.
- Added clearer ally/enemy encoding, trick progress, opening-seat emphasis, pending truco pressure states, and stronger action messaging for Truco / Aceitar / Correr / Aumentar.
- Updated the PHP i18n strings to support the new game presentation in both Portuguese and English.

## Latest Validation

- Ran the official Playwright-based validation client against the local browser edition after the redesign.
- Captured an additional full-page 4-player desktop screenshot with Playwright to verify the intended desktop composition.
- Observed no console or page errors during the latest validation run.

## Current Direction Shift

- Reframed the browser match screen away from stacked dashboard panels and toward a single-table composition with a compact HUD, simplified center table, and a dedicated player dock.
- Removed the in-match refresh action from the primary controls because it does not fit the game flow.
- Prioritized immediate playability: the player hand is now the main dock, context actions only appear when relevant, and secondary narration was reduced to a compact footer line.

## Recent Browser Edition Validation

- Visual validation completed for desktop and mobile layouts.
- Setup, lobby, and match flows were checked with real screenshots.
- Alignment, spacing, contrast, responsiveness, overflow, and consistency issues were adjusted.

## Recent Browser Edition Fixes

- Fixed a real AJAX bug where `form.action` was being shadowed by `input[name="action"]` in `browser-edition/php/ajax.js`.
- Fixed PHP 8.5 `curl_close()` warnings that were leaking into AJAX responses.

## Verified Browser Scenarios

- Setup screen on desktop
- Setup screen on mobile
- Online lobby on desktop
- Online lobby on mobile
- Offline match on desktop
- Offline match on mobile
- Online match on desktop

## Notable UI Adjustments

- The lobby invite block now wraps cleanly without colliding with action buttons.
- The mobile hand view now scrolls horizontally instead of wrapping cards onto two rows.

## 2026-03-15 Current Codex Work

- Implementing a second browser-edition pass focused on competitive desktop readability for 4-player Truco Paulista.
- Scope is limited to the PHP browser UI layer: semantic match-state classes, wireframe/polished toggle, stronger table zoning, and a quieter treatment for social/online controls.
- Planned validation loop: local Go API + PHP frontend + Playwright screenshot/state checks after the UI changes land.
- Completed a real browser validation pass with Playwright against the local PHP frontend after the redesign and after a spacing fix on the top seat/callout area.
- Latest desktop capture confirms the polished layout and the wireframe toggle both render without console errors.
