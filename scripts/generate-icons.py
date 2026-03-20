#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter


ROOT = Path(__file__).resolve().parents[1]
CANVAS = 512
SCALE = 4


def mix(a: tuple[int, int, int, int], b: tuple[int, int, int, int], t: float) -> tuple[int, int, int, int]:
    return tuple(int(round(a[i] + (b[i] - a[i]) * t)) for i in range(4))


def u(value: float, size: int) -> int:
    return round(value * size / CANVAS)


def build_svg() -> str:
    return """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" fill="none">
  <defs>
    <linearGradient id="bg" x1="48" y1="28" x2="470" y2="488" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#10261d"/>
      <stop offset="0.52" stop-color="#224034"/>
      <stop offset="1" stop-color="#6a3f20"/>
    </linearGradient>
    <radialGradient id="glow" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(372 142) rotate(90) scale(170 190)">
      <stop offset="0" stop-color="#f3c66a" stop-opacity="0.55"/>
      <stop offset="1" stop-color="#f3c66a" stop-opacity="0"/>
    </radialGradient>
    <filter id="shadow" x="-20%" y="-20%" width="160%" height="160%">
      <feDropShadow dx="0" dy="12" stdDeviation="16" flood-color="#09110c" flood-opacity="0.42"/>
    </filter>
  </defs>
  <rect width="512" height="512" rx="112" fill="url(#bg)"/>
  <rect x="18" y="18" width="476" height="476" rx="96" fill="none" stroke="#f1d7a0" stroke-opacity="0.20" stroke-width="2"/>
  <circle cx="378" cy="118" r="86" fill="url(#glow)"/>
  <g filter="url(#shadow)">
    <g transform="translate(292 182) rotate(12 80 116)">
      <rect x="0" y="0" width="160" height="232" rx="26" fill="#f8eed8" stroke="#e7c68f" stroke-width="4"/>
      <rect x="14" y="14" width="132" height="204" rx="18" fill="none" stroke="#ffffff" stroke-opacity="0.40" stroke-width="2"/>
      <g transform="translate(83 113)">
        <ellipse cx="-24" cy="-12" rx="20" ry="20" fill="#1f2a25"/>
        <ellipse cx="24" cy="-12" rx="20" ry="20" fill="#1f2a25"/>
        <ellipse cx="0" cy="-40" rx="20" ry="20" fill="#1f2a25"/>
        <rect x="-11" y="-5" width="22" height="58" rx="10" fill="#1f2a25"/>
      </g>
    </g>
    <g transform="translate(120 118) rotate(-14 88 128)">
      <rect x="0" y="0" width="176" height="256" rx="28" fill="#fff6e5" stroke="#e3c288" stroke-width="4"/>
      <rect x="16" y="16" width="144" height="224" rx="20" fill="none" stroke="#ffffff" stroke-opacity="0.45" stroke-width="2"/>
      <g transform="translate(88 122)">
        <circle cx="-24" cy="-12" r="20" fill="#c53a31"/>
        <circle cx="24" cy="-12" r="20" fill="#c53a31"/>
        <path d="M 0 52 C -20 32 -40 14 -40 -6 C -40 -24 -26 -36 -12 -36 C -2 -36 6 -30 0 -18 C 6 -30 14 -36 24 -36 C 38 -36 52 -24 52 -6 C 52 14 32 32 0 52 Z" fill="#c53a31"/>
      </g>
    </g>
  </g>
  <rect x="78" y="404" width="356" height="24" rx="12" fill="#d7b56d" fill-opacity="0.25"/>
  <circle cx="392" cy="112" r="14" fill="#f6e3af" fill-opacity="0.85"/>
</svg>
"""


