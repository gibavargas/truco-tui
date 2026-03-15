# Documentation Progress

This file tracks the latest documentation-aligned project status captured in the repository.

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
