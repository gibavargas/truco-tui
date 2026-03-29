<?php
/**
 * Return a localized card label for assistive text and tooltips.
 */
function cardLabel(array $card): string
{
    $rank = (string) ($card['Rank'] ?? '?');
    $suit = (string) ($card['Suit'] ?? '');

    return tr('card_of', $rank, tr('suit_' . $suit));
}

/**
 * Render a face-up card as HTML.
 *
 * @param array  $card     Card with 'Rank' and 'Suit'
 * @param bool   $small    Whether to use the small card style
 * @param string $keyHint  Optional hotkey hint (not used in PHP edition)
 */
function renderCard(array $card, bool $small = false, string $keyHint = ''): string
{
    $rank = htmlspecialchars($card['Rank'] ?? '?');
    $suit = $card['Suit'] ?? '';
    $symbol = suitSymbol($suit);
    $colorClass = suitColorClass($suit);
    $sizeClass = $small ? 'small' : '';

    $hint = '';
    if ($keyHint !== '') {
        $hint = '<div class="hotkey">' . htmlspecialchars($keyHint) . '</div>';
    }

    return <<<HTML
<div class="card {$sizeClass} {$colorClass}">
  <div class="corner"><span>{$rank}</span><span>{$symbol}</span></div>
  <div class="suit">{$symbol}</div>
  <div class="corner bottom"><span>{$rank}</span><span>{$symbol}</span></div>
  {$hint}
</div>
HTML;
}

/**
 * Render a face-down card (back).
 */
function renderCardBack(bool $small = true): string
{
    $sizeClass = $small ? 'small' : '';
    return '<div class="card ' . $sizeClass . ' card-back"></div>';
}