def draw_heart(draw: ImageDraw.ImageDraw, x: int, y: int, w: int, h: int, fill: tuple[int, int, int, int]) -> None:
    r = min(w, h) // 4
    left = (x + w // 4 - r, y + h // 4 - r, x + w // 4 + r, y + h // 4 + r)
    right = (x + 3 * w // 4 - r, y + h // 4 - r, x + 3 * w // 4 + r, y + h // 4 + r)
    base = [
        (x + w // 2, y + h),
        (x + w // 10, y + h // 2),
        (x + 9 * w // 10, y + h // 2),
    ]
    draw.ellipse(left, fill=fill)
    draw.ellipse(right, fill=fill)
    draw.polygon(base, fill=fill)


def draw_club(draw: ImageDraw.ImageDraw, x: int, y: int, w: int, h: int, fill: tuple[int, int, int, int]) -> None:
    r = min(w, h) // 4
    top = (x + w // 2 - r, y, x + w // 2 + r, y + 2 * r)
    left = (x, y + h // 4 - r, x + 2 * r, y + h // 4 + r)
    right = (x + w - 2 * r, y + h // 4 - r, x + w, y + h // 4 + r)
    stem = (x + w // 2 - w // 10, y + h // 2 - h // 20, x + w // 2 + w // 10, y + h)
    draw.ellipse(top, fill=fill)
    draw.ellipse(left, fill=fill)
    draw.ellipse(right, fill=fill)
    draw.rounded_rectangle(stem, radius=w // 12, fill=fill)


def create_card(width: int, height: int, *, card_fill: tuple[int, int, int, int], border: tuple[int, int, int, int], symbol: str) -> Image.Image:
    card = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(card)
    draw.rounded_rectangle((0, 0, width - 1, height - 1), radius=min(width, height) // 10, fill=card_fill, outline=border, width=max(2, min(width, height) // 64))
    inner = max(3, min(width, height) // 18)
    draw.rounded_rectangle((inner, inner, width - inner - 1, height - inner - 1), radius=min(width, height) // 14, outline=(255, 255, 255, 110), width=max(1, min(width, height) // 160))
    if symbol == "heart":
        draw_heart(draw, width // 4, height // 4, width // 2, height // 2, (196, 58, 49, 255))
    elif symbol == "club":
        draw_club(draw, width // 4, height // 4, width // 2, height // 2, (26, 34, 30, 255))
    return card


def create_shadow(width: int, height: int) -> Image.Image:
    shadow = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(shadow)
    draw.rounded_rectangle((0, 0, width - 1, height - 1), radius=min(width, height) // 10, fill=(0, 0, 0, 160))
    return shadow.filter(ImageFilter.GaussianBlur(max(6, min(width, height) // 32)))


def render_icon(size: int) -> Image.Image:
    canvas = Image.new("RGBA", (size * SCALE, size * SCALE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(canvas)

    top = (16, 38, 22, 255)
    mid = (34, 64, 52, 255)
    bottom = (108, 64, 32, 255)
    for y in range(canvas.height):
        t = y / max(1, canvas.height - 1)
        if t < 0.62:
            c = mix(top, mid, t / 0.62)
        else:
            c = mix(mid, bottom, (t - 0.62) / 0.38)
        draw.line((0, y, canvas.width, y), fill=c)

    glow = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    glow_draw = ImageDraw.Draw(glow)
    glow_draw.ellipse(
        (u(220, size) * SCALE, u(50, size) * SCALE, u(470, size) * SCALE, u(300, size) * SCALE),
        fill=(243, 198, 106, 120),
    )
    glow_draw.ellipse(
        (u(70, size) * SCALE, u(310, size) * SCALE, u(330, size) * SCALE, u(520, size) * SCALE),
        fill=(75, 190, 132, 54),
    )
    glow = glow.filter(ImageFilter.GaussianBlur(size * 0.18))
    canvas.alpha_composite(glow)

    frame = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    frame_draw = ImageDraw.Draw(frame)
    margin = u(18, size) * SCALE
    frame_draw.rounded_rectangle(
        (margin, margin, canvas.width - margin, canvas.height - margin),
        radius=u(96, size) * SCALE,
        outline=(241, 215, 160, 60),
        width=max(2, size // 96),
    )
    canvas.alpha_composite(frame)

    def paste_card(center_x: float, center_y: float, w: float, h: float, angle: float, symbol: str) -> None:
        card_w = max(1, round(w * size * SCALE))
        card_h = max(1, round(h * size * SCALE))
        card = create_card(
            card_w,
            card_h,
            card_fill=(248, 238, 216, 255),
            border=(231, 198, 143, 255),
            symbol=symbol,
        )
        shadow = create_shadow(card_w, card_h)
        shadow = shadow.rotate(angle, resample=Image.Resampling.BICUBIC, expand=True)
        card = card.rotate(angle, resample=Image.Resampling.BICUBIC, expand=True)

        x = int(center_x * size * SCALE - card.width / 2)
        y = int(center_y * size * SCALE - card.height / 2)
        canvas.alpha_composite(shadow, (x + u(8, size) * SCALE, y + u(12, size) * SCALE))
        canvas.alpha_composite(card, (x, y))

    paste_card(0.34, 0.52, 0.34, 0.50, -14, "heart")
    paste_card(0.63, 0.57, 0.31, 0.45, 12, "club")

    chip = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    chip_draw = ImageDraw.Draw(chip)
    chip_draw.ellipse(
        (u(356, size) * SCALE, u(88, size) * SCALE, u(404, size) * SCALE, u(136, size) * SCALE),
        outline=(246, 227, 175, 190),
        width=max(2, size // 128),
    )
    chip_draw.ellipse(
        (u(374, size) * SCALE, u(106, size) * SCALE, u(386, size) * SCALE, u(118, size) * SCALE),
        fill=(246, 227, 175, 190),
    )
    canvas.alpha_composite(chip)

    return canvas.resize((size, size), resample=Image.Resampling.LANCZOS)


def write_png(path: Path, size: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    render_icon(size).save(path)


def write_ico(path: Path, source_size: int = 1024) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    render_icon(source_size).save(path, format="ICO", sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)])


def write_svg(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(build_svg(), encoding="utf-8")


def main() -> None:
    svg_targets = [
        ROOT / "native/linux-gtk/dev.truco.Native.svg",
        ROOT / "browser-edition/php/favicon.svg",
        ROOT / "browser-edition/dist/favicon.svg",
    ]
    for target in svg_targets:
        write_svg(target)

    for base in [ROOT / "browser-edition/php", ROOT / "browser-edition/dist"]:
        write_png(base / "apple-touch-icon.png", 180)
        render_icon(512).save(base / "favicon.png")
        write_ico(base / "favicon.ico")

    write_ico(ROOT / "native/windows-winui/Assets/truco.ico")

    mac_icon_dir = ROOT / "native/macos/Truco/Truco/Assets.xcassets/AppIcon.appiconset"
    mac_sizes = {
        "icon_16x16.png": 16,
        "icon_16x16@2x.png": 32,
        "icon_32x32.png": 32,
        "icon_32x32@2x.png": 64,
        "icon_128x128.png": 128,
        "icon_128x128@2x.png": 256,
        "icon_256x256.png": 256,
        "icon_256x256@2x.png": 512,
        "icon_512x512.png": 512,
        "icon_512x512@2x.png": 1024,
    }
    for filename, size in mac_sizes.items():
        write_png(mac_icon_dir / filename, size)


if __name__ == "__main__":
    main()
